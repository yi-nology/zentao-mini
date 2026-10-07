package zentao

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yi-nology/common/biz/zentao"
)

// GetBugs 获取产品的 Bug 列表
// 旧版禅道（自建/开源老版本）会把 assignedTo/openedBy 返回为纯字符串账号，
// SDK 的 UserRef 结构解析会直接报 "cannot unmarshal string" 使整个列表失败，
// 此时降级到本地容错解析（字符串包装成 UserRef 对象）。
func (c *Client) GetBugs(productID int, page, pageSize int) ([]zentao.Bug, error) {
	cacheKey := DefaultKeyBuilder.Build("zentao:bugs", strconv.Itoa(productID), strconv.Itoa(page), strconv.Itoa(pageSize))

	result, err := GlobalCache.GetOrLoadWithLock(cacheKey, func() (interface{}, error) {
		var response *zentao.BugListResponse
		err := c.withTokenRetry("GetBugs", func(client *zentao.Client) error {
			var err error
			response, err = client.GetBugs(productID, page, pageSize)
			if err != nil && isUnmarshalTypeError(err) {
				var fallback *zentao.BugListResponse
				fallback, err = c.getBugsTolerant(productID, page, pageSize)
				if err == nil {
					response = fallback
				}
			}
			return err
		})
		if err != nil {
			return nil, err
		}
		return response.Bugs, nil
	}, 2*time.Minute)

	if err != nil {
		return nil, err
	}
	return result.([]zentao.Bug), nil
}

func isUnmarshalTypeError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "cannot unmarshal") || strings.Contains(msg, "解析响应失败")
}

// getBugsTolerant 绕过 SDK 的严格 UserRef 解析：
// 原始响应里 assignedTo/openedBy 为字符串时包装成 {"account": <字符串>} 再解到 zentao.BugListResponse。
func (c *Client) getBugsTolerant(productID, page, pageSize int) (*zentao.BugListResponse, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}
	c.mu.RLock()
	server := c.server
	c.mu.RUnlock()
	if server == "" {
		return nil, fmt.Errorf("禅道服务器地址为空")
	}

	httpClient := &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	url := fmt.Sprintf("%s/api.php/v1/products/%d/bugs?page=%d&limit=%d", server, productID, page, pageSize)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Token", token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求禅道 bug 列表失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败, 状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result zentao.BugListResponse
	if err := json.Unmarshal(normalizeBugUserRefs(body), &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}
	return &result, nil
}

// normalizeBugUserRefs 把 bugs 数组中字符串形式的 assignedTo/openedBy 规范成对象形式，
// 兼容旧版禅道的返回格式；响应结构不符时原样返回。
func normalizeBugUserRefs(body []byte) []byte {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(body, &doc); err != nil {
		return body
	}
	rawBugs, ok := doc["bugs"]
	if !ok {
		return body
	}
	var bugs []map[string]json.RawMessage
	if err := json.Unmarshal(rawBugs, &bugs); err != nil {
		return body
	}
	changed := false
	for i, bug := range bugs {
		for _, field := range []string{"assignedTo", "openedBy"} {
			raw, exists := bug[field]
			if !exists {
				continue
			}
			var account string
			if err := json.Unmarshal(raw, &account); err == nil {
				bug[field], _ = json.Marshal(map[string]string{"account": account})
				changed = true
			}
		}
		bugs[i] = bug
	}
	if !changed {
		return body
	}
	doc["bugs"], _ = json.Marshal(bugs)
	out, err := json.Marshal(doc)
	if err != nil {
		return body
	}
	return out
}

// GetBugsByProject 根据项目 ID 过滤 Bug 列表
// 复用 GetBugs（含旧版禅道字符串 assignedTo 容错），在内存中过滤，与 SDK 语义一致
func (c *Client) GetBugsByProject(productID, projectID int, page, pageSize int) ([]zentao.Bug, error) {
	bugs, err := c.GetBugs(productID, page, pageSize)
	if err != nil {
		return nil, err
	}
	filtered := make([]zentao.Bug, 0)
	for _, bug := range bugs {
		if bug.Project == projectID {
			filtered = append(filtered, bug)
		}
	}
	return filtered, nil
}

// GetBugsByStatus 根据状态过滤 Bug 列表
func (c *Client) GetBugsByStatus(productID int, status string, page, pageSize int) ([]zentao.Bug, error) {
	bugs, err := c.GetBugs(productID, page, pageSize)
	if err != nil {
		return nil, err
	}
	filtered := make([]zentao.Bug, 0)
	for _, bug := range bugs {
		if bug.Status == status {
			filtered = append(filtered, bug)
		}
	}
	return filtered, nil
}

