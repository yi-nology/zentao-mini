package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	myzentao "chandao-mini/backend/core/zentao"
	"chandao-mini/backend/core/vo"
	"github.com/yi-nology/common/biz/zentao"
)

// DashboardService 仪表盘服务
type DashboardService struct {
	client *myzentao.Client
}

// NewDashboardService 创建仪表盘服务
func NewDashboardService(client *myzentao.Client) *DashboardService {
	return &DashboardService{client: client}
}

// GetDashboard 获取仪表盘数据
func (s *DashboardService) GetDashboard(productID int) (*vo.DashboardVO, error) {
	dashboard := &vo.DashboardVO{}

	bugs, _ := s.client.GetAllBugs(productID)
	if bugs != nil {
		dashboard.Bugs = calcBugStats(bugs)
		n := len(bugs)
		if n > 5 {
			n = 5
		}
		dashboard.RecentBugs = convertBugs(bugs[:n])
	}

	stories, _ := s.client.GetAllStories(productID)
	if stories != nil {
		dashboard.Stories = calcStoryStats(stories)
	}

	execCtxs, _ := s.client.GetExecutionsByProduct(productID)
	if execCtxs != nil {
		allTasks := collectTasks(s.client, execCtxs)
		dashboard.Tasks = calcTaskStats(allTasks)
		n := len(allTasks)
		if n > 5 {
			n = 5
		}
		dashboard.RecentTasks = convertTasks(allTasks[:n])
	}

	return dashboard, nil
}

// GetProjectOverview 获取项目概览
func (s *DashboardService) GetProjectOverview(projectID int) (*vo.ProjectOverviewVO, error) {
	overview := &vo.ProjectOverviewVO{}

	project, err := s.client.GetProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("获取项目信息失败: %w", err)
	}
	overview.Project = vo.ProjectInfoVO{
		ID:     project.ID,
		Name:   project.Name,
		Code:   project.Code,
		Status: project.Status,
		Begin:  project.Begin,
		End:    project.End,
	}

	bugs, _ := s.client.GetAllBugsByProject(projectID)
	if bugs != nil {
		overview.Bugs = calcBugStats(bugs)
	}

	tasks := collectTasksByProject(s.client, projectID)
	overview.Tasks = calcTaskStats(tasks)

	return overview, nil
}

// GetPersonalTimelog 获取个人工时报表
func (s *DashboardService) GetPersonalTimelog(account string, productID int, dateFrom string, dateTo string, groupBy string) (*vo.PersonalTimelogVO, error) {
	result := &vo.PersonalTimelogVO{}

	execCtxs, err := s.client.GetExecutionsByProduct(productID)
	if err != nil {
		return nil, err
	}

	dateMap := make(map[string]float64)
	projectMap := make(map[int]float64)
	projectNames := make(map[int]string)
	var details []vo.TimelogEntryVO

	for _, ec := range execCtxs {
		tasks, err := s.client.GetTasks(ec.Exec.ID, 1, 200)
		if err != nil || len(tasks) == 0 {
			continue
		}
		for _, t := range tasks {
			consumed := toFloat64(t.Consumed)
			if consumed <= 0 {
				continue
			}
			efforts, err := s.client.GetTaskEfforts(t.ID)
			if err != nil {
				continue
			}
			for _, e := range efforts {
				eAccount := e.Account
				if account != "" && eAccount != account {
					continue
				}
				if dateFrom != "" && e.Date < dateFrom {
					continue
				}
				if dateTo != "" && e.Date > dateTo {
					continue
				}

				c := toFloat64(e.Consumed)
				result.TotalHours += c

				dateKey := e.Date
				if groupBy == "week" {
					dateKey = toWeekKey(e.Date)
				} else if groupBy == "month" {
					if len(e.Date) >= 7 {
						dateKey = e.Date[:7]
					}
				}
				dateMap[dateKey] += c

				pid := toInt(e.Project)
				projectMap[pid] += c
				projectNames[pid] = ec.ProjName

				work := e.Work
				details = append(details, vo.TimelogEntryVO{
					ID:          e.ID,
					Work:        work,
					Date:        e.Date,
					Consumed:    c,
					ProjectID:   pid,
					ProjectName: projectNames[pid],
				})
			}
		}
	}

	for date, hours := range dateMap {
		result.ByDate = append(result.ByDate, vo.DateHoursVO{Date: date, Hours: hours})
	}
	for pid, hours := range projectMap {
		result.ByProject = append(result.ByProject, vo.ProjectHoursVO{
			ProjectID:   pid,
			ProjectName: projectNames[pid],
			Hours:       hours,
		})
	}
	result.Details = details
	return result, nil
}

