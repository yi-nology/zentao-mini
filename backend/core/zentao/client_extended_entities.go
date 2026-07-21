package zentao

import (
	"context"

	"github.com/yi-nology/common/biz/zentao"
)

// 本文件为 SDK 已支持但 wrapper 此前未暴露的 7 个实体补齐读取方法。
// 每个方法都做 token/session 双模式分发，签名与 SDK 对齐。

// ---- Cases ----

func (c *Client) GetCasesByProduct(productID, page, pageSize int) ([]zentao.Case, error) {
	if c.IsSessionMode() {
		return c.getCasesByProductSession(context.Background(), productID, page, pageSize)
	}
	var response *zentao.CaseListResponse
	err := c.withTokenRetry("GetCasesByProduct", func(client *zentao.Client) error {
		var err error
		response, err = client.GetCasesByProduct(productID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Testcases, nil
}

func (c *Client) GetAllCases(productID int) ([]zentao.Case, error) {
	if c.IsSessionMode() {
		return c.getAllCasesSession(context.Background(), productID)
	}
	var all []zentao.Case
	page := 1
	for {
		cases, err := c.GetCasesByProduct(productID, page, 100)
		if err != nil {
			return all, err
		}
		all = append(all, cases...)
		if len(cases) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// ---- Plans ----

func (c *Client) GetPlans(productID, page, pageSize int) ([]zentao.Plan, error) {
	if c.IsSessionMode() {
		return c.getPlansSession(context.Background(), productID, page, pageSize)
	}
	var response *zentao.PlanListResponse
	err := c.withTokenRetry("GetPlans", func(client *zentao.Client) error {
		var err error
		response, err = client.GetPlans(productID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Plans, nil
}

// ---- Programs ----

func (c *Client) GetPrograms(page, pageSize int) ([]zentao.Program, error) {
	if c.IsSessionMode() {
		return c.getProgramsSession(context.Background(), page, pageSize)
	}
	var response *zentao.ProgramListResponse
	err := c.withTokenRetry("GetPrograms", func(client *zentao.Client) error {
		var err error
		response, err = client.GetPrograms(page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Programs, nil
}

// ---- Releases ----

func (c *Client) GetReleasesByProduct(productID, page, pageSize int) ([]zentao.Release, error) {
	if c.IsSessionMode() {
		return c.getReleasesByProductSession(context.Background(), productID, page, pageSize)
	}
	var response *zentao.ReleaseListResponse
	err := c.withTokenRetry("GetReleasesByProduct", func(client *zentao.Client) error {
		var err error
		response, err = client.GetReleasesByProduct(productID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Releases, nil
}

// ---- Test Tasks ----

func (c *Client) GetTestTasksByProduct(productID, page, pageSize int) ([]zentao.TestTask, error) {
	if c.IsSessionMode() {
		return c.getTestTasksSession(context.Background(), productID, page, pageSize)
	}
	// token 模式：SDK 的 GetTestTasksByProject 按 project 取；这里用 product 下所有 project 合并。
	projects, err := c.GetProjectsByProduct(productID, 1, 200)
	if err != nil {
		return nil, err
	}
	seen := make(map[int]bool)
	var all []zentao.TestTask
	for _, p := range projects {
		var resp *zentao.ProjectTestTaskListResponse
		if err := c.withTokenRetry("GetTestTasksByProject", func(client *zentao.Client) error {
			var e error
			resp, e = client.GetTestTasksByProject(p.ID, page, pageSize)
			return e
		}); err != nil {
			continue
		}
		for _, t := range resp.Testtasks {
			if !seen[t.ID] {
				seen[t.ID] = true
				all = append(all, t)
			}
		}
	}
	return all, nil
}

// ---- Tickets ----

func (c *Client) GetTickets(browseType string, param, page, pageSize int) ([]zentao.Ticket, error) {
	if c.IsSessionMode() {
		// session 模式 param 当 productID 用
		return c.getTicketsSession(context.Background(), param, page, pageSize)
	}
	var response *zentao.TicketListResponse
	err := c.withTokenRetry("GetTickets", func(client *zentao.Client) error {
		var err error
		response, err = client.GetTickets(browseType, param, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Tickets, nil
}

// ---- Feedbacks ----

func (c *Client) GetFeedbacks(page, pageSize int) ([]zentao.Feedback, error) {
	if c.IsSessionMode() {
		return c.getFeedbacksSession(context.Background(), page, pageSize)
	}
	var response *zentao.FeedbackListResponse
	err := c.withTokenRetry("GetFeedbacks", func(client *zentao.Client) error {
		var err error
		response, err = client.GetFeedbacks(page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Feedbacks, nil
}
