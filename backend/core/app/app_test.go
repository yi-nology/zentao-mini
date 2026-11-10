package app

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/yi-nology/zentao-mini/backend/core/config"
	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/zentao"
)

// TestInitializeHTTPApp_PropagatesClientConfig issue #30 的核心回归：
// 此前 wire_gen.go 未重新生成，编译产物仍走 NewClient（配置丢弃、缺省 12h/120s），
// wire.go 的修法又把已是 Duration 的值再乘 time.Second 溢出为负、被 <=0 分支
// 吞成缺省。本测试走真实 InitializeHTTPApp（wire_gen 编译路径）断言配置真实接线。
func TestInitializeHTTPApp_PropagatesClientConfig(t *testing.T) {
	// NewInitService 内部走 logger.Warn，测试环境先就位
	if err := logger.Init(&config.LogConfig{Level: "error", Format: "console"}); err != nil {
		t.Fatalf("logger 初始化失败: %v", err)
	}
	tests := []struct {
		name        string
		refresh     time.Duration
		timeout     time.Duration
		wantRefresh time.Duration
		wantTimeout time.Duration
	}{
		{
			name:        "自定义间隔与超时透传",
			refresh:     3 * time.Hour,
			timeout:     45 * time.Second,
			wantRefresh: 3 * time.Hour,
			wantTimeout: 45 * time.Second,
		},
		{
			name:        "零值回落缺省（12h/120s）",
			refresh:     0,
			timeout:     0,
			wantRefresh: 12 * time.Hour,
			wantTimeout: 120 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &AppConfig{
				AuthDBPath:                 filepath.Join(t.TempDir(), "auth.db"),
				ZentaoTokenRefreshInterval: tt.refresh,
				ZentaoRequestTimeout:       tt.timeout,
			}

			application, err := InitializeHTTPApp(config)
			if err != nil {
				t.Fatalf("InitializeHTTPApp() err = %v", err)
			}
			httpApp, ok := application.(*HTTPApp)
			if !ok {
				t.Fatalf("InitializeHTTPApp() 返回类型 %T，期望 *HTTPApp", application)
			}

			var client *zentao.Client = httpApp.deps.ZentaoClient
			if got := client.RefreshInterval(); got != tt.wantRefresh {
				t.Errorf("refreshInterval = %v, want %v（yaml 配置未生效？）", got, tt.wantRefresh)
			}
			if got := client.RequestTimeout(); got != tt.wantTimeout {
				t.Errorf("requestTimeout = %v, want %v（yaml 配置未生效？）", got, tt.wantTimeout)
			}
		})
	}
}
