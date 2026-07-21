package zentao

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/yi-nology/common/biz/zentao"
	neturl "net/url"
)

// 本文件补齐 SDK 的剩余写操作（Phase2c 扩展）。
// 与 client_writes.go 一样的双模式分发：token 模式调 SDK，session 模式 POST PHP 表单端点。
//
// PHP 写端点格式约定（标准禅道）：
//   - create: /{module}-create.json（body 含 product/project/execution 等父ID）
//   - update: /{module}-edit-{id}.json
//   - delete: /{module}-delete-{id}.json
//   - 特殊动作: /{module}-{action}-{id}.json
//
// 返回值：create/update 在 session 模式下无法可靠返回新建对象（PHP 端点返回的是重定向
// 或简单 result），统一返回 nil 指针 + 错误状态；调用方如需对象应再查一次。
// 为保持签名与 SDK 一致，create/update 返回 (*Type, error)，session 模式成功时返回 &零值{}。

// ---- helper ----

// joinInts 把 []int 转成逗号分隔字符串（plan link/unlink 用）。
func joinInts(ids []int) string {
	s := make([]string, len(ids))
	for i, id := range ids {
		s[i] = strconv.Itoa(id)
	}
	return strings.Join(s, ",")
}

// ---- Bug create/update/delete ----

func (c *Client) CreateBug(productID int, req zentao.BugCreateRequest) (*zentao.Bug, error) {
	if c.IsSessionMode() {
		form := bugCreateForm(productID, req)
		if err := c.doSessionPost(context.Background(), "CreateBug", "/bug-create.json", form); err != nil {
			return nil, err
		}
		return &zentao.Bug{}, nil
	}
	var result *zentao.Bug
	err := c.withTokenRetry("CreateBug", func(client *zentao.Client) error {
		var e error
		result, e = client.CreateBug(productID, req)
		return e
	})
	return result, err
}

func bugCreateForm(productID int, req zentao.BugCreateRequest) neturl.Values {
	form := neturl.Values{}
	form.Set("product", strconv.Itoa(productID))
	form.Set("title", req.Title)
	form.Set("openedBuild", strings.Join(req.OpenedBuild, ","))
	form.Set("assignedTo", req.AssignedTo)
	form.Set("type", req.Type)
	form.Set("severity", strconv.Itoa(req.Severity))
	form.Set("pri", strconv.Itoa(req.Pri))
	form.Set("steps", req.Steps)
	form.Set("os", req.OS)
	form.Set("browser", req.Browser)
	form.Set("hardware", req.Hardware)
	if req.Module != 0 {
		form.Set("module", strconv.Itoa(req.Module))
	}
	if req.Execution != 0 {
		form.Set("execution", strconv.Itoa(req.Execution))
	}
	if req.Project != 0 {
		form.Set("project", strconv.Itoa(req.Project))
	}
	return form
}

func (c *Client) UpdateBug(bugID int, req zentao.BugUpdateRequest) (*zentao.Bug, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("title", req.Title)
		form.Set("type", req.Type)
		form.Set("severity", strconv.Itoa(req.Severity))
		form.Set("pri", strconv.Itoa(req.Pri))
		form.Set("steps", req.Steps)
		form.Set("assignedTo", req.AssignedTo)
		form.Set("openedBuild", strings.Join(req.OpenedBuild, ","))
		if req.Module != 0 {
			form.Set("module", strconv.Itoa(req.Module))
		}
		if err := c.doSessionPost(context.Background(), "UpdateBug",
			fmt.Sprintf("/bug-edit-%d.json", bugID), form); err != nil {
			return nil, err
		}
		return &zentao.Bug{}, nil
	}
	var result *zentao.Bug
	err := c.withTokenRetry("UpdateBug", func(client *zentao.Client) error {
		var e error
		result, e = client.UpdateBug(bugID, req)
		return e
	})
	return result, err
}

