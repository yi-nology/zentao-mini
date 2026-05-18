package service

import (
	"errors"
	"testing"

	"github.com/yi-nology/common/biz/zentao"

	"github.com/yi-nology/zentao-mini/backend/core/dto"
	"github.com/yi-nology/zentao-mini/backend/core/models"
)

type MockClient struct {
	GetBugsFunc          func(productID int, pageSize int) ([]zentao.Bug, error)
	GetBugsByProjectFunc func(productID, projectID int, pageSize int) ([]zentao.Bug, error)
	SearchBugsFunc       func(params zentao.BugSearchParams) ([]zentao.Bug, error)
}

func (m *MockClient) GetBugs(productID int, pageSize int) ([]zentao.Bug, error) {
	if m.GetBugsFunc != nil {
		return m.GetBugsFunc(productID, pageSize)
	}
	return []zentao.Bug{}, nil
}

func (m *MockClient) GetBugsByProject(productID, projectID int, pageSize int) ([]zentao.Bug, error) {
	if m.GetBugsByProjectFunc != nil {
		return m.GetBugsByProjectFunc(productID, projectID, pageSize)
	}
	return []zentao.Bug{}, nil
}

func (m *MockClient) SearchBugs(params zentao.BugSearchParams) ([]zentao.Bug, error) {
	if m.SearchBugsFunc != nil {
		return m.SearchBugsFunc(params)
	}
	return []zentao.Bug{}, nil
}

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
				PageSize:  20,
			},
			mockBugs:      createMockBugs(),
			expectedTotal: 3,
			expectedLen:   3,
			expectError:   false,
		},
		{
			name: "无产品ID",
			query: &dto.BugQueryDTO{
				Page:     1,
				PageSize: 20,
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
				PageSize:  20,
			},
			mockError:   errors.New("连接失败"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{
				GetBugsFunc: func(productID int, pageSize int) ([]zentao.Bug, error) {
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

			service := NewBugServiceWithClient(mockClient)

			result, err := service.GetBugs(tt.query)

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

			if result.Total != tt.expectedTotal {
				t.Errorf("期望 total=%d, 实际 total=%d", tt.expectedTotal, result.Total)
			}

			if len(result.List.([]zentao.Bug)) != tt.expectedLen {
				t.Errorf("期望列表长度=%d, 实际列表长度=%d", tt.expectedLen, len(result.List.([]zentao.Bug)))
			}
		})
	}
}

type BugServiceWithClient struct {
	client BugClient
}

type BugClient interface {
	GetBugs(productID int, pageSize int) ([]zentao.Bug, error)
	GetBugsByProject(productID, projectID int, pageSize int) ([]zentao.Bug, error)
	SearchBugs(params zentao.BugSearchParams) ([]zentao.Bug, error)
}

func NewBugServiceWithClient(client BugClient) *BugServiceWithClient {
	return &BugServiceWithClient{client: client}
}

func (s *BugServiceWithClient) GetBugs(query *dto.BugQueryDTO) (*models.PaginatedResult, error) {
	var bugs []zentao.Bug
	var err error

	if query.ProductID != 0 {
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

	return &models.PaginatedResult{
		List:     bugs,
		Total:    len(bugs),
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}
