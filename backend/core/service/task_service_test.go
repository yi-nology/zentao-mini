package service

import (
	"testing"

	"github.com/yi-nology/zentao-mini/backend/core/dto"
	"github.com/yi-nology/zentao-mini/backend/core/vo"
)

// TestTaskService_GetTasks_NoProductReturnsEmpty 未选产品/执行时应返回空列表而不是报错（与 BugService 口径一致）
func TestTaskService_GetTasks_NoProductReturnsEmpty(t *testing.T) {
	service := &TaskService{client: nil}

	result, err := service.GetTasks(&dto.TaskQueryDTO{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if list, ok := result.List.([]vo.TaskVO); ok && len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}

// TestTaskService_GetTaskStatusCounts_NoProductReturnsZeros 未选产品时统计应全为 0
func TestTaskService_GetTaskStatusCounts_NoProductReturnsZeros(t *testing.T) {
	service := &TaskService{client: nil}

	counts, err := service.GetTaskStatusCounts(&dto.TaskQueryDTO{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	for _, status := range []string{"doing", "wait", "done", "closed"} {
		if counts[status] != 0 {
			t.Errorf("expected counts[%s]=0, got %d", status, counts[status])
		}
	}
}

// TestTaskQueryDTO_AllowEmptyProduct 空产品应为合法查询（分页参数仍需规范化）
func TestTaskQueryDTO_AllowEmptyProduct(t *testing.T) {
	query := &dto.TaskQueryDTO{}
	if err := query.Validate(); err != nil {
		t.Fatalf("expected no error for empty product, got %v", err)
	}
	if query.Page != 1 || query.PageSize != 20 {
		t.Errorf("expected page=1 pageSize=20, got page=%d pageSize=%d", query.Page, query.PageSize)
	}
}
