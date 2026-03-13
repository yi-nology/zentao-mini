package service

import (
	"errors"
	"testing"

	"github.com/yi-nology/common/biz/zentao"

	"chandao-mini/backend/core/dto"
	myzentao "chandao-mini/backend/core/zentao"
)

// MockClient 模拟禅道客户端
type MockClient struct {
	GetBugsFunc          func(productID int, limit int) ([]zentao.Bug, error)
	GetBugsByProjectFunc func(productID, projectID int, limit int) ([]zentao.Bug, error)
	SearchBugsFunc       func(params zentao.BugSearchParams) ([]zentao.Bug, error)
}

// GetBugs 模拟获取Bug列表
func (m *MockClient) GetBugs(productID int, limit int) ([]zentao.Bug, error) {
	if m.GetBugsFunc != nil {
		return m.GetBugsFunc(productID, limit)
	}
	return []zentao.Bug{}, nil
}

// GetBugsByProject 模拟按项目获取Bug
func (m *MockClient) GetBugsByProject(productID, projectID int, limit int) ([]zentao.Bug, error) {
	if m.GetBugsByProjectFunc != nil {
		return m.GetBugsByProjectFunc(productID, projectID, limit)
	}
	return []zentao.Bug{}, nil
}

// SearchBugs 模拟搜索Bug
func (m *MockClient) SearchBugs(params zentao.BugSearchParams) ([]zentao.Bug, error) {
	if m.SearchBugsFunc != nil {
		return m.SearchBugsFunc(params)
	}
	return []zentao.Bug{}, nil
}

// createMockBugs 创建测试用的Bug数据
func createMockBugs() []zentao.Bug {
	return []zentao.Bug{
		{
			ID:         1,
			Project:    10,
			Product:    100,
			Title:      "Bug 1",
			Status:     "active",
			OpenedDate: "2024-01-15 10:00:00",
			AssignedTo: zentao.UserRef{Account: "user1", Realname: "User 1"},
			OpenedBy:   zentao.UserRef{Account: "user2", Realname: "User 2"},
		},
		{
			ID:         2,
			Project:    10,
			Product:    100,
			Title:      "Bug 2",
			Status:     "resolved",
			OpenedDate: "2024-01-16 11:00:00",
			AssignedTo: zentao.UserRef{Account: "user1", Realname: "User 1"},
			OpenedBy:   zentao.UserRef{Account: "user3", Realname: "User 3"},
		},
		{
			ID:         3,
			Project:    11,
			Product:    100,
			Title:      "Bug 3",
			Status:     "active",
			OpenedDate: "2024-01-17 12:00:00",
			AssignedTo: zentao.UserRef{Account: "user2", Realname: "User 2"},
			OpenedBy:   zentao.UserRef{Account: "user1", Realname: "User 1"},
		},
	}
}

