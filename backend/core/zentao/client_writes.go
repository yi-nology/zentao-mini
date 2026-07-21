package zentao

import (
	"context"
	"fmt"

	"github.com/yi-nology/common/biz/zentao"
	neturl "net/url"
)

// 本文件为 wrapper 补齐写操作（Phase2c）。
// 之前 wrapper 是纯只读的；这里加入 bug/task/story 的常用写操作：
//   - bug: Resolve / Close / Assign / Confirm / Activate
//   - task: Start / Finish / Pause / Assign
//   - story: Change
//
// 每个方法做 token/session 双模式分发：
//   - token 模式：调 SDK 对应写方法（经 withTokenRetry）
//   - session 模式：POST 禅道 PHP 表单端点（经 doSessionPost）
//
// PHP 写端点格式：/{module}-{action}-{id}.json，表单字段与 SDK request 对齐。

// ---- Bug writes ----

func (c *Client) ResolveBug(bugID int, req zentao.BugResolveRequest) error {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("resolution", req.Resolution)
		form.Set("resolvedBuild", req.ResolvedBuild)
		if req.Comment != "" {
			form.Set("comment", req.Comment)
		}
		return c.doSessionPost(context.Background(), "ResolveBug",
			fmt.Sprintf("/bug-resolve-%d.json", bugID), form)
	}
	return c.withTokenRetry("ResolveBug", func(client *zentao.Client) error {
		return client.ResolveBug(bugID, req)
	})
}

func (c *Client) CloseBug(bugID int, req zentao.BugCloseRequest) error {
	if c.IsSessionMode() {
		form := neturl.Values{}
		if req.Comment != "" {
			form.Set("comment", req.Comment)
		}
		return c.doSessionPost(context.Background(), "CloseBug",
			fmt.Sprintf("/bug-close-%d.json", bugID), form)
	}
	return c.withTokenRetry("CloseBug", func(client *zentao.Client) error {
		return client.CloseBug(bugID, req)
	})
}

func (c *Client) AssignBug(bugID int, req zentao.BugAssignRequest) error {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("assignedTo", req.AssignedTo)
		if req.Comment != "" {
			form.Set("comment", req.Comment)
		}
		return c.doSessionPost(context.Background(), "AssignBug",
			fmt.Sprintf("/bug-assignTo-%d.json", bugID), form)
	}
	return c.withTokenRetry("AssignBug", func(client *zentao.Client) error {
		return client.AssignBug(bugID, req)
	})
}

func (c *Client) ConfirmBug(bugID int, req zentao.BugConfirmRequest) error {
	if c.IsSessionMode() {
		form := neturl.Values{}
		if req.Comment != "" {
			form.Set("comment", req.Comment)
		}
		return c.doSessionPost(context.Background(), "ConfirmBug",
			fmt.Sprintf("/bug-confirm-%d.json", bugID), form)
	}
	return c.withTokenRetry("ConfirmBug", func(client *zentao.Client) error {
		return client.ConfirmBug(bugID, req)
	})
}

func (c *Client) ActivateBug(bugID int, req zentao.BugActivateRequest) error {
	if c.IsSessionMode() {
		form := neturl.Values{}
		if req.AssignedTo != "" {
			form.Set("assignedTo", req.AssignedTo)
		}
		if req.Comment != "" {
			form.Set("comment", req.Comment)
		}
		return c.doSessionPost(context.Background(), "ActivateBug",
			fmt.Sprintf("/bug-activate-%d.json", bugID), form)
	}
	return c.withTokenRetry("ActivateBug", func(client *zentao.Client) error {
		return client.ActivateBug(bugID, req)
	})
}

// ---- Task writes ----

func (c *Client) StartTask(taskID int, req zentao.TaskStartRequest) error {
	if c.IsSessionMode() {
		form := neturl.Values{}
		if req.RealStarted != "" {
			form.Set("realStarted", req.RealStarted)
		}
		form.Set("consumed", fmt.Sprintf("%g", req.Consumed))
		form.Set("left", fmt.Sprintf("%g", req.Left))
		if req.Comment != "" {
			form.Set("comment", req.Comment)
		}
		return c.doSessionPost(context.Background(), "StartTask",
			fmt.Sprintf("/task-start-%d.json", taskID), form)
	}
	return c.withTokenRetry("StartTask", func(client *zentao.Client) error {
		return client.StartTask(taskID, req)
	})
}

func (c *Client) FinishTask(taskID int, req zentao.TaskFinishRequest) error {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("consumed", fmt.Sprintf("%g", req.Consumed))
		if req.FinishedDate != "" {
			form.Set("finishedDate", req.FinishedDate)
		}
		if req.Comment != "" {
			form.Set("comment", req.Comment)
		}
		return c.doSessionPost(context.Background(), "FinishTask",
			fmt.Sprintf("/task-finish-%d.json", taskID), form)
	}
	return c.withTokenRetry("FinishTask", func(client *zentao.Client) error {
		return client.FinishTask(taskID, req)
	})
}

func (c *Client) PauseTask(taskID int, req zentao.TaskPauseRequest) error {
	if c.IsSessionMode() {
		form := neturl.Values{}
		if req.Comment != "" {
			form.Set("comment", req.Comment)
		}
		return c.doSessionPost(context.Background(), "PauseTask",
			fmt.Sprintf("/task-pause-%d.json", taskID), form)
	}
	return c.withTokenRetry("PauseTask", func(client *zentao.Client) error {
		return client.PauseTask(taskID, req)
	})
}

func (c *Client) AssignTask(taskID int, req zentao.TaskAssignRequest) error {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("assignedTo", req.AssignedTo)
		form.Set("left", fmt.Sprintf("%g", req.Left))
		if req.Comment != "" {
			form.Set("comment", req.Comment)
		}
		return c.doSessionPost(context.Background(), "AssignTask",
			fmt.Sprintf("/task-assignedTo-%d.json", taskID), form)
	}
	return c.withTokenRetry("AssignTask", func(client *zentao.Client) error {
		return client.AssignTask(taskID, req)
	})
}

// ---- Story writes ----

func (c *Client) ChangeStory(storyID int, spec, verify string) error {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("spec", spec)
		form.Set("verify", verify)
		return c.doSessionPost(context.Background(), "ChangeStory",
			fmt.Sprintf("/story-change-%d.json", storyID), form)
	}
	return c.withTokenRetry("ChangeStory", func(client *zentao.Client) error {
		return client.ChangeStory(storyID, spec, verify)
	})
}
