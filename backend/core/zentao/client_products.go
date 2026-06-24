package zentao

import (
	"context"
	"time"

	"github.com/yi-nology/zentao-mini/backend/core/logger"
	"github.com/yi-nology/zentao-mini/backend/core/metrics"

	"github.com/yi-nology/common/biz/zentao"
	"go.uber.org/zap"
)

// GetProducts 获取产品列表
func (c *Client) GetProducts() ([]zentao.Product, error) {
	cacheKey := "zentao:products:all"

	result, err := GlobalCache.GetOrLoadWithLock(cacheKey, func() (interface{}, error) {
		metrics.RecordCacheMiss("products")
		start := time.Now()

		if _, err := c.getToken(); err != nil {
			return nil, err
		}

		var firstPageResponse *zentao.ProductListResponse
		if err := c.withTokenRetry("GetProducts", func(client *zentao.Client) error {
			var err error
			firstPageResponse, err = client.GetProducts(1, 100)
			return err
		}); err != nil {
			metrics.RecordZentaoAPIRequest("products", "GET", time.Since(start), err)
			return nil, err
		}

		var allProducts []zentao.Product
		allProducts = append(allProducts, firstPageResponse.Products...)

		total := firstPageResponse.Total
		pageSize := 100
		totalPages := (total + pageSize - 1) / pageSize

		if totalPages > 1 {
			for page := 2; page <= totalPages; page++ {
				var pageResponse *zentao.ProductListResponse
				if err := c.withTokenRetry("GetProducts", func(client *zentao.Client) error {
					var err error
					pageResponse, err = client.GetProducts(page, pageSize)
					return err
				}); err != nil {
					metrics.RecordZentaoAPIRequest("products", "GET", time.Since(start), err)
					return nil, err
				}
				allProducts = append(allProducts, pageResponse.Products...)
			}
		}

		duration := time.Since(start)
		metrics.RecordCacheOperation("products", "fetch", duration)
		metrics.RecordZentaoAPIRequest("products", "GET", duration, nil)

		logger.Info("Products fetched",
			zap.Int("count", len(allProducts)),
			zap.Duration("duration", duration),
		)

		return allProducts, nil
	}, 5*time.Minute)

	if err != nil {
		return nil, err
	}
	metrics.RecordCacheHit("products")
	return result.([]zentao.Product), nil
}

// GetProduct 获取产品详情
func (c *Client) GetProduct(productID int) (*zentao.Product, error) {
	var result *zentao.Product
	err := c.withTokenRetry("GetProduct", func(client *zentao.Client) error {
		var err error
		result, err = client.GetProduct(productID)
		return err
	})
	return result, err
}

// GetProductContext 获取产品详情（支持 context 取消）
func (c *Client) GetProductContext(ctx context.Context, productID int) (*zentao.Product, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	var result *zentao.Product
	err := c.withTokenRetryContext(ctx, "GetProduct", func(client *zentao.Client) error {
		var err error
		result, err = client.GetProduct(productID)
		return err
	})
	return result, err
}