// TestBugService_GetBugs 测试获取Bug列表
func TestBugService_GetBugs(t *testing.T) {
	tests := []struct {
		name          string
		query         *dto.BugQueryDTO
		mockBugs      []zentao.Bug
		mockError     error
		expectedTotal int
		expectedLen   int
		expectError   bool
	}{
		{
			name: "获取所有Bug",
			query: &dto.BugQueryDTO{
				ProductID: 100,
				Page:      1,
				Limit:     20,
			},
			mockBugs:      createMockBugs(),
			expectedTotal: 3,
			expectedLen:   3,
			expectError:   false,
		},
		{
			name: "按状态筛选",
			query: &dto.BugQueryDTO{
				ProductID: 100,
				Status:    "active",
				Page:      1,
				Limit:     20,
			},
			mockBugs:      createMockBugs(),
			expectedTotal: 2,
			expectedLen:   2,
			expectError:   false,
		},
		{
			name: "按指派人筛选",
			query: &dto.BugQueryDTO{
				ProductID:  100,
				AssignedTo: "user1",
				Page:       1,
				Limit:      20,
			},
			mockBugs:      createMockBugs(),
			expectedTotal: 2,
			expectedLen:   2,
			expectError:   false,
		},
		{
			name: "按日期范围筛选",
			query: &dto.BugQueryDTO{
				ProductID: 100,
				StartDate: "2024-01-16",
				EndDate:   "2024-01-17",
				Page:      1,
				Limit:     20,
			},
			mockBugs:      createMockBugs(),
			expectedTotal: 2,
			expectedLen:   2,
			expectError:   false,
		},
		{
			name: "按具体日期筛选",
			query: &dto.BugQueryDTO{
				ProductID:    100,
				SpecificDate: "2024-01-15",
				Page:         1,
				Limit:        20,
			},
			mockBugs:      createMockBugs(),
			expectedTotal: 1,
			expectedLen:   1,
			expectError:   false,
		},
		{
			name: "分页测试",
			query: &dto.BugQueryDTO{
				ProductID: 100,
				Page:      1,
				Limit:     2,
			},
			mockBugs:      createMockBugs(),
			expectedTotal: 3,
			expectedLen:   2,
			expectError:   false,
		},
		{
			name: "无产品ID",
			query: &dto.BugQueryDTO{
				Page:  1,
				Limit: 20,
			},
			mockBugs:      createMockBugs(),
			expectedTotal: 0,
			expectedLen:   0,
			expectError:   false,
		},
		{
			name: "客户端错误",
			query: &dto.BugQueryDTO{
				ProductID: 100,
				Page:      1,
				Limit:     20,
			},
			mockError:   errors.New("连接失败"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建mock客户端
			mockClient := &MockClient{
				GetBugsFunc: func(productID int, limit int) ([]zentao.Bug, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockBugs, nil
				},
				SearchBugsFunc: func(params zentao.BugSearchParams) ([]zentao.Bug, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockBugs, nil
				},
			}

			// 创建服务（使用适配器模式）
			service := NewBugServiceWithClient(mockClient)

			// 执行测试
			result, err := service.GetBugs(tt.query)

			// 验证错误
			if tt.expectError {
				if err == nil {
					t.Error("期望返回错误，但没有错误")
				}
				return
			}

			if err != nil {
				t.Errorf("不期望返回错误: %v", err)
				return
			}

			// 验证结果
			if result.Total != tt.expectedTotal {
				t.Errorf("期望 total=%d, 实际 total=%d", tt.expectedTotal, result.Total)
			}

			if len(result.List.([]interface{})) != tt.expectedLen {
				t.Errorf("期望列表长度=%d, 实际列表长度=%d", tt.expectedLen, len(result.List.([]interface{})))
			}
		})
	}
}

// BugServiceWithClient 使用接口的服务包装器
type BugServiceWithClient struct {
	client BugClient
}

// BugClient 定义客户端接口
type BugClient interface {
	GetBugs(productID int, limit int) ([]zentao.Bug, error)
	GetBugsByProject(productID, projectID int, limit int) ([]zentao.Bug, error)
	SearchBugs(params zentao.BugSearchParams) ([]zentao.Bug, error)
}

// NewBugServiceWithClient 创建使用接口的服务
func NewBugServiceWithClient(client BugClient) *BugServiceWithClient {
	return &BugServiceWithClient{client: client}
}

// GetBugs 获取Bug列表
func (s *BugServiceWithClient) GetBugs(query *dto.BugQueryDTO) (*myzentao.PaginatedResult, error) {
	var bugs []zentao.Bug
	var err error

	// 如果有产品ID，按产品查询
	if query.ProductID != 0 {
		// 优先使用SearchBugs进行多条件筛选
		if query.AssignedTo != "" || query.Status != "" {
			params := zentao.BugSearchParams{
				ProductID:  query.ProductID,
				Status:     query.Status,
				AssignedTo: query.AssignedTo,
				Limit:      1000,
				Page:       1,
			}
			bugs, err = s.client.SearchBugs(params)
		} else if query.ProjectID != 0 {
			bugs, err = s.client.GetBugsByProject(query.ProductID, query.ProjectID, 1000)
		} else {
			bugs, err = s.client.GetBugs(query.ProductID, 1000)
		}

		if err != nil {
			return nil, err
		}
	} else {
		bugs = []zentao.Bug{}
	}

	// 返回简化的结果
	return &myzentao.PaginatedResult{
		List:  bugs,
		Total: len(bugs),
		Page:  query.Page,
		Limit: query.Limit,
	}, nil
}
