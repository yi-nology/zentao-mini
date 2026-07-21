package zentao

import (
	"context"
	"time"

	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/metrics"

	"github.com/yi-nology/common/biz/zentao"
	"go.uber.org/zap"
)

// GetUsers 获取用户列表（支持分页）
func (c *Client) GetUsers(page, pageSize int) (*zentao.UserListResponse, error) {
	allUsers, err := c.GetUsersAll()
	if err != nil {
		return nil, err
	}

	metrics.RecordCacheHit("users")

	total := len(allUsers)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return &zentao.UserListResponse{
			Users: []zentao.User{},
			Page:  page,
			Total: total,
			Limit: pageSize,
		}, nil
	}
	if end > total {
		end = total
	}

	return &zentao.UserListResponse{
		Users: allUsers[start:end],
		Page:  page,
		Total: total,
		Limit: pageSize,
	}, nil
}

// GetUsersAll 获取所有用户列表
func (c *Client) GetUsersAll() ([]zentao.User, error) {
	if c.IsSessionMode() {
		return c.getUsersAllSession(context.Background())
	}
	cacheKey := "zentao:users:all"

	result, err := GlobalCache.GetOrLoadWithLock(cacheKey, func() (interface{}, error) {
		metrics.RecordCacheMiss("users")
		startTime := time.Now()

		if _, err := c.getToken(); err != nil {
			return nil, err
		}

		var allUsers []zentao.User
		currentPage := 1
		pageSize := 100

		for {
			var response *zentao.UserListResponse
			if err := c.withTokenRetry("GetUsers", func(client *zentao.Client) error {
				var err error
				response, err = client.GetUsers(currentPage, pageSize)
				return err
			}); err != nil {
				metrics.RecordZentaoAPIRequest("users", "GET", time.Since(startTime), err)
				return nil, err
			}

			allUsers = append(allUsers, response.Users...)

			if len(response.Users) < pageSize {
				break
			}

			currentPage++
		}

		duration := time.Since(startTime)
		metrics.RecordCacheOperation("users", "fetch", duration)
		metrics.RecordZentaoAPIRequest("users", "GET", duration, nil)

		logger.Info("Users fetched",
			zap.Int("count", len(allUsers)),
			zap.Duration("duration", duration),
		)

		return allUsers, nil
	}, 10*time.Minute)

	if err != nil {
		return nil, err
	}
	return result.([]zentao.User), nil
}