func (c *Client) DeleteBug(bugID int) error {
	if c.IsSessionMode() {
		return c.doSessionPost(context.Background(), "DeleteBug",
			fmt.Sprintf("/bug-delete-%d.json", bugID), neturl.Values{})
	}
	return c.withTokenRetry("DeleteBug", func(client *zentao.Client) error {
		return client.DeleteBug(bugID)
	})
}

// ---- Task create/update/delete/activate ----

func (c *Client) CreateTask(executionID int, req zentao.TaskCreateRequest) (*zentao.Task, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("execution", strconv.Itoa(executionID))
		form.Set("name", req.Name)
		form.Set("type", req.Type)
		form.Set("assignedTo", strings.Join(req.AssignedTo, ","))
		form.Set("estimate", strconv.FormatFloat(req.Estimate, 'f', -1, 64))
		form.Set("pri", strconv.Itoa(req.Pri))
		form.Set("desc", req.Desc)
		if req.Module != 0 {
			form.Set("module", strconv.Itoa(req.Module))
		}
		if req.Story != 0 {
			form.Set("story", strconv.Itoa(req.Story))
		}
		if err := c.doSessionPost(context.Background(), "CreateTask",
			"/task-create.json", form); err != nil {
			return nil, err
		}
		return &zentao.Task{}, nil
	}
	var result *zentao.Task
	err := c.withTokenRetry("CreateTask", func(client *zentao.Client) error {
		var e error
		result, e = client.CreateTask(executionID, req)
		return e
	})
	return result, err
}

func (c *Client) UpdateTask(taskID int, req zentao.TaskUpdateRequest) (*zentao.Task, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("name", req.Name)
		form.Set("type", req.Type)
		form.Set("assignedTo", strings.Join(req.AssignedTo, ","))
		form.Set("estimate", strconv.FormatFloat(req.Estimate, 'f', -1, 64))
		form.Set("consumed", strconv.FormatFloat(req.Consumed, 'f', -1, 64))
		form.Set("left", strconv.FormatFloat(req.Left, 'f', -1, 64))
		form.Set("pri", strconv.Itoa(req.Pri))
		if err := c.doSessionPost(context.Background(), "UpdateTask",
			fmt.Sprintf("/task-edit-%d.json", taskID), form); err != nil {
			return nil, err
		}
		return &zentao.Task{}, nil
	}
	var result *zentao.Task
	err := c.withTokenRetry("UpdateTask", func(client *zentao.Client) error {
		var e error
		result, e = client.UpdateTask(taskID, req)
		return e
	})
	return result, err
}

func (c *Client) DeleteTask(taskID int) error {
	if c.IsSessionMode() {
		return c.doSessionPost(context.Background(), "DeleteTask",
			fmt.Sprintf("/task-delete-%d.json", taskID), neturl.Values{})
	}
	return c.withTokenRetry("DeleteTask", func(client *zentao.Client) error {
		return client.DeleteTask(taskID)
	})
}

func (c *Client) ActivateTask(taskID int, consumed, left float64) error {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("consumed", strconv.FormatFloat(consumed, 'f', -1, 64))
		form.Set("left", strconv.FormatFloat(left, 'f', -1, 64))
		return c.doSessionPost(context.Background(), "ActivateTask",
			fmt.Sprintf("/task-restart-%d.json", taskID), form)
	}
	return c.withTokenRetry("ActivateTask", func(client *zentao.Client) error {
		return client.ActivateTask(taskID, consumed, left)
	})
}

// ---- Story create/update/delete ----

func (c *Client) CreateStory(req zentao.StoryCreateRequest) (*zentao.Story, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("product", strconv.Itoa(req.Product))
		form.Set("title", req.Title)
		form.Set("pri", strconv.Itoa(req.Pri))
		form.Set("category", req.Category)
		form.Set("spec", req.Spec)
		form.Set("verify", req.Verify)
		form.Set("estimate", strconv.FormatFloat(req.Estimate, 'f', -1, 64))
		form.Set("assignedTo", req.AssignedTo)
		if req.Module != 0 {
			form.Set("module", strconv.Itoa(req.Module))
		}
		if err := c.doSessionPost(context.Background(), "CreateStory",
			"/story-create.json", form); err != nil {
			return nil, err
		}
		return &zentao.Story{}, nil
	}
	var result *zentao.Story
	err := c.withTokenRetry("CreateStory", func(client *zentao.Client) error {
		var e error
		result, e = client.CreateStory(req)
		return e
	})
	return result, err
}

