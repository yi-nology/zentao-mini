package zentao

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yi-nology/common/biz/zentao"
)

// ---- Tickets (session mode) ----
// 端点：/ticket-browse-{productID}.json 或 /ticket-browse.json。
// 麒麟实例上 ticket 模块通常需额外权限，账号无权限时返回 user-deny → 返回空。
// 内层 tickets 是 list。

type sessionTicketBrowseResp struct {
	Title   string          `json:"title"`
	Tickets []sessionTicket `json:"tickets"`
	Pager   *sessionPager   `json:"pager"`
}

type sessionTicket struct {
	ID           int    `json:"id"`
	Product      int    `json:"product"`
	Module       int    `json:"module"`
	Title        string `json:"title"`
	Type         string `json:"type"`
	Desc         string `json:"desc"`
	OpenedBuild  string `json:"openedBuild"`
	AssignedTo   string `json:"assignedTo"`
	Pri          int    `json:"pri"`
	Estimate     float64 `json:"estimate"`
	Left         float64 `json:"left"`
	Status       string `json:"status"`
	OpenedBy     string `json:"openedBy"`
	OpenedDate   string `json:"openedDate"`
	Deadline     string `json:"deadline"`
	ResolvedBy   string `json:"resolvedBy"`
	Resolution   string `json:"resolution"`
	ClosedBy     string `json:"closedBy"`
	Deleted      string `json:"deleted"`
}

func (t sessionTicket) toSDK() zentao.Ticket {
	return zentao.Ticket{
		ID:          t.ID,
		Product:     t.Product,
		Module:      t.Module,
		Title:       t.Title,
		Type:        t.Type,
		Desc:        t.Desc,
		OpenedBuild: t.OpenedBuild,
		AssignedTo:  &t.AssignedTo,
		Pri:         t.Pri,
		Estimate:    t.Estimate,
		Left:        t.Left,
		Status:      t.Status,
		OpenedBy:    zentao.UserRef{Account: t.OpenedBy},
		OpenedDate:  t.OpenedDate,
		Deadline:    &t.Deadline,
		ResolvedBy:  t.ResolvedBy,
		Resolution:  t.Resolution,
		ClosedBy:    &t.ClosedBy,
	}
}

func (c *Client) getTicketsSession(ctx context.Context, productID, page, pageSize int) ([]zentao.Ticket, error) {
	var path string
	if productID > 0 {
		path = fmt.Sprintf("/ticket-browse-%d.json", productID)
	} else {
		path = "/ticket-browse.json"
	}
	var resp sessionTicketBrowseResp
	if err := c.doSessionJSON(ctx, "GetTickets", path, &resp); err != nil {
		if IsSessionAccessDenied(err) {
			return nil, nil
		}
		return nil, err
	}
	all := make([]zentao.Ticket, 0, len(resp.Tickets))
	for _, t := range resp.Tickets {
		all = append(all, t.toSDK())
	}
	if pageSize <= 0 {
		return all, nil
	}
	return doPaginate(all, page, pageSize), nil
}

// ---- Feedbacks (session mode) ----
// 端点：/feedback-admin-browse.json（管理员视图，麒麟实例普通账号可访问）。
// 内层 feedbacks 是 list。

type sessionFeedbackBrowseResp struct {
	Title      string             `json:"title"`
	Feedbacks  []sessionFeedback  `json:"feedbacks"`
	Pager      *sessionPager      `json:"pager"`
}

type sessionFeedback struct {
	ID         int    `json:"id"`
	Product    int    `json:"product"`
	Module     int    `json:"module"`
	Title      string `json:"title"`
	Type       string `json:"type"`
	Solution   string `json:"solution"`
	Desc       string `json:"desc"`
	Status     string `json:"status"`
	Public     json.Number `json:"public"`
	OpenedBy   string `json:"openedBy"`
	OpenedDate string `json:"openedDate"`
	AssignedTo string `json:"assignedTo"`
	ClosedBy   string `json:"closedBy"`
	ClosedDate string `json:"closedDate"`
	Deleted    json.Number `json:"deleted"`
}

func (f sessionFeedback) toSDK() zentao.Feedback {
	pub, _ := f.Public.Int64()
	del, _ := f.Deleted.Int64()
	assignedTo := &zentao.UserRef{Account: f.AssignedTo}
	closedBy := f.ClosedBy
	closedDate := f.ClosedDate
	return zentao.Feedback{
		ID:         f.ID,
		Product:    f.Product,
		Module:     f.Module,
		Title:      f.Title,
		Type:       f.Type,
		Solution:   f.Solution,
		Desc:       f.Desc,
		Status:     f.Status,
		Public:     int(pub),
		OpenedBy:   zentao.UserRef{Account: f.OpenedBy},
		OpenedDate: f.OpenedDate,
		AssignedTo: assignedTo,
		ClosedBy:   &closedBy,
		ClosedDate: &closedDate,
		Deleted:    int(del),
	}
}

func (c *Client) getFeedbacksSession(ctx context.Context, page, pageSize int) ([]zentao.Feedback, error) {
	var resp sessionFeedbackBrowseResp
	if err := c.doSessionJSON(ctx, "GetFeedbacks", "/feedback-admin-browse.json", &resp); err != nil {
		if IsSessionAccessDenied(err) {
			return nil, nil
		}
		return nil, err
	}
	all := make([]zentao.Feedback, 0, len(resp.Feedbacks))
	for _, f := range resp.Feedbacks {
		all = append(all, f.toSDK())
	}
	if pageSize <= 0 {
		return all, nil
	}
	return doPaginate(all, page, pageSize), nil
}
