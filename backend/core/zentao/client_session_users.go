package zentao

import (
	"context"
	"fmt"

	"github.com/yi-nology/common/biz/zentao"
)

// ---- Users (session mode) ----
//
// 麒麟实例上 /user-browse.json 返回 0 字节（账号无用户管理权限）。
// 会话模式下：当前用户走 /user-profile.json（拿完整 user 对象）；
// 用户列表能力受限，GetUsersAll 在权限允许时走 /user-browse.json，
// 不允许时返回仅含当前用户的单元素列表（保证 dashboard 的 current-user 合成可用）。

type sessionUserProfileResp struct {
	Title    string      `json:"title"`
	User     sessionUser `json:"user"`
	// DeptPath 在部分实例上是数组，这里不做强类型解析。
	DeptPath interface{} `json:"deptPath"`
}

type sessionUser struct {
	ID       int    `json:"id"`
	Account  string `json:"account"`
	Realname string `json:"realname"`
	Role     string `json:"role"`
	Dept     int    `json:"dept"`
	Email    string `json:"email"`
	Gender   string `json:"gender"`
	Phone    string `json:"phone"`
	Status   string `json:"status"`
	Join     string `json:"join"`
}

func (u sessionUser) toSDK() zentao.User {
	return zentao.User{
		ID:       u.ID,
		Account:  u.Account,
		Realname: u.Realname,
		Role:     u.Role,
		Dept:     u.Dept,
		Email:    u.Email,
		Gender:   u.Gender,
		Phone:    u.Phone,
		Status:   u.Status,
		Join:     u.Join,
	}
}

// getCurrentUserSession 通过 /user-profile.json 获取当前登录用户。
func (c *Client) getCurrentUserSession(ctx context.Context) (*zentao.User, error) {
	var resp sessionUserProfileResp
	if err := c.doSessionJSON(ctx, "GetCurrentUser", "/user-profile.json", &resp); err != nil {
		return nil, err
	}
	u := resp.User.toSDK()
	return &u, nil
}

// getUsersAllSession 获取用户列表。权限允许时走 /user-browse.json；
// 否则回退到仅当前用户。前端 dashboard 的 current-user 合成依赖此方法。
func (c *Client) getUsersAllSession(ctx context.Context) ([]zentao.User, error) {
	var resp struct {
		Title    string                `json:"title"`
		Users    map[string]sessionUser `json:"users"`
		Members  map[string]sessionUser `json:"memberPairs"`
		Pager    *sessionPager          `json:"pager"`
	}
	// /user-browse.json：权限允许时返回用户列表；不允许时返回空 data 或 deny。
	if err := c.doSessionJSON(ctx, "GetUsersAll", "/user-browse.json", &resp); err != nil {
		if IsSessionAccessDenied(err) {
			// 回退：仅当前用户。
			me, merr := c.getCurrentUserSession(ctx)
			if merr != nil {
				return nil, merr
			}
			return []zentao.User{*me}, nil
		}
		return nil, err
	}
	out := make([]zentao.User, 0, len(resp.Users))
	for _, u := range resp.Users {
		out = append(out, u.toSDK())
	}
	if len(out) == 0 {
		// data 为空（麒麟实例行为）：回退到当前用户。
		me, err := c.getCurrentUserSession(ctx)
		if err != nil {
			return nil, err
		}
		return []zentao.User{*me}, nil
	}
	return out, nil
}

// getUsersSession 分页包装 GetUsersAll（内存分页，与 Token 模式行为一致）。
func (c *Client) getUsersSession(ctx context.Context, page, pageSize int) ([]zentao.User, error) {
	all, err := c.getUsersAllSession(ctx)
	if err != nil {
		return nil, err
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

// getAccountNameSession 返回当前登录账号名（与 c.GetAccount() 一致，
// 但优先用 user-profile 的 account 字段，更贴近服务端实际态）。
func (c *Client) getAccountNameSession(ctx context.Context) string {
	if u, err := c.getCurrentUserSession(ctx); err == nil && u.Account != "" {
		return u.Account
	}
	return c.GetAccount()
}

var _ = fmt.Sprintf // 保留 fmt 以便后续扩展
