package mcp

// Tool 定义 MCP 工具.
type Tool struct {
	Name        string           `json:"name"`
	Title       string           `json:"title,omitempty"` // 2025-06-18+ 人类可读名称
	Description string           `json:"description"`
	InputSchema InputSchema      `json:"inputSchema"`
	Annotations *ToolAnnotations `json:"annotations,omitempty"` // 2025-06-18+ 行为提示
}

// ToolAnnotations 工具行为提示（2025-06-18+）。zentao-mini 的 9 个工具全部只读。
type ToolAnnotations struct {
	ReadOnlyHint    *bool `json:"readOnlyHint,omitempty"`
	DestructiveHint *bool `json:"destructiveHint,omitempty"`
	IdempotentHint  *bool `json:"idempotentHint,omitempty"`
	OpenWorldHint   *bool `json:"openWorldHint,omitempty"`
}

// readOnlyAnnotations 只读工具的行为提示
func readOnlyAnnotations() *ToolAnnotations {
	t := true
	f := false
	return &ToolAnnotations{
		ReadOnlyHint:    &t,
		DestructiveHint: &f,
		IdempotentHint:  &t,
		OpenWorldHint:   &t,
	}
}

// writeAnnotations 写操作工具的行为提示（改动禅道数据，不可重放）
func writeAnnotations() *ToolAnnotations {
	t := true
	f := false
	return &ToolAnnotations{
		ReadOnlyHint:    &f,
		DestructiveHint: &f,
		IdempotentHint:  &f,
		OpenWorldHint:   &t,
	}
}

// InputSchema JSON Schema 格式的参数定义.
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// Property 参数属性.
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// Tools 所有可用的 MCP 工具.
var Tools = []Tool{
	{
		Name:        "ping",
		Title:       "连通性测试",
		Description: "测试 MCP 服务是否正常运行",
		InputSchema: InputSchema{
			Type:       "object",
			Properties: map[string]Property{},
		},
	},
	{
		Name:        "get_products",
		Title:       "产品列表",
		Description: "获取禅道产品列表",
		InputSchema: InputSchema{
			Type:       "object",
			Properties: map[string]Property{},
		},
	},
	{
		Name:        "get_projects",
		Title:       "项目列表",
		Description: "获取项目列表，可按产品 ID 过滤",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"productId": {
					Type:        "string",
					Description: "产品 ID（可选）",
				},
			},
		},
	},
	{
		Name:        "get_executions",
		Title:       "执行/迭代列表",
		Description: "获取执行/迭代列表，可按项目 ID 或产品 ID 过滤",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"projectId": {
					Type:        "string",
					Description: "项目 ID（可选）",
				},
				"productId": {
					Type:        "string",
					Description: "产品 ID（可选）",
				},
			},
		},
	},
	{
		Name:        "get_bugs",
		Title:       "Bug 列表",
		Description: "获取 Bug 列表，可按产品 ID 和状态过滤",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"productId": {
					Type:        "string",
					Description: "产品 ID（可选）",
				},
				"status": {
					Type:        "string",
					Description: "Bug 状态（可选，如 active, resolved, closed）",
				},
			},
		},
	},
	{
		Name:        "get_stories",
		Title:       "需求列表",
		Description: "获取需求列表，可按产品 ID 过滤",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"productId": {
					Type:        "string",
					Description: "产品 ID（可选）",
				},
			},
		},
	},
	{
		Name:        "get_tasks",
		Title:       "任务列表",
		Description: "获取任务列表，可按产品 ID 和执行 ID 过滤",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"productId": {
					Type:        "string",
					Description: "产品 ID（可选）",
				},
				"executionId": {
					Type:        "string",
					Description: "执行/迭代 ID（可选）",
				},
			},
		},
	},
	{
		Name:        "get_users",
		Title:       "用户列表",
		Description: "获取用户列表",
		InputSchema: InputSchema{
			Type:       "object",
			Properties: map[string]Property{},
		},
	},
	{
		Name:        "get_timelog",
		Title:       "工时统计",
		Description: "获取工时统计数据，可按产品 ID 和日期范围过滤",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"productId": {
					Type:        "string",
					Description: "产品 ID（可选）",
				},
				"dateFrom": {
					Type:        "string",
					Description: "开始日期，格式 YYYY-MM-DD（可选）",
				},
				"dateTo": {
					Type:        "string",
					Description: "结束日期，格式 YYYY-MM-DD（可选）",
				},
			},
		},
	},
	{
		Name:        "add_bug_comment",
		Title:       "Bug 评论",
		Description: "给指定 Bug 添加备注/评论（写操作，以禅道配置账号名义发表）",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"bugId": {
					Type:        "string",
					Description: "Bug ID（必填）",
				},
				"comment": {
					Type:        "string",
					Description: "评论内容（必填，支持禅道富文本 HTML）",
				},
			},
			Required: []string{"bugId", "comment"},
		},
	},
	{
		Name:        "transition_bug",
		Title:       "Bug 状态流转",
		Description: "执行 Bug 状态流转（写操作）：confirm=确认、resolve=解决（需 resolution）、close=关闭、activate=激活（重新打开）、assign=指派（需 assignedTo）。所有动作均可附 comment 记入操作历史",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"bugId": {
					Type:        "string",
					Description: "Bug ID（必填）",
				},
				"action": {
					Type:        "string",
					Description: "流转动作（必填）",
					Enum:        []string{"confirm", "resolve", "close", "activate", "assign"},
				},
				"resolution": {
					Type:        "string",
					Description: "解决方式，resolve 必填",
					Enum:        []string{"fixed", "bydesign", "duplicate", "notrepro", "postponed", "willnotfix", "external"},
				},
				"resolvedBuild": {
					Type:        "string",
					Description: "解决版本（resolve 时建议填写，如 trunk 或版本号）",
				},
				"assignedTo": {
					Type:        "string",
					Description: "禅道账号，assign 必填；activate 可选（重新指派）",
				},
				"comment": {
					Type:        "string",
					Description: "随流转记录的备注（可选）",
				},
			},
			Required: []string{"bugId", "action"},
		},
	},
}

// GetToolByName 根据名称获取工具定义.
func GetToolByName(name string) *Tool {
	for _, t := range Tools {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

// 写操作工具名单（annotations 标记 readOnlyHint=false；只读模式据此拦截）
var writeTools = map[string]bool{
	"add_bug_comment": true,
	"transition_bug":  true,
}

// annotations：只读工具统一 readOnly，写工具单独标记；已显式设置的不再覆盖
func init() {
	for i := range Tools {
		if Tools[i].Annotations != nil {
			continue
		}
		if writeTools[Tools[i].Name] {
			Tools[i].Annotations = writeAnnotations()
		} else {
			Tools[i].Annotations = readOnlyAnnotations()
		}
	}
}
