package zentao

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 回归：GetOrLoadWithLock 的 loadFunc（网络请求）执行期间不得阻塞 Clear/Set 等全局操作，
// 否则与 UpdateConfig（持 client.mu 调 Clear）形成互相等待死锁。
func TestGetOrLoadWithLockDoesNotBlockClear(t *testing.T) {
	cache := NewMemoryCache()
	release := make(chan struct{})
	loaded := make(chan struct{})

	go func() {
		_, _ = cache.GetOrLoadWithLock("slow-key", func() (interface{}, error) {
			close(loaded)
			<-release
			return "value", nil
		}, time.Minute)
	}()

	<-loaded
	cleared := make(chan struct{})
	go func() {
		cache.Clear()
		close(cleared)
	}()

	select {
	case <-cleared:
	case <-time.After(2 * time.Second):
		t.Fatal("loadFunc 执行期间 Clear 被阻塞，存在全局死锁风险")
	}
	close(release)
}

// 同 key 并发加载应去重（只加载一次），不同 key 应并行加载。
func TestGetOrLoadWithLockKeyIsolation(t *testing.T) {
	cache := NewMemoryCache()
	release := make(chan struct{})

	var loadsA, loadsB int32
	loadA := func() (interface{}, error) {
		atomic.AddInt32(&loadsA, 1)
		<-release
		return "a", nil
	}
	loadB := func() (interface{}, error) {
		atomic.AddInt32(&loadsB, 1)
		<-release
		return "b", nil
	}

	// 同 key 三个并发 + 另一个 key 一个并发
	var wg sync.WaitGroup
	started := make(chan struct{}, 4)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			started <- struct{}{}
			_, _ = cache.GetOrLoadWithLock("key-a", loadA, time.Minute)
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		started <- struct{}{}
		_, _ = cache.GetOrLoadWithLock("key-b", loadB, time.Minute)
	}()

	// key-b 的加载不应被 key-a 阻塞
	bDone := make(chan struct{})
	go func() {
		_, _ = cache.GetOrLoadWithLock("key-b", loadB, time.Minute)
		close(bDone)
	}()
	select {
	case <-bDone:
		t.Fatal("不同 key 的加载被互相阻塞，per-key 隔离失效")
	case <-time.After(300 * time.Millisecond):
	}

	close(release)
	wg.Wait()

	if n := atomic.LoadInt32(&loadsA); n != 1 {
		t.Fatalf("同 key 并发加载应去重为 1 次, 实际 %d 次", n)
	}
}

// 回归：并发 401 时 RefreshToken 的 CAS 输家不应立即报"刷新进行中"放弃，
// 而应等待持锁者完成后复用结果（2026-10-08 每日bug 定时推送线上故障）。
func TestRefreshOrWaitWaitsForInFlightRefresh(t *testing.T) {
	c := NewClient("", "", "")
	c.refreshing.Store(true) // 模拟另一协程正在刷新

	done := make(chan error, 1)
	go func() { done <- c.refreshOrWait(nil) }()

	time.Sleep(300 * time.Millisecond)
	select {
	case err := <-done:
		t.Fatalf("持锁未释放时不应返回: %v", err)
	default:
	}

	c.refreshing.Store(false) // 持锁者完成
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("等待刷新完成后应返回 nil: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("refreshOrWait 未在持锁者完成后及时返回")
	}
}
