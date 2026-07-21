package zentao

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/yi-nology/common/biz/zentao"
)

// effortItem 工时记录项
type effortItem struct {
	ID        int     `json:"id"`
	TaskID    int     `json:"taskId"`
	TaskName  string  `json:"taskName"`
	TaskType  string  `json:"taskType"`
	Product   string  `json:"product"`
	Project   string  `json:"project"`
	Execution string  `json:"execution"`
	Account   string  `json:"account"`
	Date      string  `json:"date"`
	Consumed  float64 `json:"consumed"`
	Left      float64 `json:"left"`
	Work      string  `json:"work"`
}

// statItem 统计项
type statItem struct {
	Name  string  `json:"name"`
	Hours float64 `json:"hours"`
	Count int     `json:"count"`
}

// dailyStat 每日统计
type dailyStat struct {
	Date  string  `json:"date"`
	Hours float64 `json:"hours"`
	Count int     `json:"count"`
}

func mapToSlice(m map[string]*statItem) []statItem {
	result := make([]statItem, 0, len(m))
	for _, v := range m {
		result = append(result, *v)
	}
	return result
}

// GetTaskEfforts 获取任务的工时记录
func (c *Client) GetTaskEfforts(taskID int) ([]zentao.EffortEntry, error) {
	if c.IsSessionMode() {
		return c.getTaskEffortsSession(context.Background(), taskID)
	}
	var result []zentao.EffortEntry
	err := c.withTokenRetry("GetTaskEfforts", func(client *zentao.Client) error {
		var err error
		result, err = client.GetTaskEfforts(taskID)
		return err
	})
	return result, err
}

