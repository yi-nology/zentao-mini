package zentao

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/yi-nology/common/biz/zentao"
)

// ---- Test Cases (session mode) ----
// 端点：/testcase-browse-{productID}.json（第一页）/ 带分页段翻页。
// 注意：case 的 id 在麒麟实例上是字符串 "case_2343233"，需特殊解析。
// 内层 cases 是 list（不是 map），pager.recTotal/pageTotal 用于翻页。

type sessionTestCaseBrowseResp struct {
	Title string          `json:"title"`
	Cases []sessionCase   `json:"cases"`
	Pager *sessionPager   `json:"pager"`
}

type sessionCase struct {
	// ID 在麒麟实例上是字符串 "case_2343233"，在标准实例上是 int。
	// 用 RawMessage 兼容两者，由 parseCaseID 提取数字。
	ID           json.RawMessage `json:"id"`
	Product      int             `json:"product"`
	Project      int             `json:"project"`
	Module       json.Number     `json:"module"`
	Story        json.Number     `json:"story"`
	Title        string          `json:"title"`
	Precondition string          `json:"precondition"`
	Keywords     string          `json:"keywords"`
	Pri          json.Number     `json:"pri"`
	Type         string          `json:"type"`
	Status       string          `json:"status"`
	OpenedBy     string          `json:"openedBy"`
	OpenedDate   string          `json:"openedDate"`
	LastEditedBy string          `json:"lastEditedBy"`
	Version      json.Number     `json:"version"`
	Deleted      string          `json:"deleted"`
}

func (c sessionCase) toSDK() zentao.Case {
	id := parseCaseID(c.ID)
	module, _ := c.Module.Int64()
	story, _ := c.Story.Int64()
	pri, _ := c.Pri.Int64()
	ver, _ := c.Version.Int64()
	return zentao.Case{
		ID:           id,
		Product:      c.Product,
		Project:      c.Project,
		Module:       int(module),
		Story:        int(story),
		Title:        c.Title,
		Precondition: c.Precondition,
		Keywords:     c.Keywords,
		Pri:          int(pri),
		Type:         c.Type,
		Status:       c.Status,
		OpenedBy:     zentao.UserRef{Account: c.OpenedBy},
		OpenedDate:   c.OpenedDate,
		LastEditedBy: &zentao.UserRef{Account: c.LastEditedBy},
		Version:      int(ver),
	}
}

// parseCaseID 从 case id 的原始 JSON（字符串或数字）提取数字部分。
// 麒麟实例返回 "case_2343233"，标准实例返回 int。
func parseCaseID(raw json.RawMessage) int {
	s := string(raw)
	// 去掉 JSON 字符串的引号
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	digits := ""
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digits += string(r)
		}
	}
	if digits == "" {
		return 0
	}
	id, _ := strconv.Atoi(digits)
	return id
}

func (c *Client) getCasesByProductSession(ctx context.Context, productID, page, pageSize int) ([]zentao.Case, error) {
	// testcase-browse 分页段格式：/testcase-browse-{pid}-{branch}-{browseType}-{recTotal}-{recPerPage}-{page}.json
	// 第一页用裸路径拿 recTotal，后续页带完整段。
	var out []zentao.Case
	recTotal := 0
	const recPerPage = 100
	for p := page; ; p++ {
		var path string
		if recTotal == 0 {
			path = fmt.Sprintf("/testcase-browse-%d.json", productID)
		} else {
			path = fmt.Sprintf("/testcase-browse-%d-0-all-%d-%d-%d.json", productID, recTotal, recPerPage, p)
		}
		var resp sessionTestCaseBrowseResp
		if err := c.doSessionJSON(ctx, "GetCasesByProduct", path, &resp); err != nil {
			if IsSessionAccessDenied(err) {
				return nil, nil
			}
			return out, err
		}
		if recTotal == 0 && resp.Pager != nil {
			recTotal = pagerInt(resp.Pager.RecTotal)
		}
		for _, cs := range resp.Cases {
			out = append(out, cs.toSDK())
		}
		// 只取请求的单页（page/pageSize 调用方做内存分页），翻页仅用于 GetAll。
		if pageSize > 0 {
			return out, nil
		}
		if len(resp.Cases) < recPerPage {
			return out, nil
		}
		if resp.Pager != nil && p >= pagerInt(resp.Pager.PageTotal) {
			return out, nil
		}
		if p > 500 {
			return out, nil
		}
	}
}

func (c *Client) getAllCasesSession(ctx context.Context, productID int) ([]zentao.Case, error) {
	return c.getCasesByProductSession(ctx, productID, 1, 0) // pageSize=0 表示翻页取全部
}

// ---- Test Tasks (session mode) ----
// 端点：/testtask-browse-{productID}.json，内层 tasks 是 map[id]testtask。

type sessionTestTaskBrowseResp struct {
	Title  string                       `json:"title"`
	Tasks  map[string]sessionTestTask   `json:"tasks"`
	Pager  *sessionPager                `json:"pager"`
}

type sessionTestTask struct {
	ID        int    `json:"id"`
	Project   int    `json:"project"`
	Product   int    `json:"product"`
	Name      string `json:"name"`
	Execution int    `json:"execution"`
	Build     json.Number `json:"build"`
	Type      string `json:"type"`
	Owner     string `json:"owner"`
	Pri       int    `json:"pri"`
	Begin     string `json:"begin"`
	End       string `json:"end"`
	Status    string `json:"status"`
	Deleted   string `json:"deleted"`
}

func (t sessionTestTask) toSDK() zentao.TestTask {
	build, _ := t.Build.Int64()
	return zentao.TestTask{
		ID:        t.ID,
		Project:   t.Project,
		Product:   t.Product,
		Name:      t.Name,
		Execution: t.Execution,
		Build:     int(build),
		Type:      t.Type,
		Owner:     zentao.UserRef{Account: t.Owner},
		Pri:       t.Pri,
		Begin:     t.Begin,
		End:       t.End,
		Status:    t.Status,
		Deleted:   t.Deleted,
	}
}

func (c *Client) getTestTasksSession(ctx context.Context, productID, page, pageSize int) ([]zentao.TestTask, error) {
	path := fmt.Sprintf("/testtask-browse-%d.json", productID)
	var resp sessionTestTaskBrowseResp
	if err := c.doSessionJSON(ctx, "GetTestTasks", path, &resp); err != nil {
		if IsSessionAccessDenied(err) {
			return nil, nil
		}
		return nil, err
	}
	all := make([]zentao.TestTask, 0, len(resp.Tasks))
	for _, t := range resp.Tasks {
		all = append(all, t.toSDK())
	}
	if pageSize <= 0 {
		return all, nil
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
