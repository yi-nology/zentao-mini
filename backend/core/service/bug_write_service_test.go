package service

import (
	"strings"
	"testing"
)

// 写操作入参校验：这些分支在触达禅道客户端前就应拒绝，
// 因此 client 为 nil 也不会 panic。

func TestBugService_AddBugCommentValidation(t *testing.T) {
	s := &BugService{client: nil}

	if _, err := s.AddBugComment(0, "评论"); err == nil {
		t.Error("无效 Bug ID 应报错")
	}
	if _, err := s.AddBugComment(123, ""); err == nil {
		t.Error("空评论应报错")
	}
	if _, err := s.AddBugComment(123, "   "); err == nil {
		t.Error("空白评论应报错")
	}
}

func TestBugService_TransitionValidation(t *testing.T) {
	s := &BugService{client: nil}

	cases := []struct {
		name      string
		bugID     int
		input     *BugTransitionInput
		errSubstr string
	}{
		{"无效ID", 0, &BugTransitionInput{Action: "confirm"}, "无效的 Bug ID"},
		{"缺少参数", 123, nil, "缺少流转参数"},
		{"未知动作", 123, &BugTransitionInput{Action: "delete"}, "不支持的状态流转动作"},
		{"resolve缺resolution", 123, &BugTransitionInput{Action: "resolve"}, "必须指定 resolution"},
		{"resolve非法resolution", 123, &BugTransitionInput{Action: "resolve", Resolution: "magic"}, "不支持的 resolution"},
		{"assign缺assignedTo", 123, &BugTransitionInput{Action: "assign"}, "必须指定 assignedTo"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.TransitionBug(tc.bugID, tc.input)
			if err == nil {
				t.Fatalf("应报错: %+v", tc.input)
			}
			if !strings.Contains(err.Error(), tc.errSubstr) {
				t.Errorf("错误信息应包含 %q, got %q", tc.errSubstr, err.Error())
			}
		})
	}
}
