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

// GetProducts 获取产品列表
func (c *Client) GetProducts() ([]zentao.Product, error) {
	cacheKey := "zentao:products:all"

	metrics.RecordCacheHit("products")
	return withCachedList(cacheKey, 5*time.Minute, func() ([]zentao.Product, error) {
		metrics.RecordCacheMiss("products")
		start := time.Now()

		if _, err := c.getToken(); err != nil {
			return nil, err
		}

		// 原实现按首页 Total 算 totalPages 逐页走；Walk 的短页即末页在
		// 同一接口上等价（末页必短），整除边界多一次空页请求
		const pageSize = 100
		allProducts, err := pagination.Walk(context.Background(), pageSize, func(_ context.Context, page int) ([]zentao.Product, error) {
			var pageResponse *zentao.ProductListResponse
			if err := c.withTokenRetry("GetProducts", func(client *zentao.Client) error {
				var err error
				pageResponse, err = client.GetProducts(page, pageSize)
				return err
			}); err != nil {
				metrics.RecordZentaoAPIRequest("products", "GET", time.Since(start), err)
				return nil, err
			}
			return pageResponse.Products, nil
		})
		if err != nil {
			return nil, err
		}

		duration := time.Since(start)
		metrics.RecordCacheOperation("products", "fetch", duration)
		metrics.RecordZentaoAPIRequest("products", "GET", duration, nil)

		logger.Info("Products fetched",
			zap.Int("count", len(allProducts)),
			zap.Duration("duration", duration),
		)

		return allProducts, nil
	})
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