func (c *Client) UpdateStory(storyID int, req zentao.StoryUpdateRequest) (*zentao.Story, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("title", req.Title)
		form.Set("spec", req.Spec)
		form.Set("verify", req.Verify)
		form.Set("type", req.Type)
		form.Set("pri", strconv.Itoa(req.Pri))
		form.Set("estimate", strconv.FormatFloat(req.Estimate, 'f', -1, 64))
		form.Set("assignedTo", req.AssignedTo)
		if err := c.doSessionPost(context.Background(), "UpdateStory",
			fmt.Sprintf("/story-edit-%d.json", storyID), form); err != nil {
			return nil, err
		}
		return &zentao.Story{}, nil
	}
	var result *zentao.Story
	err := c.withTokenRetry("UpdateStory", func(client *zentao.Client) error {
		var e error
		result, e = client.UpdateStory(storyID, req)
		return e
	})
	return result, err
}

func (c *Client) DeleteStory(storyID int) error {
	if c.IsSessionMode() {
		return c.doSessionPost(context.Background(), "DeleteStory",
			fmt.Sprintf("/story-delete-%d.json", storyID), neturl.Values{})
	}
	return c.withTokenRetry("DeleteStory", func(client *zentao.Client) error {
		return client.DeleteStory(storyID)
	})
}

// ---- Effort record ----

func (c *Client) RecordEffort(taskID int, date string, consumed, left float64, work string) error {
	if c.IsSessionMode() {
		if date == "" {
			date = nowDate()
		}
		form := neturl.Values{}
		form.Set("dates[]", date)
		form.Set("consumed[]", strconv.FormatFloat(consumed, 'f', -1, 64))
		form.Set("left[]", strconv.FormatFloat(left, 'f', -1, 64))
		form.Set("work[]", work)
		return c.doSessionPost(context.Background(), "RecordEffort",
			fmt.Sprintf("/task-recordEstimate-%d.json", taskID), form)
	}
	return c.withTokenRetry("RecordEffort", func(client *zentao.Client) error {
		return client.RecordEffort(taskID, date, consumed, left, work)
	})
}

// ---- Plan create/update/delete + link/unlink ----

func (c *Client) CreatePlan(productID int, req zentao.PlanCreateRequest) (*zentao.Plan, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("product", strconv.Itoa(productID))
		form.Set("title", req.Title)
		form.Set("begin", req.Begin)
		form.Set("end", req.End)
		form.Set("desc", req.Desc)
		if err := c.doSessionPost(context.Background(), "CreatePlan",
			"/plan-create.json", form); err != nil {
			return nil, err
		}
		return &zentao.Plan{}, nil
	}
	var result *zentao.Plan
	err := c.withTokenRetry("CreatePlan", func(client *zentao.Client) error {
		var e error
		result, e = client.CreatePlan(productID, req)
		return e
	})
	return result, err
}

func (c *Client) UpdatePlan(planID int, req zentao.PlanCreateRequest) (*zentao.Plan, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("title", req.Title)
		form.Set("begin", req.Begin)
		form.Set("end", req.End)
		form.Set("desc", req.Desc)
		if err := c.doSessionPost(context.Background(), "UpdatePlan",
			fmt.Sprintf("/plan-edit-%d.json", planID), form); err != nil {
			return nil, err
		}
		return &zentao.Plan{}, nil
	}
	var result *zentao.Plan
	err := c.withTokenRetry("UpdatePlan", func(client *zentao.Client) error {
		var e error
		result, e = client.UpdatePlan(planID, req)
		return e
	})
	return result, err
}

