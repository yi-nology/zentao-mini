package zentao

import (
	"github.com/yi-nology/common/biz/zentao"
)

// Bug 写操作：走禅道 REST API v1（POST /api.php/v1/bugs/{id}/{confirm|resolve|close|activate|assign}）。
// 禅道官方 API 没有独立的"加备注"端点，纯评论见 client_bug_comment.go 的 Web 会话通道；
// 但所有状态流转接口都接受 comment 字段，操作历史里会带上备注。

// ConfirmBug 确认 Bug
func (c *Client) ConfirmBug(bugID int, comment string) error {
	return c.withTokenRetry("ConfirmBug", func(sdk *zentao.Client) error {
		return sdk.ConfirmBug(bugID, zentao.BugConfirmRequest{Comment: comment})
	})
}

// ResolveBug 解决 Bug。resolution 取值：fixed/bydesign/duplicate/notrepro/postponed/willnotfix/external
func (c *Client) ResolveBug(bugID int, resolution, resolvedBuild, comment string) error {
	req := zentao.BugResolveRequest{
		Resolution:    resolution,
		ResolvedBuild: resolvedBuild,
		Comment:       comment,
	}
	return c.withTokenRetry("ResolveBug", func(sdk *zentao.Client) error {
		return sdk.ResolveBug(bugID, req)
	})
}

// CloseBug 关闭 Bug
func (c *Client) CloseBug(bugID int, comment string) error {
	return c.withTokenRetry("CloseBug", func(sdk *zentao.Client) error {
		return sdk.CloseBug(bugID, zentao.BugCloseRequest{Comment: comment})
	})
}

// ActivateBug 重新激活 Bug（resolved/closed 状态），可选重新指派
func (c *Client) ActivateBug(bugID int, assignedTo, comment string) error {
	req := zentao.BugActivateRequest{
		AssignedTo: assignedTo,
		Comment:    comment,
	}
	return c.withTokenRetry("ActivateBug", func(sdk *zentao.Client) error {
		return sdk.ActivateBug(bugID, req)
	})
}

// AssignBug 指派 Bug
func (c *Client) AssignBug(bugID int, assignedTo, comment string) error {
	req := zentao.BugAssignRequest{
		AssignedTo: assignedTo,
		Comment:    comment,
	}
	return c.withTokenRetry("AssignBug", func(sdk *zentao.Client) error {
		return sdk.AssignBug(bugID, req)
	})
}

// BugValidResolutions 禅道支持的解决方式（供上层校验用）
var BugValidResolutions = []string{"fixed", "bydesign", "duplicate", "notrepro", "postponed", "willnotfix", "external"}

// BugValidTransitions 禅道支持的状态流转动作
var BugValidTransitions = []string{"confirm", "resolve", "close", "activate", "assign"}