// Search 全局搜索
func (s *DashboardService) Search(keyword string, productID int, page int, pageSize int) (*vo.SearchVO, error) {
	result := &vo.SearchVO{}
	var items []vo.SearchItem
	kw := strings.ToLower(keyword)

	if productID > 0 {
		// Bugs
		bugs, _ := s.client.GetAllBugs(productID)
		if bugs != nil {
			for _, b := range bugs {
				if strings.Contains(strings.ToLower(b.Title), kw) {
					items = append(items, vo.SearchItem{
						Type:   "bug",
						ID:     b.ID,
						Title:  b.Title,
						Status: b.Status,
						Extra:  map[string]interface{}{"severity": b.Severity, "pri": b.Pri},
					})
				}
			}
		}

		// Stories
		stories, _ := s.client.GetAllStories(productID)
		if stories != nil {
			for _, st := range stories {
				if strings.Contains(strings.ToLower(st.Title), kw) {
					items = append(items, vo.SearchItem{
						Type:   "story",
						ID:     st.ID,
						Title:  st.Title,
						Status: st.Status,
						Extra:  map[string]interface{}{"stage": st.Stage, "pri": st.Pri},
					})
				}
			}
		}

		// Tasks
		execCtxs, _ := s.client.GetExecutionsByProduct(productID)
		if execCtxs != nil {
			for _, ec := range execCtxs {
				tasks, err := s.client.GetTasks(ec.Exec.ID, 1, 200)
				if err != nil {
					continue
				}
				for _, t := range tasks {
					if strings.Contains(strings.ToLower(t.Name), kw) {
						items = append(items, vo.SearchItem{
							Type:   "task",
							ID:     t.ID,
							Title:  t.Name,
							Status: t.Status,
							Extra:  map[string]interface{}{"execution": ec.ExecName},
						})
					}
				}
			}
		}
	}

	result.Total = len(items)
	start := (page - 1) * pageSize
	if start > len(items) {
		start = len(items)
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	result.Items = items[start:end]
	return result, nil
}

// ========== 统计辅助 ==========

func calcBugStats(bugs []zentao.Bug) vo.BugStatsVO {
	stats := vo.BugStatsVO{
		BySeverity: make(map[string]int),
		Total:      len(bugs),
	}
	for _, b := range bugs {
		switch b.Status {
		case "active":
			stats.Active++
		case "resolved":
			stats.Resolved++
		case "closed":
			stats.Closed++
		}
		stats.BySeverity[fmt.Sprintf("%v", b.Severity)]++
	}
	return stats
}

func calcStoryStats(stories []zentao.Story) vo.StoryStatsVO {
	stats := vo.StoryStatsVO{
		ByStage: make(map[string]int),
		Total:   len(stories),
	}
	for _, st := range stories {
		switch st.Status {
		case "draft":
			stats.Draft++
		case "closed":
			stats.Closed++
		default:
			stats.Active++
		}
		stats.ByStage[fmt.Sprintf("%v", st.Stage)]++
	}
	return stats
}

func calcTaskStats(tasks []zentao.Task) vo.TaskStatsVO {
	stats := vo.TaskStatsVO{Total: len(tasks)}
	for _, t := range tasks {
		switch t.Status {
		case "wait":
			stats.Wait++
		case "doing":
			stats.Doing++
		case "done":
			stats.Done++
		case "closed":
			stats.Closed++
		}
	}
	return stats
}

func collectTasks(client *myzentao.Client, execCtxs []myzentao.ExecutionContext) []zentao.Task {
	var all []zentao.Task
	for _, ec := range execCtxs {
		tasks, err := client.GetTasks(ec.Exec.ID, 1, 200)
		if err == nil && tasks != nil {
			all = append(all, tasks...)
		}
	}
	return all
}

func collectTasksByProject(client *myzentao.Client, projectID int) []zentao.Task {
	executions, err := client.GetExecutions(projectID, 1, 200)
	if err != nil {
		return nil
	}
	var all []zentao.Task
	for _, e := range executions {
		tasks, err := client.GetTasks(e.ID, 1, 200)
		if err == nil && tasks != nil {
			all = append(all, tasks...)
		}
	}
	return all
}

func convertBugs(bugs []zentao.Bug) []vo.BugVO {
	result := make([]vo.BugVO, 0, len(bugs))
	for _, b := range bugs {
		result = append(result, vo.BugVO{
			ID:       b.ID,
			Project:  b.Project,
			Product:  b.Product,
			Title:    b.Title,
			Severity: b.Severity,
			Pri:      b.Pri,
			Type:     b.Type,
			Status:   b.Status,
		})
	}
	return result
}

func convertTasks(tasks []zentao.Task) []vo.TaskVO {
	result := make([]vo.TaskVO, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, vo.TaskVO{
			ID:        t.ID,
			Project:   t.Project,
			Execution: t.Execution,
			Name:      t.Name,
			Type:      t.Type,
			Pri:       t.Pri,
			Status:    t.Status,
		})
	}
	return result
}

func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	case string:
		f, _ := time.ParseDuration(val)
		return f.Hours()
	}
	return 0
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case json.Number:
		n, _ := val.Int64()
		return int(n)
	}
	return 0
}

func toWeekKey(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	year, week := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", year, week)
}
