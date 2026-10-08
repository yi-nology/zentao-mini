package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yi-nology/common/biz/zentao"

	myzentao "github.com/yi-nology/zentao-mini/backend/core/zentao"
)

// Bug 写操作业务逻辑：入参校验 + 调用禅道客户端 + 返回流转后的 Bug 快照。
// 权限控制在 REST 层（平台认证：匿名只读），MCP 层（mcp.token + 只读模式）。

// BugTransitionInput Bug 状态流转入参
type BugTransitionInput struct {
	Action        string `json:"action"`                  // confirm/resolve/close/activate/assign
	Resolution    string `json:"resolution,omitempty"`    // resolve 必填：fixed/bydesign/duplicate/notrepro/postponed/willnotfix/external
	ResolvedBuild string `json:"resolvedBuild,omitempty"` // resolve 建议填写：解决版本
	AssignedTo    string `json:"assignedTo,omitempty"`    // assign 必填；activate 可选（重新指派）
	Comment       string `json:"comment,omitempty"`       // 任意动作可附带的备注，会记入操作历史
}

// validationError 入参校验失败（HTTP 400），区别于禅道侧错误（50002）。
// 复用包内既有的 ValidationError 类型（story_service.go 定义）。
func validationError(format string, args ...interface{}) error {
	return &ValidationError{Message: fmt.Sprintf(format, args...)}
}

// AddBugComment 给 Bug 添加备注，成功后返回 Bug 最新快照
func (s *BugService) AddBugComment(bugID int, comment string) (*zentao.Bug, error) {
	if bugID <= 0 {
		return nil, validationError("无效的 Bug ID: %d", bugID)
	}
	if strings.TrimSpace(comment) == "" {
		return nil, validationError("评论内容不能为空")
	}
	if err := s.client.AddBugComment(bugID, comment); err != nil {
		return nil, err
	}
	return s.client.GetBug(bugID)
}

// TransitionBug 执行 Bug 状态流转，成功后返回 Bug 最新快照。
// 禅道自身会校验状态机（如 resolved 状态不能再次 resolve），本层做入参校验，
// 并在 REST 端点缺失（老版本禅道无 activate）或静默失效时自动降级到 Web 会话通道。
func (s *BugService) TransitionBug(bugID int, input *BugTransitionInput) (*zentao.Bug, error) {
	if bugID <= 0 {
		return nil, validationError("无效的 Bug ID: %d", bugID)
	}
	if input == nil {
		return nil, validationError("缺少流转参数")
	}
	action := strings.ToLower(strings.TrimSpace(input.Action))

	switch action {
	case "confirm":
		if err := s.client.ConfirmBug(bugID, input.Comment); err != nil {
			return nil, err
		}
	case "resolve":
		resolution := strings.ToLower(strings.TrimSpace(input.Resolution))
		if resolution == "" {
			return nil, validationError("解决 Bug 必须指定 resolution（%s）", strings.Join(myzentao.BugValidResolutions, "/"))
		}
		if !isValidResolution(resolution) {
			return nil, validationError("不支持的 resolution: %s（支持 %s）", resolution, strings.Join(myzentao.BugValidResolutions, "/"))
		}
		if err := s.client.ResolveBug(bugID, resolution, input.ResolvedBuild, input.Comment); err != nil {
			return nil, err
		}
	case "close":
		if err := s.client.CloseBug(bugID, input.Comment); err != nil {
			return nil, err
		}
		// 部分禅道版本 REST close 返回 200 但实际未生效，校验后走 Web 通道兜底
		if !s.bugHasStatus(bugID, "closed") {
			if err := s.client.WebCloseBug(bugID, input.Comment); err != nil {
				return nil, err
			}
		}
	case "activate":
		restErr := s.client.ActivateBug(bugID, input.AssignedTo, input.Comment)
		// 老版本禅道 REST 无 activate 端点（404），或返回后状态未变 → Web 通道兜底
		if restErr != nil || !s.bugHasStatus(bugID, "active") {
			if err := s.client.WebActivateBug(bugID, input.AssignedTo, input.Comment); err != nil {
				if restErr != nil {
					return nil, fmt.Errorf("激活失败（REST: %v；Web 兜底: %w）", restErr, err)
				}
				return nil, err
			}
		}
	case "assign":
		if strings.TrimSpace(input.AssignedTo) == "" {
			return nil, validationError("指派 Bug 必须指定 assignedTo（禅道账号）")
		}
		if err := s.client.AssignBug(bugID, input.AssignedTo, input.Comment); err != nil {
			return nil, err
		}
	default:
		return nil, validationError("不支持的状态流转动作: %s（支持 %s）", input.Action, strings.Join(myzentao.BugValidTransitions, "/"))
	}

	return s.client.GetBug(bugID)
}

// bugHasStatus 查询 Bug 当前状态是否为期望值（查询失败视为不匹配，交由上层兜底/报错）
func (s *BugService) bugHasStatus(bugID int, want string) bool {
	bug, err := s.client.GetBug(bugID)
	if err != nil {
		return false
	}
	return bug.Status == want
}

func isValidResolution(resolution string) bool {
	for _, r := range myzentao.BugValidResolutions {
		if r == resolution {
			return true
		}
	}
	return false
}

// ParseID 解析路径参数中的资源 ID
func ParseID(raw string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("无效的资源 ID: %s", raw)
	}
	return id, nil
}

// GetBugActions 获取 Bug 动作历史（流转/处置记录，只读）——「今天流转多少」与处置
// 审计的承接面（扁鹊批次二百三十四）。走 Web 会话通道 api-getModel，错误原样透出。
func (s *BugService) GetBugActions(bugID int) ([]myzentao.BugAction, error) {
	if bugID <= 0 {
		return nil, validationError("无效的 Bug ID: %d", bugID)
	}
	return s.client.GetBugActions(bugID)
}
