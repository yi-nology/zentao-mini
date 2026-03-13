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
	taskChan  chan Task
	resultChan chan TaskResult
	workers   int
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
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

// SubmitWithResult 提交任务并等待结果
func (p *WorkerPool) SubmitWithResult(task Task) (interface{}, error) {
	if !p.Submit(task) {
		return nil, p.ctx.Err()
	}

	select {
	case <-p.ctx.Done():
		return nil, p.ctx.Err()
	case result := <-p.resultChan:
		return result.Value, result.Error
	}
}

// Results 获取结果通道
func (p *WorkerPool) Results() <-chan TaskResult {
	return p.resultChan
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
			break
		case result := <-p.resultChan:
			results = append(results, result)
		}
	}

	return results
}

// ProcessBatchWithCallback 批量处理任务并使用回调处理结果
// tasks: 任务列表
// callback: 结果回调函数
func (p *WorkerPool) ProcessBatchWithCallback(tasks []Task, callback func(result TaskResult)) {
	// 提交所有任务
	for _, task := range tasks {
		if !p.Submit(task) {
			log.Printf("[WorkerPool] Failed to submit task: context cancelled")
			break
		}
	}

	// 收集并处理结果
	for i := 0; i < len(tasks); i++ {
		select {
		case <-p.ctx.Done():
			callback(TaskResult{Error: p.ctx.Err()})
			return
		case result := <-p.resultChan:
			callback(result)
		}
	}
}

// ParallelExecute 并行执行多个任务并收集结果
// 这是一个便捷方法，适用于简单的并行场景
func ParallelExecute(tasks []Task, workers int) []TaskResult {
	if len(tasks) == 0 {
		return []TaskResult{}
	}

	// 如果任务数少于工作协程数，调整工作协程数
	if len(tasks) < workers {
		workers = len(tasks)
	}

	pool := NewWorkerPool(workers, len(tasks))
	defer pool.Shutdown()

	return pool.ProcessBatch(tasks)
}

// ParallelExecuteWithCallback 并行执行任务并通过回调处理结果
func ParallelExecuteWithCallback(tasks []Task, workers int, callback func(result TaskResult)) {
	if len(tasks) == 0 {
		return
	}

	// 如果任务数少于工作协程数，调整工作协程数
	if len(tasks) < workers {
		workers = len(tasks)
	}

	pool := NewWorkerPool(workers, len(tasks))
	defer pool.Shutdown()

	pool.ProcessBatchWithCallback(tasks, callback)
}