func (c *Client) DeletePlan(planID int) error {
	if c.IsSessionMode() {
		return c.doSessionPost(context.Background(), "DeletePlan",
			fmt.Sprintf("/plan-delete-%d.json", planID), neturl.Values{})
	}
	return c.withTokenRetry("DeletePlan", func(client *zentao.Client) error {
		return client.DeletePlan(planID)
	})
}

func (c *Client) LinkStoriesToPlan(planID int, storyIDs []int) error {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("stories", joinInts(storyIDs))
		return c.doSessionPost(context.Background(), "LinkStoriesToPlan",
			fmt.Sprintf("/plan-linkStory-%d.json", planID), form)
	}
	return c.withTokenRetry("LinkStoriesToPlan", func(client *zentao.Client) error {
		return client.LinkStoriesToPlan(planID, storyIDs)
	})
}

func (c *Client) UnlinkStoriesFromPlan(planID int, storyIDs []int) error {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("stories", joinInts(storyIDs))
		return c.doSessionPost(context.Background(), "UnlinkStoriesFromPlan",
			fmt.Sprintf("/plan-unlinkStory-%d.json", planID), form)
	}
	return c.withTokenRetry("UnlinkStoriesFromPlan", func(client *zentao.Client) error {
		return client.UnlinkStoriesFromPlan(planID, storyIDs)
	})
}

func (c *Client) LinkBugsToPlan(planID int, bugIDs []int) error {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("bugs", joinInts(bugIDs))
		return c.doSessionPost(context.Background(), "LinkBugsToPlan",
			fmt.Sprintf("/plan-linkBug-%d.json", planID), form)
	}
	return c.withTokenRetry("LinkBugsToPlan", func(client *zentao.Client) error {
		return client.LinkBugsToPlan(planID, bugIDs)
	})
}

func (c *Client) UnlinkBugsFromPlan(planID int, bugIDs []int) error {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("bugs", joinInts(bugIDs))
		return c.doSessionPost(context.Background(), "UnlinkBugsFromPlan",
			fmt.Sprintf("/plan-unlinkBug-%d.json", planID), form)
	}
	return c.withTokenRetry("UnlinkBugsFromPlan", func(client *zentao.Client) error {
		return client.UnlinkBugsFromPlan(planID, bugIDs)
	})
}

// ---- Case create/update/delete ----

func (c *Client) CreateCase(productID int, req zentao.CaseCreateRequest) (*zentao.Case, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("product", strconv.Itoa(productID))
		form.Set("title", req.Title)
		form.Set("type", req.Type)
		form.Set("pri", strconv.Itoa(req.Pri))
		form.Set("precondition", req.Precondition)
		form.Set("keywords", req.Keywords)
		if req.Module != 0 {
			form.Set("module", strconv.Itoa(req.Module))
		}
		if req.Story != 0 {
			form.Set("story", strconv.Itoa(req.Story))
		}
		if err := c.doSessionPost(context.Background(), "CreateCase",
			"/testcase-create.json", form); err != nil {
			return nil, err
		}
		return &zentao.Case{}, nil
	}
	var result *zentao.Case
	err := c.withTokenRetry("CreateCase", func(client *zentao.Client) error {
		var e error
		result, e = client.CreateCase(productID, req)
		return e
	})
	return result, err
}

func (c *Client) UpdateCase(caseID int, req zentao.CaseUpdateRequest) (*zentao.Case, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("title", req.Title)
		form.Set("type", req.Type)
		form.Set("pri", strconv.Itoa(req.Pri))
		form.Set("precondition", req.Precondition)
		if err := c.doSessionPost(context.Background(), "UpdateCase",
			fmt.Sprintf("/testcase-edit-%d.json", caseID), form); err != nil {
			return nil, err
		}
		return &zentao.Case{}, nil
	}
	var result *zentao.Case
	err := c.withTokenRetry("UpdateCase", func(client *zentao.Client) error {
		var e error
		result, e = client.UpdateCase(caseID, req)
		return e
	})
	return result, err
}

