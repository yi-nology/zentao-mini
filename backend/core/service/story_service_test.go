package service

import (
	"testing"

	"github.com/yi-nology/common/biz/zentao"

	"github.com/yi-nology/zentao-mini/backend/core/dto"
)

func createMockStories() []zentao.Story {
	return []zentao.Story{
		{ID: 1, Title: "S1", Status: "active", AssignedTo: map[string]interface{}{"account": "user1"}},
		{ID: 2, Title: "S2", Status: "reviewing"},
		{ID: 3, Title: "S3", Status: "closed"},
		{ID: 4, Title: "S4", Status: "active", AssignedTo: "user2"},
	}
}

// TestStoryService_StatusCounts_FilterChain 统计基于过滤链结果按状态计数
func TestStoryService_StatusCounts_FilterChain(t *testing.T) {
	counts := map[string]int{"draft": 0, "active": 0, "reviewing": 0, "changed": 0, "closed": 0}
	for _, story := range filterStoryChain(createMockStories(), &dto.StoryQueryDTO{Status: "active"}).Result() {
		if _, ok := counts[story.Status]; ok {
			counts[story.Status]++
		}
	}
	if counts["active"] != 2 {
		t.Errorf("expected active=2, got %v", counts)
	}
}

// TestStoryService_StatusCounts_AssignedToRef 指派人对象/字符串两种形态都能过滤
func TestStoryService_StatusCounts_AssignedToRef(t *testing.T) {
	matched := filterStoryChain(createMockStories(), &dto.StoryQueryDTO{AssignedTo: "USER1"}).Result()
	if len(matched) != 1 || matched[0].ID != 1 {
		t.Errorf("expected only story 1 (case-insensitive object ref), got %d matches", len(matched))
	}
}

// TestStoryService_StatusCounts_NoScope 未提供产品/项目/执行时应返回验证错误
func TestStoryService_StatusCounts_NoScope(t *testing.T) {
	service := &StoryService{client: nil}
	if _, err := service.GetStoryStatusCounts(&dto.StoryQueryDTO{Page: 1, PageSize: 20}); err == nil {
		t.Error("expected ValidationError for missing scope")
	}
}

// TestStoryService_ConvertToVO 基本 VO 转换
func TestStoryService_ConvertToVO(t *testing.T) {
	service := &StoryService{client: nil}
	vos := service.convertToVO(createMockStories())
	if len(vos) != 4 || vos[0].ID != 1 || vos[0].Title != "S1" {
		t.Errorf("unexpected VO conversion result: %+v", vos)
	}
}
