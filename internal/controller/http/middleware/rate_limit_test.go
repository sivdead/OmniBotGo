package middleware

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrentLimiter(t *testing.T) {
	// 创建测试logger
	l := logger.New("test")

	t.Run("工作协程限制测试", func(t *testing.T) {
		app := fiber.New()

		// 设置很小的限制用于测试
		config := ConcurrentLimitConfig{
			MaxWorkers:  2,
			MaxRequests: -1, // 不限制请求数
		}

		app.Use(NewConcurrentLimiter(config, l))

		// 添加一个会阻塞的处理器
		app.Get("/test", func(c *fiber.Ctx) error {
			time.Sleep(100 * time.Millisecond)
			return c.SendString("ok")
		})

		// 并发发送多个请求
		var wg sync.WaitGroup
		results := make([]int, 5)

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				req := httptest.NewRequest("GET", "/test", nil)
				resp, err := app.Test(req, 200) // 200ms超时
				if err != nil {
					results[index] = 500
				} else {
					results[index] = resp.StatusCode
				}
			}(i)
		}

		wg.Wait()

		// 应该有一些请求被拒绝（503状态码）
		rejectedCount := 0
		successCount := 0
		for _, status := range results {
			if status == 503 {
				rejectedCount++
			} else if status == 200 {
				successCount++
			}
		}

		assert.True(t, rejectedCount > 0, "应该有请求被拒绝")
		assert.True(t, successCount > 0, "应该有请求成功")
	})

	t.Run("请求数量限制测试", func(t *testing.T) {
		app := fiber.New()

		config := ConcurrentLimitConfig{
			MaxWorkers:  -1, // 不限制工作协程
			MaxRequests: 3,
		}

		app.Use(NewConcurrentLimiter(config, l))

		app.Get("/test", func(c *fiber.Ctx) error {
			time.Sleep(50 * time.Millisecond)
			return c.SendString("ok")
		})

		// 并发发送多个请求
		var wg sync.WaitGroup
		results := make([]int, 6)

		for i := 0; i < 6; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				req := httptest.NewRequest("GET", "/test", nil)
				resp, err := app.Test(req, 200)
				if err != nil {
					results[index] = 500
				} else {
					results[index] = resp.StatusCode
				}
			}(i)
		}

		wg.Wait()

		// 统计结果
		rejectedCount := 0
		successCount := 0
		for _, status := range results {
			if status == 503 {
				rejectedCount++
			} else if status == 200 {
				successCount++
			}
		}

		assert.True(t, rejectedCount > 0, "应该有请求被拒绝")
		assert.True(t, successCount <= 3, "成功请求数不应超过限制")
	})

	t.Run("不限制测试", func(t *testing.T) {
		app := fiber.New()

		// 设置为-1表示不限制
		config := ConcurrentLimitConfig{
			MaxWorkers:  -1, // 不限制工作协程
			MaxRequests: -1, // 不限制请求数
		}

		app.Use(NewConcurrentLimiter(config, l))

		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})

		// 发送多个请求，都应该成功
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode, "所有请求都应该成功")
		}
	})
}

func TestRateLimiter(t *testing.T) {
	l := logger.New("test")

	t.Run("IP限流测试", func(t *testing.T) {
		app := fiber.New()

		config := RateLimitConfig{
			GlobalMax:         100,
			GlobalExpiration:  time.Minute,
			PerIPMax:          2, // 每个IP只允许2个请求
			PerIPExpiration:   time.Minute,
			PerUserMax:        100,
			PerUserExpiration: time.Minute,
			SkipPaths:         []string{},
			Logger:            l,
		}

		app.Use(NewRateLimiter(config))

		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})

		// 从同一IP发送多个请求
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.1") // 模拟相同IP

			resp, err := app.Test(req)
			require.NoError(t, err)

			if i < 2 {
				assert.Equal(t, 200, resp.StatusCode, "前两个请求应该成功")
			} else {
				assert.Equal(t, 429, resp.StatusCode, "后续请求应该被限流")
			}
		}
	})

	t.Run("跳过路径测试", func(t *testing.T) {
		app := fiber.New()

		config := RateLimitConfig{
			GlobalMax:         1, // 全局只允许1个请求
			GlobalExpiration:  time.Minute,
			PerIPMax:          1,
			PerIPExpiration:   time.Minute,
			PerUserMax:        1,
			PerUserExpiration: time.Minute,
			SkipPaths:         []string{"/health"},
			Logger:            l,
		}

		app.Use(NewRateLimiter(config))

		app.Get("/health", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})

		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})

		// 健康检查路径应该不受限制
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest("GET", "/health", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode, "健康检查路径应该不受限流影响")
		}

		// 普通路径应该受限制
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode, "第一个请求应该成功")

		req = httptest.NewRequest("GET", "/test", nil)
		resp, err = app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 429, resp.StatusCode, "第二个请求应该被限流")
	})

	t.Run("速率限制不限制测试", func(t *testing.T) {
		app := fiber.New()

		// 设置为-1表示不限制
		config := RateLimitConfig{
			GlobalMax:         -1, // 不限制全局请求
			GlobalExpiration:  time.Minute,
			PerIPMax:          -1, // 不限制IP请求
			PerIPExpiration:   time.Minute,
			PerUserMax:        -1, // 不限制用户请求
			PerUserExpiration: time.Minute,
			SkipPaths:         []string{},
			Logger:            l,
		}

		app.Use(NewRateLimiter(config))

		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})

		// 从同一IP发送多个请求，都应该成功
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Forwarded-For", "192.168.1.1")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode, "所有请求都应该成功")
		}
	})
}

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()
	defer store.Close()

	t.Run("基本操作测试", func(t *testing.T) {
		key := "test_key"

		// 初始值应该为0
		hits, err := store.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, 0, hits)

		// 设置值
		err = store.Set(key, 5, time.Second)
		assert.NoError(t, err)

		// 获取值
		hits, err = store.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, 5, hits)
	})

	t.Run("过期测试", func(t *testing.T) {
		key := "expire_test"

		// 设置一个很短的过期时间
		err := store.Set(key, 10, 50*time.Millisecond)
		assert.NoError(t, err)

		// 立即获取应该有值
		hits, err := store.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, 10, hits)

		// 等待过期
		time.Sleep(100 * time.Millisecond)

		// 过期后应该返回0
		hits, err = store.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, 0, hits)
	})

	t.Run("重置测试", func(t *testing.T) {
		// 设置一些值
		store.Set("key1", 1, time.Minute)
		store.Set("key2", 2, time.Minute)

		// 重置
		err := store.Reset()
		assert.NoError(t, err)

		// 所有值都应该被清除
		hits, err := store.Get("key1")
		assert.NoError(t, err)
		assert.Equal(t, 0, hits)

		hits, err = store.Get("key2")
		assert.NoError(t, err)
		assert.Equal(t, 0, hits)
	})
}
