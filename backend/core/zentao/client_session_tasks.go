package zentao

import (
	"context"
	"fmt"

	"github.com/yi-nology/common/biz/zentao"
)

// ---- Tasks (session mode) ----
//
// 端点：/execution-task-{executionID}.json
// 内层 tasks 是 map[id]task，按 execution 取。单页即全部（麒麟实例 exec 4 有 97 条）。

type sessionExecutionTaskResp struct {
	Title       string                `json:"title"`
	ExecutionID int                   `json:"executionID"`
	Tasks       map[string]sessionTask `json:"tasks"`
	HasTasks    bool                  `json:"hasTasks"`
	Pager       *sessionPager         `json:"pager"`
}

type sessionTask struct {
	ID           int     `json:"id"`
	Project      int     `json:"project"`
	Parent       int     `json:"parent"`
	Execution    int     `json:"execution"`
	Module       int     `json:"module"`
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Pri          int     `json:"pri"`
	Estimate     float64 `json:"estimate"`
	Consumed     float64 `json:"consumed"`
	Left         float64 `json:"left"`
	Deadline     string  `json:"deadline"`
	Status       string  `json:"status"`
	OpenedBy     string  `json:"openedBy"`
	OpenedDate   string  `json:"openedDate"`
	AssignedTo   string  `json:"assignedTo"`
	AssignedDate string  `json:"assignedDate"`
	EstStarted   string  `json:"estStarted"`
	RealStarted  string  `json:"realStarted"`
	FinishedBy   string  `json:"finishedBy"`
	FinishedDate string  `json:"finishedDate"`
	ClosedBy     string  `json:"closedBy"`
	ClosedDate   string  `json:"closedDate"`
	Story        int     `json:"story"`
	Deleted      string  `json:"deleted"`
}

func (t sessionTask) toSDK() zentao.Task {
	return zentao.Task{
		ID:           t.ID,
		Project:      t.Project,
		Execution:    t.Execution,
		Name:         t.Name,
		Type:         t.Type,
		Pri:          t.Pri,
		Status:       t.Status,
		AssignedTo:   t.AssignedTo,
		EstStarted:   t.EstStarted,
		Deadline:     t.Deadline,
		Estimate:     t.Estimate,
		Consumed:     t.Consumed,
		Left:         t.Left,
		OpenedBy:     t.OpenedBy,
		OpenedDate:   t.OpenedDate,
		FinishedBy:   t.FinishedBy,
		FinishedDate: t.FinishedDate,
		ClosedBy:     t.ClosedBy,
		ClosedDate:   t.ClosedDate,
		StatusName:   t.Status,
	}
}

func (c *Client) getTasksSession(ctx context.Context, executionID, page, pageSize int) ([]zentao.Task, error) {
	path := fmt.Sprintf("/execution-task-%d.json", executionID)
	var resp sessionExecutionTaskResp
	if err := c.doSessionJSON(ctx, "GetTasks", path, &resp); err != nil {
		return nil, err
	}
	all := make([]zentao.Task, 0, len(resp.Tasks))
	for _, t := range resp.Tasks {
		all = append(all, t.toSDK())
	}
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	end := start + pageSize
	if end > len(all) {
		end = len(all)
	}
	if start >= len(all) {
		return nil, nil
	}
	return all[start:end], nil
}

func (c *Client) getAllTasksByProjectSession(ctx context.Context, projectID int) ([]zentao.Task, error) {
	execs, err := c.getExecutionsSession(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var out []zentao.Task
	for _, e := range execs {
		ts, terr := c.getTasksSession(ctx, e.ID, 1, 100000)
		if terr != nil {
			continue
		}
		out = append(out, ts...)
	}
	return out, nil
}

func (c *Client) getAllTasksByProductSession(ctx context.Context, productID int) ([]zentao.Task, error) {
	projects, err := c.getProjectsByProductSession(ctx, productID)
	if err != nil {
		return nil, err
	}
	var out []zentao.Task
	for _, p := range projects {
		ts, terr := c.getAllTasksByProjectSession(ctx, p.ID)
		if terr != nil {
			continue
		}
		out = append(out, ts...)
	}
	return out, nil
}