func (c *Client) DeleteCase(caseID int) error {
	if c.IsSessionMode() {
		return c.doSessionPost(context.Background(), "DeleteCase",
			fmt.Sprintf("/testcase-delete-%d.json", caseID), neturl.Values{})
	}
	return c.withTokenRetry("DeleteCase", func(client *zentao.Client) error {
		return client.DeleteCase(caseID)
	})
}

// ---- Ticket create/update/delete ----

func (c *Client) CreateTicket(req zentao.TicketCreateRequest) (*zentao.Ticket, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("product", strconv.Itoa(req.Product))
		form.Set("module", strconv.Itoa(req.Module))
		form.Set("title", req.Title)
		form.Set("type", req.Type)
		if err := c.doSessionPost(context.Background(), "CreateTicket",
			"/ticket-create.json", form); err != nil {
			return nil, err
		}
		return &zentao.Ticket{}, nil
	}
	var result *zentao.Ticket
	err := c.withTokenRetry("CreateTicket", func(client *zentao.Client) error {
		var e error
		result, e = client.CreateTicket(req)
		return e
	})
	return result, err
}

func (c *Client) UpdateTicket(ticketID int, req zentao.TicketUpdateRequest) (*zentao.Ticket, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("title", req.Title)
		form.Set("type", req.Type)
		form.Set("desc", req.Desc)
		form.Set("pri", strconv.Itoa(req.Pri))
		form.Set("assignedTo", req.AssignedTo)
		if err := c.doSessionPost(context.Background(), "UpdateTicket",
			fmt.Sprintf("/ticket-edit-%d.json", ticketID), form); err != nil {
			return nil, err
		}
		return &zentao.Ticket{}, nil
	}
	var result *zentao.Ticket
	err := c.withTokenRetry("UpdateTicket", func(client *zentao.Client) error {
		var e error
		result, e = client.UpdateTicket(ticketID, req)
		return e
	})
	return result, err
}

func (c *Client) DeleteTicket(ticketID int) error {
	if c.IsSessionMode() {
		return c.doSessionPost(context.Background(), "DeleteTicket",
			fmt.Sprintf("/ticket-delete-%d.json", ticketID), neturl.Values{})
	}
	return c.withTokenRetry("DeleteTicket", func(client *zentao.Client) error {
		return client.DeleteTicket(ticketID)
	})
}

// ---- Feedback create/update/assign/close/delete ----

func (c *Client) CreateFeedback(req zentao.FeedbackCreateRequest) (*zentao.Feedback, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("product", strconv.Itoa(req.Product))
		if req.Module != 0 {
			form.Set("module", strconv.Itoa(req.Module))
		}
		form.Set("title", req.Title)
		form.Set("type", req.Type)
		form.Set("desc", req.Desc)
		form.Set("feedbackBy", req.FeedbackBy)
		if err := c.doSessionPost(context.Background(), "CreateFeedback",
			"/feedback-create.json", form); err != nil {
			return nil, err
		}
		return &zentao.Feedback{}, nil
	}
	var result *zentao.Feedback
	err := c.withTokenRetry("CreateFeedback", func(client *zentao.Client) error {
		var e error
		result, e = client.CreateFeedback(req)
		return e
	})
	return result, err
}

func (c *Client) UpdateFeedback(feedbackID int, req zentao.FeedbackUpdateRequest) (*zentao.Feedback, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("title", req.Title)
		form.Set("type", req.Type)
		form.Set("desc", req.Desc)
		if req.Product != 0 {
			form.Set("product", strconv.Itoa(req.Product))
		}
		if err := c.doSessionPost(context.Background(), "UpdateFeedback",
			fmt.Sprintf("/feedback-edit-%d.json", feedbackID), form); err != nil {
			return nil, err
		}
		return &zentao.Feedback{}, nil
	}
	var result *zentao.Feedback
	err := c.withTokenRetry("UpdateFeedback", func(client *zentao.Client) error {
		var e error
		result, e = client.UpdateFeedback(feedbackID, req)
		return e
	})
	return result, err
}