// GetTaskEffortsContext 获取任务的工时记录（支持 context 取消）
func (c *Client) GetTaskEffortsContext(ctx context.Context, taskID int) ([]zentao.EffortEntry, error) {
	if c.IsSessionMode() {
		return c.getTaskEffortsSession(ctx, taskID)
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var result []zentao.EffortEntry
	err := c.withTokenRetryContext(ctx, "GetTaskEfforts", func(client *zentao.Client) error {
		var err error
		result, err = client.GetTaskEfforts(taskID)
		return err
	})
	return result, err
}

// GetTimelogAnalysis 获取工时统计分析
//
// 该方法原本直接调用 c.sdkClient.*（绕过 wrapper），导致会话模式无法工作。
// 重构为通过 wrapper 方法（GetProduct/GetProjectsByProduct/GetExecutions/
// GetTasks/GetTaskEfforts）获取数据 —— 这些方法已经做了 token/session 双模式分发，
// 因此本方法自动支持两种模式，无需额外的 session 分支。
func (c *Client) GetTimelogAnalysis(productID, projectID, executionID, assignedTo, dateFrom, dateTo string) (map[string]interface{}, error) {
	ctx := context.Background()

	prodID, err := strconv.Atoi(productID)
	if err != nil {
		return nil, fmt.Errorf("无效的productId")
	}

	var filterProjectID, filterExecutionID int
	if projectID != "" {
		filterProjectID, _ = strconv.Atoi(projectID)
	}
	if executionID != "" {
		filterExecutionID, _ = strconv.Atoi(executionID)
	}

	productName := fmt.Sprintf("产品%d", prodID)
	if product, perr := c.GetProductContext(ctx, prodID); perr == nil && product != nil {
		productName = product.Name
	}

	projects, err := c.GetProjectsByProductContext(ctx, prodID, 1, 100)
	if err != nil {
		return nil, fmt.Errorf("获取项目列表失败: %w", err)
	}

	if filterProjectID > 0 {
		filtered := make([]zentao.Project, 0)
		for _, p := range projects {
			if p.ID == filterProjectID {
				filtered = append(filtered, p)
				break
			}
		}
		projects = filtered
	}

	type execWithContext struct {
		Exec     zentao.Execution
		ProjName string
	}

	var allExecs []execWithContext
	var mu sync.Mutex

	if filterExecutionID > 0 {
		projName := ""
		for _, p := range projects {
			projName = p.Name
			break
		}
		allExecs = append(allExecs, execWithContext{
			Exec:     zentao.Execution{ID: filterExecutionID},
			ProjName: projName,
		})
	} else {
		execPool := NewWorkerPool(3, len(projects))
		defer execPool.Shutdown()

		execTasks := make([]Task, len(projects))
		for i, proj := range projects {
			proj := proj
			execTasks[i] = func() (interface{}, error) {
				execs, eerr := c.GetExecutionsContext(ctx, proj.ID, 1, 100)
				if eerr != nil {
					return nil, eerr
				}
				return execs, nil
			}
		}

		for i, result := range execPool.ProcessBatch(execTasks) {
			if result.Error == nil && result.Value != nil {
				execs := result.Value.([]zentao.Execution)
				mu.Lock()
				for _, e := range execs {
					allExecs = append(allExecs, execWithContext{Exec: e, ProjName: projects[i].Name})
				}
				mu.Unlock()
			}
		}
	}

	type taskContext struct {
		Task     zentao.Task
		ProjName string
		ExecName string
	}
	var allTaskCtx []taskContext

	taskPool := NewWorkerPool(3, len(allExecs))
	defer taskPool.Shutdown()

	taskTasks := make([]Task, len(allExecs))
	for i, ec := range allExecs {
		ec := ec
		taskTasks[i] = func() (interface{}, error) {
			tasks, terr := c.GetTasksContext(ctx, ec.Exec.ID, 1, 500)
			if terr != nil {
				return nil, terr
			}
			var filteredTasks []zentao.Task
			for _, t := range tasks {
				if toFloat64(t.Consumed) > 0 {
					filteredTasks = append(filteredTasks, t)
				}
			}
			return struct {
				Tasks    []zentao.Task
				ProjName string
				ExecName string
			}{Tasks: filteredTasks, ProjName: ec.ProjName, ExecName: ec.Exec.Name}, nil
		}
	}

	for _, result := range taskPool.ProcessBatch(taskTasks) {
		if result.Error == nil && result.Value != nil {
			data := result.Value.(struct {
				Tasks    []zentao.Task
				ProjName string
				ExecName string
			})
			mu.Lock()
			for _, t := range data.Tasks {
				allTaskCtx = append(allTaskCtx, taskContext{
					Task:     t,
					ProjName: data.ProjName,
					ExecName: data.ExecName,
				})
			}
			mu.Unlock()
		}
	}

	var allEfforts []effortItem

	effortPool := NewWorkerPool(5, len(allTaskCtx))
	defer effortPool.Shutdown()

	effortTasks := make([]Task, len(allTaskCtx))
	for i, tc := range allTaskCtx {
		tc := tc
		effortTasks[i] = func() (interface{}, error) {
			efforts, ferr := c.GetTaskEffortsContext(ctx, tc.Task.ID)
			if ferr != nil {
				return nil, ferr
			}

			var filteredEfforts []effortItem
			for _, e := range efforts {
				if assignedTo != "" && e.Account != assignedTo {
					continue
				}
				if dateFrom != "" && e.Date < dateFrom {
					continue
				}
				if dateTo != "" && e.Date > dateTo {
					continue
				}

				filteredEfforts = append(filteredEfforts, effortItem{
					ID:        e.ID,
					TaskID:    tc.Task.ID,
					TaskName:  tc.Task.Name,
					TaskType:  tc.Task.Type,
					Product:   productName,
					Project:   tc.ProjName,
					Execution: tc.ExecName,
					Account:   e.Account,
					Date:      e.Date,
					Consumed:  e.Consumed,
					Left:      e.Left,
					Work:      e.Work,
				})
			}
			return filteredEfforts, nil
		}
	}

	for _, result := range effortPool.ProcessBatch(effortTasks) {
		if result.Error == nil && result.Value != nil {
			efforts := result.Value.([]effortItem)
			mu.Lock()
			allEfforts = append(allEfforts, efforts...)
			mu.Unlock()
		}
	}

	var totalHours float64
	byProjectMap := make(map[string]*statItem)
	byTypeMap := make(map[string]*statItem)
	byDateMap := make(map[string]*dailyStat)

	typeNames := map[string]string{
		"devel": "开发", "design": "设计", "test": "测试",
		"study": "研究", "discuss": "讨论", "ui": "界面",
		"affair": "事务", "misc": "其他",
	}

	for _, e := range allEfforts {
		totalHours += e.Consumed

		if _, ok := byProjectMap[e.Project]; !ok {
			byProjectMap[e.Project] = &statItem{Name: e.Project}
		}
		byProjectMap[e.Project].Hours += e.Consumed
		byProjectMap[e.Project].Count++

		typeName := e.TaskType
		if tn, ok := typeNames[e.TaskType]; ok {
			typeName = tn
		}
		if _, ok := byTypeMap[typeName]; !ok {
			byTypeMap[typeName] = &statItem{Name: typeName}
		}
		byTypeMap[typeName].Hours += e.Consumed
		byTypeMap[typeName].Count++

		if _, ok := byDateMap[e.Date]; !ok {
			byDateMap[e.Date] = &dailyStat{Date: e.Date}
		}
		byDateMap[e.Date].Hours += e.Consumed
		byDateMap[e.Date].Count++
	}

	byProject := mapToSlice(byProjectMap)
	byType := mapToSlice(byTypeMap)

	byDate := make([]dailyStat, 0, len(byDateMap))
	for _, v := range byDateMap {
		byDate = append(byDate, *v)
	}

	effortsData := make([]map[string]interface{}, len(allEfforts))
	for i, e := range allEfforts {
		effortsData[i] = map[string]interface{}{
			"id":        e.ID,
			"taskId":    e.TaskID,
			"taskName":  e.TaskName,
			"taskType":  e.TaskType,
			"product":   e.Product,
			"project":   e.Project,
			"execution": e.Execution,
			"account":   e.Account,
			"date":      e.Date,
			"consumed":  e.Consumed,
			"left":      e.Left,
			"work":      e.Work,
		}
	}

	return map[string]interface{}{
		"totalHours":  totalHours,
		"effortCount": len(allEfforts),
		"taskCount":   len(allTaskCtx),
		"byProject":   byProject,
		"byType":      byType,
		"byDate":      byDate,
		"efforts":     effortsData,
	}, nil
}

// toFloat64 把 SDK 里 interface{} 类型的 Consumed/Left 安全转 float64。
// session 模式映射时这些字段已是 float64；token 模式下 SDK 可能返回 float64 或 json.Number。
func toFloat64(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(n, 64)
		return f
	}
	return 0
}
