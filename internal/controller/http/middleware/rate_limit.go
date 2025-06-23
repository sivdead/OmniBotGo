package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

const (
	// 默认速率限制值
	DefaultGlobalMax         = 1000
	DefaultPerIPMax          = 100
	DefaultPerUserMax        = 200
	DefaultGlobalExpiration  = 1 * time.Minute
	DefaultPerIPExpiration   = 1 * time.Minute
	DefaultPerUserExpiration = 1 * time.Minute
	
	// 垃圾回收间隔
	GCInterval = 10 * time.Second
)

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	// 全局限制
	GlobalMax        int           // 全局最大请求数
	GlobalExpiration time.Duration // 全局时间窗口

	// 每个IP限制
	PerIPMax        int           // 每个IP最大请求数
	PerIPExpiration time.Duration // 每个IP时间窗口

	// 每个用户限制（需要认证）
	PerUserMax        int           // 每个用户最大请求数
	PerUserExpiration time.Duration // 每个用户时间窗口

	// 跳过的路径
	SkipPaths []string

	// 日志
	Logger logger.Interface
}

// DefaultRateLimitConfig 默认速率限制配置
func DefaultRateLimitConfig(l logger.Interface) RateLimitConfig {
	return RateLimitConfig{
		GlobalMax:         DefaultGlobalMax,
		GlobalExpiration:  DefaultGlobalExpiration,
		PerIPMax:          DefaultPerIPMax,
		PerIPExpiration:   DefaultPerIPExpiration,
		PerUserMax:        DefaultPerUserMax,
		PerUserExpiration: DefaultPerUserExpiration,
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/swagger",
		},
		Logger: l,
	}
}

// RateLimiter 速率限制器
type RateLimiter struct {
	config RateLimitConfig
	stores map[string]*MemoryStore
	mu     sync.RWMutex
	done   chan struct{} // 用于优雅关闭
}

// NewRateLimiter 创建速率限制中间件
func NewRateLimiter(config RateLimitConfig) fiber.Handler {
	limiter := &RateLimiter{
		config: config,
		stores: map[string]*MemoryStore{
			"global": NewMemoryStore(),
			"ip":     NewMemoryStore(),
			"user":   NewMemoryStore(),
		},
		done: make(chan struct{}),
	}

	return limiter.Handle
}

// Shutdown 优雅关闭速率限制器
func (r *RateLimiter) Shutdown() {
	close(r.done)
	r.mu.RLock()
	for _, store := range r.stores {
		store.Shutdown()
	}
	r.mu.RUnlock()
}

// Handle 处理请求
func (r *RateLimiter) Handle(c *fiber.Ctx) error {
	// 检查是否应该跳过
	if r.shouldSkip(c) {
		return c.Next()
	}

	// 记录请求
	r.config.Logger.Debug("Rate limit check",
		"path", c.Path(),
		"ip", c.IP(),
		"method", c.Method())

	// 检查全局限制
	if !r.checkLimit("global", "global", r.config.GlobalMax, r.config.GlobalExpiration) {
		return r.limitReached(c)
	}

	// 检查IP限制
	if !r.checkLimit("ip", c.IP(), r.config.PerIPMax, r.config.PerIPExpiration) {
		return r.limitReached(c)
	}

	// 检查用户限制
	userID := c.Locals("userID")
	if userID != nil {
		key := fmt.Sprintf("user:%v", userID)
		if !r.checkLimit("user", key, r.config.PerUserMax, r.config.PerUserExpiration) {
			return r.limitReached(c)
		}
	}

	return c.Next()
}

// checkLimit 检查限制
func (r *RateLimiter) checkLimit(storeKey, key string, max int, expiration time.Duration) bool {
	// -1 或 0 表示不限制
	if max <= 0 {
		return true
	}

	r.mu.RLock()
	store := r.stores[storeKey]
	r.mu.RUnlock()

	hits, _ := store.Get(key)
	if hits >= max {
		return false
	}

	store.Set(key, hits+1, expiration)
	return true
}

// shouldSkip 检查是否应该跳过限制
func (r *RateLimiter) shouldSkip(c *fiber.Ctx) bool {
	path := c.Path()
	for _, skipPath := range r.config.SkipPaths {
		if path == skipPath || (len(skipPath) > 0 && skipPath[len(skipPath)-1] == '*' &&
			len(path) >= len(skipPath)-1 && path[:len(skipPath)-1] == skipPath[:len(skipPath)-1]) {
			return true
		}
	}
	return false
}

// limitReached 处理限制达到
func (r *RateLimiter) limitReached(c *fiber.Ctx) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
		"error": "Too many requests, please try again later",
	})
}

// MemoryStore 内存存储实现
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]entry
	done chan struct{} // 用于停止垃圾回收
}

type entry struct {
	hits   int
	expire time.Time
}

// NewMemoryStore 创建内存存储
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		data: make(map[string]entry),
		done: make(chan struct{}),
	}
	// 启动垃圾回收
	go store.gc()
	return store
}

// Get 获取计数
func (s *MemoryStore) Get(key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if e, ok := s.data[key]; ok {
		if time.Now().Before(e.expire) {
			return e.hits, nil
		}
	}
	return 0, nil
}

// Set 设置计数
func (s *MemoryStore) Set(key string, hits int, exp time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = entry{
		hits:   hits,
		expire: time.Now().Add(exp),
	}
	return nil
}

// Reset 重置存储
func (s *MemoryStore) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]entry)
	return nil
}

// Shutdown 关闭存储
func (s *MemoryStore) Shutdown() {
	close(s.done)
}

