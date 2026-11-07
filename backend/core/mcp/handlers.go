package mcp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yi-nology/zentao-mini/backend/core/dto"
)

// handlePing 处理 ping 请求.
func (s *MCPServer) handlePing(_ map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"status":  "ok",
		"message": "Pong",
		"version": "1.0",
	}, nil
}

// handleGetProducts 获取产品列表.
func (s *MCPServer) handleGetProducts(_ map[string]interface{}) (interface{}, error) {
	result, err := s.productService.GetProducts()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status":  "ok",
		"message": "Products retrieved successfully",
		"data":    result,
	}, nil
}

// intFromParams 取整型参数：JSON 数字反序列化为 float64，fmt %v 会输出
// 科学计数法（1000000 → "1e+06"，Atoi 直接失败）——数字先走整型转换，
// 字符串再走 Atoi。 handlers_write.go 的 bugIDFromParams 是同一形状的
// 必填版；这里是可选版（缺参/非法返回 0 与 false，调用方保持零值）。
//
// string 参数一律走 handlers_write.go 的 stringFromParams（issue #20 第 4 项）：
// DTO 零值即 ""，缺参直接赋零值等价于旧的「存在才赋值」；TrimSpace 语义在
// 调用点显式保留（status/productId 等不 trim，日期窗/assignedTo trim）。
func optionalIntFromParams(params map[string]interface{}, key string) (int, bool) {
	v, ok := params[key]
	if !ok {
		return 0, false
	}
	if f, ok := v.(float64); ok {
		return int(f), true
	}
	if n, err := strconv.Atoi(fmt.Sprintf("%v", v)); err == nil {
		return n, true
	}
	return 0, false
}

// optionalIntRangeFromParams 同上，附带值域钳制（超出范围视同未提供）。
func optionalIntRangeFromParams(params map[string]interface{}, key string, lo, hi int) (int, bool) {
	n, ok := optionalIntFromParams(params, key)
	if !ok || n < lo || n > hi {
		return 0, false
	}
	return n, true
}

// handleGetProjects 获取项目列表.
func (s *MCPServer) handleGetProjects(params map[string]interface{}) (interface{}, error) {
	query := &dto.ProjectQueryDTO{}
	if n, ok := optionalIntFromParams(params, "productId"); ok {
		query.ProductID = n
	}

	result, err := s.projectService.GetProjects(query)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status":  "ok",
		"message": "Projects retrieved successfully",
		"data":    result,
	}, nil
}

// handleGetExecutions 获取执行/迭代列表.
func (s *MCPServer) handleGetExecutions(params map[string]interface{}) (interface{}, error) {
	query := &dto.ExecutionQueryDTO{}
	if n, ok := optionalIntFromParams(params, "projectId"); ok {
		query.ProjectID = n
	}
	if n, ok := optionalIntFromParams(params, "productId"); ok {
		query.ProductID = n
	}

	result, err := s.executionService.GetExecutions(query)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status":  "ok",
		"message": "Executions retrieved successfully",
		"data":    result,
	}, nil
}

// handleGetBugs 获取 Bug 列表.
// MCP 面缺省 Lite=true（Steps 剥标签截断）——完整重现步骤走 get_bug 按 ID 取；
// 分页/过滤齐备（assignedTo=账号、severity=1-5），Validate 统一钳制分页上界。
func (s *MCPServer) handleGetBugs(params map[string]interface{}) (interface{}, error) {
	query := &dto.BugQueryDTO{Lite: true}
	if n, ok := optionalIntFromParams(params, "productId"); ok {
		query.ProductID = n
	}
	query.Status = stringFromParams(params, "status")
	query.AssignedTo = strings.TrimSpace(stringFromParams(params, "assignedTo"))
	if n, ok := optionalIntRangeFromParams(params, "severity", 1, 5); ok {
		query.Severity = n
	}
	if n, ok := optionalIntFromParams(params, "page"); ok {
		query.Page = n
	}
	if n, ok := optionalIntFromParams(params, "pageSize"); ok {
		query.PageSize = n
	}
	// 日期窗（扁鹊批次二百三十四）：startDate/endDate 作用于 openedDate（创建），
	// resolvedStartDate/resolvedEndDate 作用于 resolvedDate（解决）——「今天新增多少/
	// 解决多少」的承接面；specificDate 与 startDate 同域（单日快写）。YYYY-MM-DD。
	query.StartDate = strings.TrimSpace(stringFromParams(params, "startDate"))
	query.EndDate = strings.TrimSpace(stringFromParams(params, "endDate"))
	query.SpecificDate = strings.TrimSpace(stringFromParams(params, "specificDate"))
	query.ResolvedStartDate = strings.TrimSpace(stringFromParams(params, "resolvedStartDate"))
	query.ResolvedEndDate = strings.TrimSpace(stringFromParams(params, "resolvedEndDate"))
	query.ClosedStartDate = strings.TrimSpace(stringFromParams(params, "closedStartDate"))
	query.ClosedEndDate = strings.TrimSpace(stringFromParams(params, "closedEndDate"))
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 100
	}
	if err := query.Validate(); err != nil {
		return nil, err
	}

	result, err := s.bugService.GetBugs(query)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status":  "ok",
		"message": "Bugs retrieved successfully",
		"data":    result,
	}, nil
}

// handleGetStories 获取需求列表.
func (s *MCPServer) handleGetStories(params map[string]interface{}) (interface{}, error) {
	query := &dto.StoryQueryDTO{}
	if n, ok := optionalIntFromParams(params, "productId"); ok {
		query.ProductID = n
	}
	query.Page = 1
	query.PageSize = 100

	result, err := s.storyService.GetStories(query)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status":  "ok",
		"message": "Stories retrieved successfully",
		"data":    result,
	}, nil
}

// handleGetTasks 获取任务列表.
func (s *MCPServer) handleGetTasks(params map[string]interface{}) (interface{}, error) {
	query := &dto.TaskQueryDTO{}
	if n, ok := optionalIntFromParams(params, "productId"); ok {
		query.ProductID = n
	}
	if n, ok := optionalIntFromParams(params, "executionId"); ok {
		query.ExecutionID = n
	}
	query.Page = 1
	query.PageSize = 100

	result, err := s.taskService.GetTasks(query)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status":  "ok",
		"message": "Tasks retrieved successfully",
		"data":    result,
	}, nil
}

// handleGetUsers 获取用户列表.
func (s *MCPServer) handleGetUsers(_ map[string]interface{}) (interface{}, error) {
	result, err := s.userService.GetUsers(1, 100)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status":  "ok",
		"message": "Users retrieved successfully",
		"data":    result,
	}, nil
}

// handleGetTimelog 获取工时数据.
func (s *MCPServer) handleGetTimelog(params map[string]interface{}) (interface{}, error) {
	query := &dto.TimelogQueryDTO{}
	query.ProductID = stringFromParams(params, "productId")
	query.DateFrom = stringFromParams(params, "dateFrom")
	query.DateTo = stringFromParams(params, "dateTo")

	result, err := s.timelogService.GetTimelogDashboard(query)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status":  "ok",
		"message": "Timelog retrieved successfully",
		"data":    result,
	}, nil
}