func (c *Client) AssignFeedback(feedbackID int, req zentao.FeedbackAssignRequest) (*zentao.Feedback, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("assignedTo", req.AssignedTo)
		if req.Comment != "" {
			form.Set("comment", req.Comment)
		}
		if err := c.doSessionPost(context.Background(), "AssignFeedback",
			fmt.Sprintf("/feedback-assignTo-%d.json", feedbackID), form); err != nil {
			return nil, err
		}
		return &zentao.Feedback{}, nil
	}
	var result *zentao.Feedback
	err := c.withTokenRetry("AssignFeedback", func(client *zentao.Client) error {
		var e error
		result, e = client.AssignFeedback(feedbackID, req)
		return e
	})
	return result, err
}

func (c *Client) CloseFeedback(feedbackID int, req zentao.FeedbackCloseRequest) (*zentao.Feedback, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("closedReason", req.ClosedReason)
		if req.Comment != "" {
			form.Set("comment", req.Comment)
		}
		if err := c.doSessionPost(context.Background(), "CloseFeedback",
			fmt.Sprintf("/feedback-close-%d.json", feedbackID), form); err != nil {
			return nil, err
		}
		return &zentao.Feedback{}, nil
	}
	var result *zentao.Feedback
	err := c.withTokenRetry("CloseFeedback", func(client *zentao.Client) error {
		var e error
		result, e = client.CloseFeedback(feedbackID, req)
		return e
	})
	return result, err
}

func (c *Client) DeleteFeedback(feedbackID int) error {
	if c.IsSessionMode() {
		return c.doSessionPost(context.Background(), "DeleteFeedback",
			fmt.Sprintf("/feedback-delete-%d.json", feedbackID), neturl.Values{})
	}
	return c.withTokenRetry("DeleteFeedback", func(client *zentao.Client) error {
		return client.DeleteFeedback(feedbackID)
	})
}

// ---- Product/Project/Program/Execution/Build/User create/update/delete ----
// 这些是基础实体管理，session 模式同样走 PHP 表单端点。

func (c *Client) CreateProduct(req zentao.ProductCreateRequest) (*zentao.Product, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("name", req.Name)
		form.Set("code", req.Code)
		form.Set("type", req.Type)
		form.Set("desc", req.Desc)
		if err := c.doSessionPost(context.Background(), "CreateProduct",
			"/product-create.json", form); err != nil {
			return nil, err
		}
		return &zentao.Product{}, nil
	}
	var result *zentao.Product
	err := c.withTokenRetry("CreateProduct", func(client *zentao.Client) error {
		var e error
		result, e = client.CreateProduct(req)
		return e
	})
	return result, err
}

func (c *Client) UpdateProduct(productID int, req zentao.ProductCreateRequest) (*zentao.Product, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("name", req.Name)
		form.Set("code", req.Code)
		form.Set("type", req.Type)
		if err := c.doSessionPost(context.Background(), "UpdateProduct",
			fmt.Sprintf("/product-edit-%d.json", productID), form); err != nil {
			return nil, err
		}
		return &zentao.Product{}, nil
	}
	var result *zentao.Product
	err := c.withTokenRetry("UpdateProduct", func(client *zentao.Client) error {
		var e error
		result, e = client.UpdateProduct(productID, req)
		return e
	})
	return result, err
}

func (c *Client) DeleteProduct(productID int) error {
	if c.IsSessionMode() {
		return c.doSessionPost(context.Background(), "DeleteProduct",
			fmt.Sprintf("/product-delete-%d.json", productID), neturl.Values{})
	}
	return c.withTokenRetry("DeleteProduct", func(client *zentao.Client) error {
		return client.DeleteProduct(productID)
	})
}

func (c *Client) CreateProject(req zentao.ProjectCreateRequest) (*zentao.Project, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("name", req.Name)
		form.Set("code", req.Code)
		form.Set("model", req.Model)
		form.Set("begin", req.Begin)
		form.Set("end", req.End)
		if err := c.doSessionPost(context.Background(), "CreateProject",
			"/project-create.json", form); err != nil {
			return nil, err
		}
		return &zentao.Project{}, nil
	}
	var result *zentao.Project
	err := c.withTokenRetry("CreateProject", func(client *zentao.Client) error {
		var e error
		result, e = client.CreateProject(req)
		return e
	})
	return result, err
}

