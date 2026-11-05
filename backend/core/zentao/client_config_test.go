package zentao

import (
	"testing"
	"time"
)

// TestNewClientWithConfigConsumesValues 配置面接线的钉子：此前
// token_refresh_interval / request_timeout 被 viper 读入、setDefaults 校验，
// client 却硬编码 2h ticker 与 120s 超时——「配置在撒谎」。这里钉住
// NewClientWithConfig 把值真的装进 client（ticker 间隔与超时从此跟随配置）。
func TestNewClientWithConfigConsumesValues(t *testing.T) {
	c := NewClientWithConfig("http://zt.example.com", "user", "pass", 6*time.Hour, 30*time.Second)
	if c.refreshInterval != 6*time.Hour {
		t.Fatalf("refreshInterval = %v, want 6h（配置必须生效）", c.refreshInterval)
	}
	if c.requestTimeout != 30*time.Second {
		t.Fatalf("requestTimeout = %v, want 30s（配置必须生效）", c.requestTimeout)
	}
	// 零值 = 缺省（12h / 120s）。
	c2 := NewClientWithConfig("http://zt.example.com", "user", "pass", 0, 0)
	if c2.refreshInterval != defaultRefreshInterval || c2.requestTimeout != defaultRequestTimeout {
		t.Fatalf("零值应回落缺省: interval=%v timeout=%v", c2.refreshInterval, c2.requestTimeout)
	}
	// 旧签名 NewClient = 缺省形态。
	c3 := NewClient("http://zt.example.com", "user", "pass")
	if c3.refreshInterval != defaultRefreshInterval || c3.requestTimeout != defaultRequestTimeout {
		t.Fatalf("NewClient 应保持缺省形态: interval=%v timeout=%v", c3.refreshInterval, c3.requestTimeout)
	}
}