// gc 垃圾回收过期条目
func (s *MemoryStore) gc() {
	ticker := time.NewTicker(GCInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			for key, e := range s.data {
				if now.After(e.expire) {
					delete(s.data, key)
				}
			}
			s.mu.Unlock()
		case <-s.done:
			return // 优雅退出
		}
	}
}

// PlatformRateLimiter 平台特定的速率限制器
type PlatformRateLimiter struct {
	limits map[string]PlatformLimit
	stores map[string]*MemoryStore
	mu     sync.RWMutex
	logger logger.Interface
}

// PlatformLimit 平台限制配置
type PlatformLimit struct {
	Platform   string
	Max        int
	Expiration time.Duration
}

// NewPlatformRateLimiter 创建平台速率限制器
func NewPlatformRateLimiter(logger logger.Interface) *PlatformRateLimiter {
	return &PlatformRateLimiter{
		limits: make(map[string]PlatformLimit),
		stores: make(map[string]*MemoryStore),
		logger: logger,
	}
}

// SetPlatformLimit 设置平台限制
func (p *PlatformRateLimiter) SetPlatformLimit(platform string, max int, expiration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.limits[platform] = PlatformLimit{
		Platform:   platform,
		Max:        max,
		Expiration: expiration,
	}

	if _, ok := p.stores[platform]; !ok {
		p.stores[platform] = NewMemoryStore()
	}
}

// CheckLimit 检查平台限制
func (p *PlatformRateLimiter) CheckLimit(platform, key string) (bool, error) {
	p.mu.RLock()
	limit, hasLimit := p.limits[platform]
	store, hasStore := p.stores[platform]
	p.mu.RUnlock()

	if !hasLimit || !hasStore {
		// 没有限制，允许通过
		return true, nil
	}

	// 获取当前计数
	hits, err := store.Get(key)
	if err != nil {
		return false, err
	}

	// 检查是否超限
	if hits >= limit.Max {
		p.logger.Warn("Platform rate limit exceeded",
			"platform", platform,
			"key", key,
			"hits", hits,
			"max", limit.Max)
		return false, nil
	}

	// 增加计数
	err = store.Set(key, hits+1, limit.Expiration)
	return true, err
}

// ConcurrentLimitConfig 并发限制配置
type ConcurrentLimitConfig struct {
	// 最大并发工作协程数（-1表示不限制）
	MaxWorkers int
	// 最大同时处理请求数（-1表示不限制）
	MaxRequests int
}

// DefaultConcurrentLimitConfig 默认并发限制配置
func DefaultConcurrentLimitConfig() ConcurrentLimitConfig {
	return ConcurrentLimitConfig{
		MaxWorkers:  100,  // 最大100个并发工作协程，-1表示不限制
		MaxRequests: 1000, // 最大1000个同时处理的请求，-1表示不限制
	}
}

// ConcurrentLimiter 并发限制器
type ConcurrentLimiter struct {
	config         ConcurrentLimitConfig
	workerSem      chan struct{} // 工作协程信号量
	requestCounter *RequestCounter
	logger         logger.Interface
}

// RequestCounter 请求计数器
type RequestCounter struct {
	current int
	mu      sync.RWMutex
}

// NewConcurrentLimiter 创建并发限制中间件
func NewConcurrentLimiter(config ConcurrentLimitConfig, l logger.Interface) fiber.Handler {
	limiter := &ConcurrentLimiter{
		config:         config,
		requestCounter: &RequestCounter{},
		logger:         l,
	}

	// 初始化工作协程信号量
	if config.MaxWorkers > 0 {
		limiter.workerSem = make(chan struct{}, config.MaxWorkers)
		l.Info("启用工作协程限制", "max_workers", config.MaxWorkers)
	} else {
		l.Info("工作协程限制已禁用", "max_workers", config.MaxWorkers)
	}

	if config.MaxRequests > 0 {
		l.Info("启用请求数量限制", "max_requests", config.MaxRequests)
	} else {
		l.Info("请求数量限制已禁用", "max_requests", config.MaxRequests)
	}

	return limiter.Handle
}

// Handle 处理请求
func (cl *ConcurrentLimiter) Handle(c *fiber.Ctx) error {
	// 工作协程限制
	if cl.config.MaxWorkers > 0 {
		select {
		case cl.workerSem <- struct{}{}:
			defer func() { <-cl.workerSem }()
		default:
			cl.logger.Warn("工作协程数量已达上限",
				"max_workers", cl.config.MaxWorkers,
				"path", c.Path(),
				"ip", c.IP())
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Service temporarily unavailable, too many concurrent workers",
			})
		}
	}

	// 请求数量限制
	if cl.config.MaxRequests > 0 {
		if !cl.requestCounter.TryIncrement(cl.config.MaxRequests) {
			cl.logger.Warn("同时处理请求数已达上限",
				"max_requests", cl.config.MaxRequests,
				"current", cl.requestCounter.GetCurrent(),
				"path", c.Path(),
				"ip", c.IP())
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Too many concurrent requests",
			})
		}
		defer cl.requestCounter.Decrement()
	}

	return c.Next()
}

// TryIncrement 尝试增加计数器
func (rc *RequestCounter) TryIncrement(max int) bool {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.current >= max {
		return false
	}
	rc.current++
	return true
}

// Decrement 减少计数器
func (rc *RequestCounter) Decrement() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.current > 0 {
		rc.current--
	}
}

// GetCurrent 获取当前计数
func (rc *RequestCounter) GetCurrent() int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.current
}
