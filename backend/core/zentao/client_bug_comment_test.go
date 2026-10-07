package zentao

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yi-nology/zentao-mini/backend/core/config"
	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/metrics"
)

func TestAddBugCommentViaWebSession(t *testing.T) {
	if err := metrics.Init(); err != nil {
		t.Fatalf("metrics 初始化失败: %v", err)
	}
	if err := logger.Init(&config.LogConfig{Level: "error", Format: "console"}); err != nil {
		t.Fatalf("logger 初始化失败: %v", err)
	}
	loggedIn := false
	commentReceived := ""

	mux := http.NewServeMux()
	mux.HandleFunc("/user-login.html", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			http.SetCookie(w, &http.Cookie{Name: "zentaosid", Value: "sid-abc"})
			_, _ = w.Write([]byte("<html>login page</html>"))
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				t.Errorf("解析登录表单失败: %v", err)
			}
			if r.Form.Get("account") == "tester" && r.Form.Get("password") == "secret" {
				loggedIn = true
				_, _ = w.Write([]byte(`{"result":"success","locate":"/index.html"}`))
				return
			}
			_, _ = w.Write([]byte(`{"result":"fail","message":"登录失败"}`))
		}
	})
	mux.HandleFunc("/action-comment-bug-456.html", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("zentaosid")
		if err != nil || cookie.Value != "sid-abc" || !loggedIn {
			http.Redirect(w, r, "/user-login.html", http.StatusFound)
			return
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("解析评论表单失败: %v", err)
		}
		commentReceived = r.Form.Get("comment")
		_, _ = w.Write([]byte(`{"status":"success","closeModal":true}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "tester", "secret")
	defer GlobalCache.Clear()
	defer client.ResetWebSession()

	if err := client.AddBugComment(456, "zentao-mini 写能力测试评论"); err != nil {
		t.Fatalf("AddBugComment 失败: %v", err)
	}
	if commentReceived != "zentao-mini 写能力测试评论" {
		t.Errorf("禅道应收到评论内容, got %q", commentReceived)
	}
}

func TestAddBugCommentLoginRejected(t *testing.T) {
	if err := metrics.Init(); err != nil {
		t.Fatalf("metrics 初始化失败: %v", err)
	}
	if err := logger.Init(&config.LogConfig{Level: "error", Format: "console"}); err != nil {
		t.Fatalf("logger 初始化失败: %v", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/user-login.html", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":"fail","message":"用户名或密码错误"}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "tester", "wrong-password")
	defer GlobalCache.Clear()
	defer client.ResetWebSession()

	err := client.AddBugComment(456, "不应成功的评论")
	if err == nil {
		t.Fatal("登录失败时 AddBugComment 应报错")
	}
	if !strings.Contains(err.Error(), "登录") {
		t.Errorf("错误信息应说明登录问题, got %q", err.Error())
	}
}
