package mcp

import (
	"fmt"
	"strconv"

	"github.com/yi-nology/zentao-mini/backend/core/service"
)

// Bug 写工具处理器：与 REST 同一套 service 校验，
// 权限由 MCP 自身的 token 鉴权 + 只读模式（IsWriteAction）控制。

func bugIDFromParams(params map[string]interface{}) (int, error) {
	v, ok := params["bugId"]
	if !ok {
		return 0, fmt.Errorf("缺少必填参数 bugId")
	}
	// JSON 数字反序列化为 float64，fmt %v 会输出科学计数法（1e+08），需先转整型
	if f, ok := v.(float64); ok {
		id := int64(f)
		if id <= 0 {
			return 0, fmt.Errorf("无效的 bugId: %v", v)
		}
		return int(id), nil
	}
	id, err := strconv.Atoi(fmt.Sprintf("%v", v))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("无效的 bugId: %v", v)
	}
	return id, nil
}

func stringFromParams(params map[string]interface{}, key string) string {
	if v, ok := params[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// handleAddBugComment 给 Bug 添加备注.
func (s *MCPServer) handleAddBugComment(params map[string]interface{}) (interface{}, error) {
	bugID, err := bugIDFromParams(params)
	if err != nil {
		return nil, err
	}
	comment := stringFromParams(params, "comment")

	bug, err := s.bugService.AddBugComment(bugID, comment)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status":  "ok",
		"message": "评论成功",
		"data":    bug,
	}, nil
}

// handleTransitionBug 执行 Bug 状态流转（confirm/resolve/close/activate/assign）.
func (s *MCPServer) handleTransitionBug(params map[string]interface{}) (interface{}, error) {
	bugID, err := bugIDFromParams(params)
	if err != nil {
		return nil, err
	}
	input := &service.BugTransitionInput{
		Action:        stringFromParams(params, "action"),
		Resolution:    stringFromParams(params, "resolution"),
		ResolvedBuild: stringFromParams(params, "resolvedBuild"),
		AssignedTo:    stringFromParams(params, "assignedTo"),
		Comment:       stringFromParams(params, "comment"),
	}

	bug, err := s.bugService.TransitionBug(bugID, input)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status":  "ok",
		"message": fmt.Sprintf("状态流转成功（%s）", input.Action),
		"data":    bug,
	}, nil
}
