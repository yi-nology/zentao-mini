package zentao

import (
	"sync"
	"time"
)

// CacheItem 缓存项
type CacheItem struct {
	Value      interface{}
	Expiry     time.Time
	CreateTime time.Time
}

// IsExpired 检查缓存项是否过期
func (item *CacheItem) IsExpired() bool {
	return time.Now().After(item.Expiry)
}

// MemoryCache 内存缓存
type MemoryCache struct {
	items map[string]*CacheItem
	mu    sync.RWMutex
}

// NewMemoryCache 创建新的内存缓存
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*CacheItem),
	}

	// 启动后台清理任务
	go cache.cleanupTask()

	return cache
}

// cleanupTask 定期清理过期缓存
func (c *MemoryCache) cleanupTask() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.Expiry) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// Set 设置缓存
func (c *MemoryCache) Set(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheItem{
		Value:      value,
		Expiry:     time.Now().Add(duration),
		CreateTime: time.Now(),
	}
}

// Get 获取缓存
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if item.IsExpired() {
		return nil, false
	}

	return item.Value, true
}

// Delete 删除缓存
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear 清空所有缓存
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
}

// GetOrLoad 获取缓存，如果不存在则加载
func (c *MemoryCache) GetOrLoad(key string, loadFunc func() (interface{}, error), duration time.Duration) (interface{}, error) {
	// 先尝试从缓存获取
	if value, exists := c.Get(key); exists {
		return value, nil
	}

	// 缓存不存在，执行加载函数
	value, err := loadFunc()
	if err != nil {
		return nil, err
	}

	// 存入缓存
	c.Set(key, value, duration)

	return value, nil
}

// GetOrLoadWithLock 获取缓存，如果不存在则加载（带锁，防止缓存击穿）
func (c *MemoryCache) GetOrLoadWithLock(key string, loadFunc func() (interface{}, error), duration time.Duration) (interface{}, error) {
	// 先尝试从缓存获取
	if value, exists := c.Get(key); exists {
		return value, nil
	}

	// 使用写锁防止缓存击穿
	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查
	if item, exists := c.items[key]; exists && !item.IsExpired() {
		return item.Value, nil
	}

	// 执行加载函数
	value, err := loadFunc()
	if err != nil {
		return nil, err
	}

	// 存入缓存
	c.items[key] = &CacheItem{
		Value:      value,
		Expiry:     time.Now().Add(duration),
		CreateTime: time.Now(),
	}

	return value, nil
}

// Size 获取缓存大小
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// Stats 缓存统计信息
type CacheStats struct {
	TotalItems int
	ExpiredItems int
}

// GetStats 获取缓存统计信息
func (c *MemoryCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := CacheStats{
		TotalItems: len(c.items),
	}

	now := time.Now()
	for _, item := range c.items {
		if now.After(item.Expiry) {
			stats.ExpiredItems++
		}
	}

	return stats
}

// GlobalCache 全局缓存实例
var GlobalCache = NewMemoryCache()

// CacheKeyBuilder 缓存键构建器
type CacheKeyBuilder struct{}

// Build 构建缓存键
func (b *CacheKeyBuilder) Build(prefix string, parts ...string) string {
	key := prefix
	for _, part := range parts {
		key += ":" + part
	}
	return key
}

// ProductCacheKey 产品缓存键
func (b *CacheKeyBuilder) ProductCacheKey(productID int) string {
	return b.Build("product", string(rune(productID)))
}

// ProjectCacheKey 项目缓存键
func (b *CacheKeyBuilder) ProjectCacheKey(projectID int) string {
	return b.Build("project", string(rune(projectID)))
}

// BugsCacheKey Bug缓存键
func (b *CacheKeyBuilder) BugsCacheKey(productID int, filters ...string) string {
	parts := []string{"bugs", string(rune(productID))}
	parts = append(parts, filters...)
	return b.Build(parts[0], parts[1:]...)
}

// StoriesCacheKey 需求缓存键
func (b *CacheKeyBuilder) StoriesCacheKey(productID int, filters ...string) string {
	parts := []string{"stories", string(rune(productID))}
	parts = append(parts, filters...)
	return b.Build(parts[0], parts[1:]...)
}

// TasksCacheKey 任务缓存键
func (b *CacheKeyBuilder) TasksCacheKey(executionID int, filters ...string) string {
	parts := []string{"tasks", string(rune(executionID))}
	parts = append(parts, filters...)
	return b.Build(parts[0], parts[1:]...)
}

// UsersCacheKey 用户缓存键
func (b *CacheKeyBuilder) UsersCacheKey() string {
	return "users:all"
}

// DefaultKeyBuilder 默认键构建器
var DefaultKeyBuilder = &CacheKeyBuilder{}
