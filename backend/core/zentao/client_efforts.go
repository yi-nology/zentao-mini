package zentao

import (
	"sync"
	"time"

	"github.com/yi-nology/common/biz/zentao"
	"go.uber.org/zap"

	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/metrics"
)

// DailyEffortSummary 产品指定日期区间内按人按日的工时汇总
type DailyEffortSummary struct {
	// DailyByAccount 账号 -> 日期(YYYY-MM-DD) -> 当日填报工时
	DailyByAccount map[string]map[string]float64
	// TotalByAccount 账号 -> 区间内总工时
	TotalByAccount map[string]float64
	// ActiveAssignees 区间内"在册"人员账号集合 = 有工时记录的人 ∪ 活跃任务（wait/doing/pause）的指派人
	ActiveAssignees map[string]bool
	// TaskCount 参与统计的任务数
	TaskCount int
	// EffortCount 落在区间内的工时条数
	EffortCount int
}

// activeTaskStatuses 任务处于这些状态时视为仍在产品上投入，其指派人纳入日报检查名册
var activeTaskStatuses = map[string]bool{"wait": true, "doing": true, "pause": true}

// taskAssigneeAccount 从任务 assignedTo 提取账号，兼容对象（{account:...}）与字符串两种返回格式
func taskAssigneeAccount(v interface{}) string {
	switch ref := v.(type) {
	case zentao.UserRef:
		return ref.Account
	case map[string]interface{}:
		if a, ok := ref["account"].(string); ok {
			return a
		}
	case string:
		return ref
	}
	return ""
}

// GetProductDailyEfforts 遍历产品的 项目→执行→任务→工时 链路，按人按日汇总 [dateFrom, dateTo] 区间内的工时。
// 日期格式 YYYY-MM-DD。用于月度日报完成度检查；执行/任务读取走既有缓存，工时明细按任务并发拉取（不缓存，补填后立即可见）。
func (c *Client) GetProductDailyEfforts(productID int, dateFrom, dateTo string) (*DailyEffortSummary, error) {
	start := time.Now()

	execs, err := c.GetExecutionsByProduct(productID)
	if err != nil {
		return nil, err
	}

	seenExec := make(map[int]bool)
	var allTasks []zentao.Task
	for _, ec := range execs {
		if seenExec[ec.Exec.ID] {
			continue
		}
		seenExec[ec.Exec.ID] = true
		tasks, err := c.GetTasks(ec.Exec.ID, 1, 10000)
		if err != nil {
			continue
		}
		allTasks = append(allTasks, tasks...)
	}

	summary := &DailyEffortSummary{
		DailyByAccount:  make(map[string]map[string]float64),
		TotalByAccount:  make(map[string]float64),
		ActiveAssignees: make(map[string]bool),
		TaskCount:       len(allTasks),
	}

	var mu sync.Mutex
	pool := NewWorkerPool(5, len(allTasks))
	defer pool.Shutdown()

	type effortBatch struct {
		assigneeAccount string
		assigneeActive  bool
		efforts         []zentao.EffortEntry
	}

	effortTasks := make([]Task, len(allTasks))
	for i, t := range allTasks {
		t := t
		effortTasks[i] = func() (interface{}, error) {
			efforts, err := c.GetTaskEfforts(t.ID)
			if err != nil {
				return nil, err
			}
			acc := taskAssigneeAccount(t.AssignedTo)
			return effortBatch{
				assigneeAccount: acc,
				assigneeActive:  acc != "" && activeTaskStatuses[t.Status],
				efforts:         efforts,
			}, nil
		}
	}

	for _, result := range pool.ProcessBatch(effortTasks) {
		if result.Error != nil || result.Value == nil {
			continue
		}
		batch := result.Value.(effortBatch)
		mu.Lock()
		if batch.assigneeActive {
			summary.ActiveAssignees[batch.assigneeAccount] = true
		}
		for _, e := range batch.efforts {
			if e.Deleted == "1" {
				continue
			}
			if dateFrom != "" && e.Date < dateFrom {
				continue
			}
			if dateTo != "" && e.Date > dateTo {
				continue
			}
			if e.Account == "" {
				continue
			}
			if summary.DailyByAccount[e.Account] == nil {
				summary.DailyByAccount[e.Account] = make(map[string]float64)
			}
			summary.DailyByAccount[e.Account][e.Date] += e.Consumed
			summary.TotalByAccount[e.Account] += e.Consumed
			summary.ActiveAssignees[e.Account] = true
			summary.EffortCount++
		}
		mu.Unlock()
	}

	logger.Info("产品工时按人按日汇总完成",
		zap.Int("productID", productID),
		zap.String("dateFrom", dateFrom),
		zap.String("dateTo", dateTo),
		zap.Int("taskCount", summary.TaskCount),
		zap.Int("effortCount", summary.EffortCount),
		zap.Int("people", len(summary.ActiveAssignees)),
		zap.Duration("duration", time.Since(start)))
	metrics.RecordZentaoAPIRequest("efforts-daily", "GET", time.Since(start), nil)

	return summary, nil
}
