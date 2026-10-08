package zentao

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// BugAction 禅道对象动作历史（zt_action 行的只读投影）——流转/处置审计与「今天流转
// 多少」统计的承接面（扁鹊批次二百三十四挂账清偿：get_bugs 只有 created/resolved 日期，
// 流转次数此前无任何查询通道）。
type BugAction struct {
	ID      int    `json:"id"`
	Actor   string `json:"actor"`
	Action  string `json:"action"` // opened/confirmed/resolved/closed/activated/assigned/commented/edited...
	Date    string `json:"date"`   // YYYY-MM-DD HH:MM:SS
	Comment string `json:"comment"`
	Extra   string `json:"extra,omitempty"`
}

// GetBugActions 获取 Bug 的动作历史（流转记录）。走 Web 会话通道调 api-getModel
// （action-getObjectActions）——REST API v1 无 actions 端点，api-getModel 是版本兼容面
// 最宽的官方通道（禅道 12+ 会话鉴权恒可用；Web 会话复用 client_bug_comment 同一套
// 登录态）。PATH_INFO 风格为主、GET 风格兜底（requestType 配置差异，与评论通道同口径）。
// 动作历史只读不写，错误原样透出（含禅道侧校验/鉴权错误）由调用方转告。
func (c *Client) GetBugActions(bugID int) ([]BugAction, error) {
	actions, err := c.doGetBugActions(bugID)
	if err != nil {
		// 会话可能过期：重登一次再试（与 AddBugComment 同纪律）
		c.ResetWebSession()
		return c.doGetBugActions(bugID)
	}
	return actions, nil
}

func (c *Client) doGetBugActions(bugID int) ([]BugAction, error) {
	c.mu.RLock()
	server := c.server
	c.mu.RUnlock()
	if server == "" {
		return nil, fmt.Errorf("禅道服务器地址为空")
	}

	s, err := c.getWebSession()
	if err != nil {
		return nil, err
	}

	// api-getModel 参数以 '-' 连接的 key=value 对；PATH_INFO 为主、GET 风格兜底。
	targets := []string{
		fmt.Sprintf("%s/api-getModel-action-getObjectActions-objectTable=bug-objectID=%d.json", server, bugID),
		fmt.Sprintf("%s/index.php?m=api&f=getModel&moduleName=action&methodName=getObjectActions&objectTable=bug&objectID=%d&t=json", server, bugID),
	}

	var lastStatus int
	var lastBody string
	for _, target := range targets {
		req, err := http.NewRequest(http.MethodGet, target, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("Referer", fmt.Sprintf("%s/bug-view-%d.html", server, bugID))

		resp, err := c.webHTTPClient(s).Do(req)
		if err != nil {
			return nil, fmt.Errorf("获取动作历史请求失败: %w", err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		lastStatus, lastBody = resp.StatusCode, string(body)

		if resp.StatusCode != http.StatusOK {
			continue // 形态不可用，试下一目标
		}
		actions, parseErr := parseBugActionsResponse(string(body))
		if parseErr != nil {
			if strings.Contains(parseErr.Error(), "未登录") {
				continue // 会话失效，试 GET 形态/由上层重登
			}
			return nil, parseErr
		}
		return actions, nil
	}
	return nil, fmt.Errorf("获取动作历史失败（两形态均不可用）: 状态码 %d, 响应: %.200s", lastStatus, lastBody)
}

// parseBugActionsResponse 容忍三种返回形态：
//   - api-getModel 成功：{"status":"success","data":{"3":{...},"7":{...}}} 或 data 为数组；
//   - 空历史：data 为 [] 或 null；
//   - 失败：{"status":"failed"/"error","message":...}，或返回登录页 HTML（会话失效）。
func parseBugActionsResponse(body string) ([]BugAction, error) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" || trimmed[0] != '{' {
		return nil, fmt.Errorf("禅道返回非 JSON（可能未登录或 requestType 不支持）: %.120s", trimmed)
	}
	var envelope struct {
		Status  string          `json:"status"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
		// 部分版本 api-getModel 直接把方法结果放顶层（无 data 包裹）
		ID      int             `json:"id"`
		Actions json.RawMessage `json:"actions"`
	}
	if err := json.Unmarshal([]byte(trimmed), &envelope); err != nil {
		return nil, fmt.Errorf("解析动作历史响应失败: %w", err)
	}
	if envelope.Status != "" && envelope.Status != "success" && envelope.Status != "have data" {
		return nil, fmt.Errorf("禅道侧返回错误: %s", strings.TrimSpace(envelope.Message))
	}

	raw := envelope.Data
	if len(raw) == 0 || string(raw) == "null" {
		raw = envelope.Actions
	}
	if len(raw) == 0 || string(raw) == "null" {
		return []BugAction{}, nil
	}

	// data 形态一：对象（id→action 的关联数组，禅道惯例）
	byID := map[string]BugAction{}
	if err := json.Unmarshal(raw, &byID); err == nil && len(byID) > 0 {
		// 误判防护：空对象 {} 与真实记录都走这里，逐键校验 action 字段存在性
		actions := make([]BugAction, 0, len(byID))
		for _, a := range byID {
			if a.Action == "" && a.Actor == "" && a.ID == 0 {
				continue
			}
			actions = append(actions, a)
		}
		if len(actions) > 0 {
			return actions, nil
		}
	}
	// data 形态二：数组
	var list []BugAction
	if err := json.Unmarshal(raw, &list); err == nil {
		return list, nil
	}
	return []BugAction{}, nil
}
