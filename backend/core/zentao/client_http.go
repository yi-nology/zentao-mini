package zentao

import (
	"crypto/tls"
	"net/http"
	"sync"
	"time"
)

// zentaoHTTPClient 共享 HTTP 客户端单例：直连禅道服务的调用统一走这里。
// 自签名证书场景跳过 TLS 校验（与上游 SDK doGet 行为一致），120s 超时。
//
// 此前 getBugsTolerant 与 GetAllBugsIncludeClosed 每次调用各 new 一份
// client——Transport 不复用即没有连接池，TLS 握手每次重付。单例后同进程
// 内连接可复用（issue #20 第 1 项的「共享 newZentaoHTTPClient」）。
var (
	zentaoHTTPClientOnce sync.Once
	zentaoHTTPClientInst *http.Client
)

func newZentaoHTTPClient() *http.Client {
	zentaoHTTPClientOnce.Do(func() {
		zentaoHTTPClientInst = &http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	})
	return zentaoHTTPClientInst
}