// SearchBugs 搜索 Bug（支持多条件过滤）
func (c *Client) SearchBugs(params zentao.BugSearchParams) ([]zentao.Bug, error) {
	var response *zentao.BugListResponse
	err := c.withTokenRetry("SearchBugs", func(client *zentao.Client) error {
		var err error
		response, err = client.SearchBugs(params)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Bugs, nil
}

// GetBug 获取 Bug 详情
func (c *Client) GetBug(bugID int) (*zentao.Bug, error) {
	var result *zentao.Bug
	err := c.withTokenRetry("GetBug", func(client *zentao.Client) error {
		var err error
		result, err = client.GetBug(bugID)
		return err
	})
	return result, err
}

// GetAllBugs 获取产品全部 Bug（自动翻页）
func (c *Client) GetAllBugs(productID int) ([]zentao.Bug, error) {
	var all []zentao.Bug
	page := 1
	for {
		bugs, err := c.GetBugs(productID, page, 100)
		if err != nil {
			return all, err
		}
		all = append(all, bugs...)
		if len(bugs) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// GetAllBugsIncludeClosed 获取产品全部 Bug（含 closed 状态）
// 禅道默认 /products/{id}/bugs 接口不返回 closed bug，
// 需显式传 status=all 才能获取全部状态。自动翻页。
func (c *Client) GetAllBugsIncludeClosed(productID int) ([]zentao.Bug, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, fmt.Errorf("获取 token 失败: %w", err)
	}
	server := c.GetServer()
	if server == "" {
		return nil, fmt.Errorf("禅道服务器地址为空")
	}

	// 禅道可能使用自签名证书，跳过校验（与上游 SDK doGet 行为一致）
	httpClient := &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	var all []zentao.Bug
	page := 1
	const pageSize = 200
	for {
		url := fmt.Sprintf("%s/api.php/v1/products/%d/bugs?page=%d&limit=%d&status=all",
			server, productID, page, pageSize)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return all, err
		}
		req.Header.Set("Token", token)

		resp, err := httpClient.Do(req)
		if err != nil {
			return all, fmt.Errorf("请求禅道 bug 列表失败: %w", err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return all, fmt.Errorf("获取 bug 列表失败, 状态码: %d, 响应: %s", resp.StatusCode, string(body))
		}

		var result zentao.BugListResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return all, fmt.Errorf("解析 bug 列表失败: %w", err)
		}
		all = append(all, result.Bugs...)
		if len(result.Bugs) < pageSize {
			break
		}
		page++
	}
	return all, nil
}

// GetAllBugsByProject 获取项目全部 Bug
func (c *Client) GetAllBugsByProject(projectID int) ([]zentao.Bug, error) {
	return c.GetAllBugsByProjectWithProduct(0, projectID)
}

// GetAllBugsByProjectWithProduct 获取项目全部 Bug（指定产品ID）
func (c *Client) GetAllBugsByProjectWithProduct(productID int, projectID int) ([]zentao.Bug, error) {
	var allBugs []zentao.Bug
	page := 1
	for {
		bugs, err := c.GetBugs(productID, page, 500)
		if err != nil {
			return allBugs, err
		}
		allBugs = append(allBugs, bugs...)
		if len(bugs) < 500 {
			break
		}
		page++
	}

	if projectID <= 0 {
		return allBugs, nil
	}
	filtered := make([]zentao.Bug, 0, len(allBugs))
	for _, bug := range allBugs {
		if bug.Project == projectID {
			filtered = append(filtered, bug)
		}
	}
	return filtered, nil
}

// GetBugsContext 获取 Bug 列表（支持 context 取消）
func (c *Client) GetBugsContext(ctx context.Context, productID int, page, pageSize int) ([]zentao.Bug, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var response *zentao.BugListResponse
	err := c.withTokenRetryContext(ctx, "GetBugs", func(client *zentao.Client) error {
		var err error
		response, err = client.GetBugs(productID, page, pageSize)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response.Bugs, nil
}

// GetAllBugsContext 获取产品全部 Bug（支持 context 取消，自动翻页）
func (c *Client) GetAllBugsContext(ctx context.Context, productID int) ([]zentao.Bug, error) {
	var all []zentao.Bug
	page := 1
	for {
		select {
		case <-ctx.Done():
			return all, ctx.Err()
		default:
		}
		bugs, err := c.GetBugsContext(ctx, productID, page, 100)
		if err != nil {
			return all, err
		}
		all = append(all, bugs...)
		if len(bugs) < 100 {
			break
		}
		page++
	}
	return all, nil
}

// GetAllBugsByProjectContext 获取项目全部 Bug（支持 context 取消）
func (c *Client) GetAllBugsByProjectContext(ctx context.Context, projectID int) ([]zentao.Bug, error) {
	var all []zentao.Bug
	page := 1
	for {
		select {
		case <-ctx.Done():
			return all, ctx.Err()
		default:
		}
		var response *zentao.BugListResponse
		err := c.withTokenRetryContext(ctx, "GetBugsByProject", func(client *zentao.Client) error {
			var err error
			response, err = client.GetBugsByProject(0, projectID, page, 100)
			return err
		})
		if err != nil {
			return all, err
		}
		all = append(all, response.Bugs...)
		if len(response.Bugs) < 100 {
			break
		}
		page++
	}
	return all, nil
}