func (c *Client) UpdateProject(projectID int, req zentao.ProjectCreateRequest) (*zentao.Project, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("name", req.Name)
		form.Set("code", req.Code)
		if err := c.doSessionPost(context.Background(), "UpdateProject",
			fmt.Sprintf("/project-edit-%d.json", projectID), form); err != nil {
			return nil, err
		}
		return &zentao.Project{}, nil
	}
	var result *zentao.Project
	err := c.withTokenRetry("UpdateProject", func(client *zentao.Client) error {
		var e error
		result, e = client.UpdateProject(projectID, req)
		return e
	})
	return result, err
}

func (c *Client) DeleteProject(projectID int) error {
	if c.IsSessionMode() {
		return c.doSessionPost(context.Background(), "DeleteProject",
			fmt.Sprintf("/project-delete-%d.json", projectID), neturl.Values{})
	}
	return c.withTokenRetry("DeleteProject", func(client *zentao.Client) error {
		return client.DeleteProject(projectID)
	})
}

func (c *Client) CreateProgram(req zentao.ProgramCreateRequest) (*zentao.Program, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("name", req.Name)
		form.Set("code", req.Code)
		form.Set("begin", req.Begin)
		form.Set("end", req.End)
		if err := c.doSessionPost(context.Background(), "CreateProgram",
			"/program-create.json", form); err != nil {
			return nil, err
		}
		return &zentao.Program{}, nil
	}
	var result *zentao.Program
	err := c.withTokenRetry("CreateProgram", func(client *zentao.Client) error {
		var e error
		result, e = client.CreateProgram(req)
		return e
	})
	return result, err
}

func (c *Client) UpdateProgram(programID int, req zentao.ProgramCreateRequest) (*zentao.Program, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("name", req.Name)
		form.Set("code", req.Code)
		if err := c.doSessionPost(context.Background(), "UpdateProgram",
			fmt.Sprintf("/program-edit-%d.json", programID), form); err != nil {
			return nil, err
		}
		return &zentao.Program{}, nil
	}
	var result *zentao.Program
	err := c.withTokenRetry("UpdateProgram", func(client *zentao.Client) error {
		var e error
		result, e = client.UpdateProgram(programID, req)
		return e
	})
	return result, err
}

func (c *Client) DeleteProgram(programID int) error {
	if c.IsSessionMode() {
		return c.doSessionPost(context.Background(), "DeleteProgram",
			fmt.Sprintf("/program-delete-%d.json", programID), neturl.Values{})
	}
	return c.withTokenRetry("DeleteProgram", func(client *zentao.Client) error {
		return client.DeleteProgram(programID)
	})
}

func (c *Client) CreateExecution(projectID int, req zentao.ExecutionCreateRequest) (*zentao.Execution, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("project", strconv.Itoa(projectID))
		form.Set("name", req.Name)
		form.Set("code", req.Code)
		form.Set("begin", req.Begin)
		form.Set("end", req.End)
		if err := c.doSessionPost(context.Background(), "CreateExecution",
			"/execution-create.json", form); err != nil {
			return nil, err
		}
		return &zentao.Execution{}, nil
	}
	var result *zentao.Execution
	err := c.withTokenRetry("CreateExecution", func(client *zentao.Client) error {
		var e error
		result, e = client.CreateExecution(projectID, req)
		return e
	})
	return result, err
}

func (c *Client) UpdateExecution(executionID int, req zentao.ExecutionCreateRequest) (*zentao.Execution, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("name", req.Name)
		form.Set("code", req.Code)
		if err := c.doSessionPost(context.Background(), "UpdateExecution",
			fmt.Sprintf("/execution-edit-%d.json", executionID), form); err != nil {
			return nil, err
		}
		return &zentao.Execution{}, nil
	}
	var result *zentao.Execution
	err := c.withTokenRetry("UpdateExecution", func(client *zentao.Client) error {
		var e error
		result, e = client.UpdateExecution(executionID, req)
		return e
	})
	return result, err
}

