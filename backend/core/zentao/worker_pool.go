package zentao

import (
	"context"
	"log"
	"sync"
)

// Task 表示一个工作任务
type Task func() (interface{}, error)

// TaskResult 表示任务执行结果
type TaskResult struct {
	Value interface{}
	Error error
}

// WorkerPool 工作池，用于并发控制
type WorkerPool struct {
	taskChan   chan Task
	resultChan chan TaskResult
	workers    int
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewWorkerPool 创建新的工作池
// workers: 工作协程数量
// bufferSize: 任务缓冲区大小
func NewWorkerPool(workers int, bufferSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		taskChan:   make(chan Task, bufferSize),
		resultChan: make(chan TaskResult, bufferSize),
		workers:    workers,
		ctx:        ctx,
		cancel:     cancel,
	}

	// 启动工作协程
	pool.start()

	return pool
}

// start 启动工作协程
func (p *WorkerPool) start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// worker 工作协程
func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			// 收到停止信号，退出
			return
		case task, ok := <-p.taskChan:
			if !ok {
				// 任务通道已关闭
				return
			}

			// 执行任务
			result := TaskResult{}
			if task != nil {
				value, err := task()
				result.Value = value
				result.Error = err
			}

			// 发送结果
			select {
			case <-p.ctx.Done():
				return
			case p.resultChan <- result:
			}
		}
	}
}

// Submit 提交任务到工作池
func (p *WorkerPool) Submit(task Task) bool {
	select {
	case <-p.ctx.Done():
		return false
	case p.taskChan <- task:
		return true
	}
}

// Shutdown 优雅关闭工作池
func (p *WorkerPool) Shutdown() {
	p.cancel()
	close(p.taskChan)
	p.wg.Wait()
	close(p.resultChan)
}

// ProcessBatch 批量处理任务
// tasks: 任务列表
// 返回所有任务的结果
func (p *WorkerPool) ProcessBatch(tasks []Task) []TaskResult {
	results := make([]TaskResult, 0, len(tasks))

	// 提交所有任务
	for _, task := range tasks {
		if !p.Submit(task) {
			log.Printf("[WorkerPool] Failed to submit task: context cancelled")
			break
		}
	}

	// 收集所有结果
	for i := 0; i < len(tasks); i++ {
		select {
		case <-p.ctx.Done():
			results = append(results, TaskResult{Error: p.ctx.Err()})
			return results
		case result := <-p.resultChan:
			results = append(results, result)
		}
	}

	return results
}
