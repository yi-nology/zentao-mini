package zentao

import (
	"context"
	"time"

	"github.com/oekit/pkg/pagination"
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
	cacheKey := "zentao:users:all"

	return withCachedList(cacheKey, 10*time.Minute, func() ([]zentao.User, error) {
		metrics.RecordCacheMiss("users")
		startTime := time.Now()

		if _, err := c.getToken(); err != nil {
			return nil, err
		}

		allUsers, err := pagination.Walk(context.Background(), 100, func(_ context.Context, page int) ([]zentao.User, error) {
			var response *zentao.UserListResponse
			if err := c.withTokenRetry("GetUsers", func(client *zentao.Client) error {
				var err error
				response, err = client.GetUsers(page, 100)
				return err
			}); err != nil {
				metrics.RecordZentaoAPIRequest("users", "GET", time.Since(startTime), err)
				return nil, err
			}
			return response.Users, nil
		})
		if err != nil {
			return nil, err
		}

		duration := time.Since(startTime)
		metrics.RecordCacheOperation("users", "fetch", duration)
		metrics.RecordZentaoAPIRequest("users", "GET", duration, nil)

		logger.Info("Users fetched",
			zap.Int("count", len(allUsers)),
			zap.Duration("duration", duration),
		)

		return allUsers, nil
	})
}