func (c *Client) DeleteExecution(executionID int) error {
	if c.IsSessionMode() {
		return c.doSessionPost(context.Background(), "DeleteExecution",
			fmt.Sprintf("/execution-delete-%d.json", executionID), neturl.Values{})
	}
	return c.withTokenRetry("DeleteExecution", func(client *zentao.Client) error {
		return client.DeleteExecution(executionID)
	})
}

func (c *Client) CreateBuild(projectID int, req zentao.BuildCreateRequest) (*zentao.Build, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("project", strconv.Itoa(projectID))
		form.Set("name", req.Name)
		form.Set("date", req.Date)
		if req.Execution != 0 {
			form.Set("execution", strconv.Itoa(req.Execution))
		}
		if err := c.doSessionPost(context.Background(), "CreateBuild",
			"/build-create.json", form); err != nil {
			return nil, err
		}
		return &zentao.Build{}, nil
	}
	var result *zentao.Build
	err := c.withTokenRetry("CreateBuild", func(client *zentao.Client) error {
		var e error
		result, e = client.CreateBuild(projectID, req)
		return e
	})
	return result, err
}

func (c *Client) UpdateBuild(buildID int, req zentao.BuildCreateRequest) (*zentao.Build, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("name", req.Name)
		if err := c.doSessionPost(context.Background(), "UpdateBuild",
			fmt.Sprintf("/build-edit-%d.json", buildID), form); err != nil {
			return nil, err
		}
		return &zentao.Build{}, nil
	}
	var result *zentao.Build
	err := c.withTokenRetry("UpdateBuild", func(client *zentao.Client) error {
		var e error
		result, e = client.UpdateBuild(buildID, req)
		return e
	})
	return result, err
}

func (c *Client) DeleteBuild(buildID int) error {
	if c.IsSessionMode() {
		return c.doSessionPost(context.Background(), "DeleteBuild",
			fmt.Sprintf("/build-delete-%d.json", buildID), neturl.Values{})
	}
	return c.withTokenRetry("DeleteBuild", func(client *zentao.Client) error {
		return client.DeleteBuild(buildID)
	})
}

func (c *Client) CreateUser(req zentao.UserCreateRequest) (*zentao.User, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("account", req.Account)
		form.Set("password", req.Password)
		form.Set("realname", req.Realname)
		form.Set("email", req.Email)
		if err := c.doSessionPost(context.Background(), "CreateUser",
			"/user-create.json", form); err != nil {
			return nil, err
		}
		return &zentao.User{}, nil
	}
	var result *zentao.User
	err := c.withTokenRetry("CreateUser", func(client *zentao.Client) error {
		var e error
		result, e = client.CreateUser(req)
		return e
	})
	return result, err
}

func (c *Client) UpdateUser(userID int, req zentao.UserUpdateRequest) (*zentao.User, error) {
	if c.IsSessionMode() {
		form := neturl.Values{}
		form.Set("realname", req.Realname)
		form.Set("email", req.Email)
		if err := c.doSessionPost(context.Background(), "UpdateUser",
			fmt.Sprintf("/user-edit-%d.json", userID), form); err != nil {
			return nil, err
		}
		return &zentao.User{}, nil
	}
	var result *zentao.User
	err := c.withTokenRetry("UpdateUser", func(client *zentao.Client) error {
		var e error
		result, e = client.UpdateUser(userID, req)
		return e
	})
	return result, err
}

func (c *Client) DeleteUser(userID int) error {
	if c.IsSessionMode() {
		return c.doSessionPost(context.Background(), "DeleteUser",
			fmt.Sprintf("/user-delete-%d.json", userID), neturl.Values{})
	}
	return c.withTokenRetry("DeleteUser", func(client *zentao.Client) error {
		return client.DeleteUser(userID)
	})
}

// nowDate 返回当天日期 YYYY-MM-DD（RecordEffort 默认日期用）。
func nowDate() string {
	return timeNow().Format("2006-01-02")
}
