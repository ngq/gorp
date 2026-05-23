// Package rpc 提供商品服务的 gRPC 客户端封装。
// trade-service 结算时需要通过此客户端调用 catalog-service 获取商品价格和库存。
package rpc

import (
	"context"

	catalogv1 "nop-go/services/catalog-service/api/catalog/v1"

	"google.golang.org/grpc"
)

// CatalogClient 商品服务 gRPC 客户端。
type CatalogClient struct {
	conn *grpc.ClientConn
	cli  catalogv1.CatalogRPCClient
}

// NewCatalogClient 创建商品服务 gRPC 客户端。
func NewCatalogClient(addr string, opts ...grpc.DialOption) (*CatalogClient, error) {
	conn, err := Dial(addr, opts...)
	if err != nil {
		return nil, err
	}
	return &CatalogClient{conn: conn, cli: catalogv1.NewCatalogRPCClient(conn)}, nil
}

// Close 关闭连接。
func (c *CatalogClient) Close() error {
	return c.conn.Close()
}

// GetProduct 根据 ID 获取商品信息。
func (c *CatalogClient) GetProduct(ctx context.Context, id uint32) (*catalogv1.GetProductResp, error) {
	return c.cli.GetProduct(ctx, &catalogv1.GetProductReq{Id: id})
}

// ListProducts 获取商品列表。
func (c *CatalogClient) ListProducts(ctx context.Context, page, size int32, categoryID, manufacturerID uint32, keyword string) (*catalogv1.ListProductsResp, error) {
	return c.cli.ListProducts(ctx, &catalogv1.ListProductsReq{
		Page:           page,
		Size:           size,
		CategoryId:     categoryID,
		ManufacturerId: manufacturerID,
		Keyword:        keyword,
	})
}