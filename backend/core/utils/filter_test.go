package utils

import (
	"testing"
)

// TestFilter 测试通用过滤函数
func TestFilter(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		filter   func(int) bool
		expected []int
	}{
		{
			name:     "过滤偶数",
			slice:    []int{1, 2, 3, 4, 5, 6},
			filter:   func(n int) bool { return n%2 == 0 },
			expected: []int{2, 4, 6},
		},
		{
			name:     "过滤大于3的数",
			slice:    []int{1, 2, 3, 4, 5},
			filter:   func(n int) bool { return n > 3 },
			expected: []int{4, 5},
		},
		{
			name:     "空切片",
			slice:    []int{},
			filter:   func(n int) bool { return n > 0 },
			expected: []int{},
		},
		{
			name:     "全部过滤掉",
			slice:    []int{1, 2, 3},
			filter:   func(n int) bool { return n > 10 },
			expected: []int{},
		},
		{
			name:     "全部保留",
			slice:    []int{1, 2, 3},
			filter:   func(n int) bool { return true },
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Filter(tt.slice, tt.filter)

			if len(result) != len(tt.expected) {
				t.Errorf("期望长度=%d, 实际长度=%d", len(tt.expected), len(result))
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("索引 %d: 期望=%d, 实际=%d", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

// TestFilterByDateRange 测试按日期范围过滤
func TestFilterByDateRange(t *testing.T) {
	type testItem struct {
		name string
		date string
	}

	tests := []struct {
		name      string
		slice     []testItem
		startDate string
		endDate   string
		expected  int
	}{
		{
			name: "日期范围内过滤",
			slice: []testItem{
				{name: "item1", date: "2024-01-15"},
				{name: "item2", date: "2024-01-20"},
				{name: "item3", date: "2024-01-25"},
			},
			startDate: "2024-01-18",
			endDate:   "2024-01-22",
			expected:  1,
		},
		{
			name: "无日期范围限制",
			slice: []testItem{
				{name: "item1", date: "2024-01-15"},
				{name: "item2", date: "2024-01-20"},
			},
			startDate: "",
			endDate:   "",
			expected:  2,
		},
		{
			name: "只有开始日期",
			slice: []testItem{
				{name: "item1", date: "2024-01-15"},
				{name: "item2", date: "2024-01-20"},
				{name: "item3", date: "2024-01-25"},
			},
			startDate: "2024-01-18",
			endDate:   "",
			expected:  2,
		},
		{
			name: "只有结束日期",
			slice: []testItem{
				{name: "item1", date: "2024-01-15"},
				{name: "item2", date: "2024-01-20"},
				{name: "item3", date: "2024-01-25"},
			},
			startDate: "",
			endDate:   "2024-01-18",
			expected:  1,
		},
		{
			name: "带时间戳的日期",
			slice: []testItem{
				{name: "item1", date: "2024-01-15 10:30:00"},
				{name: "item2", date: "2024-01-20 15:45:00"},
			},
			startDate: "2024-01-15",
			endDate:   "2024-01-15",
			expected:  1,
		},
		{
			name: "空日期字段",
			slice: []testItem{
				{name: "item1", date: ""},
				{name: "item2", date: "2024-01-20"},
			},
			startDate: "2024-01-01",
			endDate:   "2024-01-31",
			expected:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterByDateRange(tt.slice, tt.startDate, tt.endDate, func(item testItem) string {
				return item.date
			})

			if len(result) != tt.expected {
				t.Errorf("期望长度=%d, 实际长度=%d", tt.expected, len(result))
			}
		})
	}
}

// TestFilterBySpecificDate 测试按具体日期过滤
func TestFilterBySpecificDate(t *testing.T) {
	type testItem struct {
		name string
		date string
	}

	tests := []struct {
		name         string
		slice        []testItem
		specificDate string
		expected     int
	}{
		{
			name: "匹配具体日期",
			slice: []testItem{
				{name: "item1", date: "2024-01-15"},
				{name: "item2", date: "2024-01-15 10:30:00"},
				{name: "item3", date: "2024-01-20"},
			},
			specificDate: "2024-01-15",
			expected:     2,
		},
		{
			name: "无匹配",
			slice: []testItem{
				{name: "item1", date: "2024-01-15"},
				{name: "item2", date: "2024-01-20"},
			},
			specificDate: "2024-01-25",
			expected:     0,
		},
		{
			name: "空具体日期",
			slice: []testItem{
				{name: "item1", date: "2024-01-15"},
			},
			specificDate: "",
			expected:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterBySpecificDate(tt.slice, tt.specificDate, func(item testItem) string {
				return item.date
			})

			if len(result) != tt.expected {
				t.Errorf("期望长度=%d, 实际长度=%d", tt.expected, len(result))
			}
		})
	}
}

// TestChainFilter 测试链式过滤器
func TestChainFilter(t *testing.T) {
	type testItem struct {
		id     int
		status string
		date   string
	}

	items := []testItem{
		{id: 1, status: "active", date: "2024-01-15"},
		{id: 2, status: "closed", date: "2024-01-16"},
		{id: 3, status: "active", date: "2024-01-17"},
		{id: 4, status: "active", date: "2024-01-18"},
		{id: 5, status: "closed", date: "2024-01-19"},
	}

	t.Run("链式过滤和分页", func(t *testing.T) {
		result := NewChainFilter(items).
			Filter(func(item testItem) bool { return item.status == "active" }).
			FilterByDate("2024-01-16", "2024-01-18", func(item testItem) string { return item.date }).
			Paginate(1, 2).
			Result()

		// 应该返回 id=3 和 id=4（active且日期在范围内）
		if len(result) != 2 {
			t.Errorf("期望长度=2, 实际长度=%d", len(result))
		}
	})

	t.Run("计数功能", func(t *testing.T) {
		count := NewChainFilter(items).
			Filter(func(item testItem) bool { return item.status == "active" }).
			Count()

		if count != 3 {
			t.Errorf("期望 count=3, 实际 count=%d", count)
		}
	})
}
