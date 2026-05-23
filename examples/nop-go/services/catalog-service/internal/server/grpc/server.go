// Package grpc 实现 catalog-service 的 gRPC 服务端。
// CatalogRPCServer 实现 catalogv1.UnimplementedCatalogRPCServer 接口，
// 将 gRPC 请求委托给已有的 ProductService 和 DirectoryService 处理。
package grpc

import (
	"context"

	catalogv1 "nop-go/services/catalog-service/api/catalog/v1"
	"nop-go/services/catalog-service/internal/service"

	"google.golang.org/grpc"
)

// CatalogRPCServer 商品服务 gRPC 服务端实现。
type CatalogRPCServer struct {
	catalogv1.UnimplementedCatalogRPCServer
	productSvc   *service.ProductService
	directorySvc *service.DirectoryService
}

// NewCatalogRPCServer 创建商品服务 gRPC 服务端。
func NewCatalogRPCServer(productSvc *service.ProductService, directorySvc *service.DirectoryService) *CatalogRPCServer {
	return &CatalogRPCServer{productSvc: productSvc, directorySvc: directorySvc}
}

// GetProduct 根据 ID 获取商品信息 —— 供 trade-service 结算和 content-service 展示调用。
func (s *CatalogRPCServer) GetProduct(ctx context.Context, req *catalogv1.GetProductReq) (*catalogv1.GetProductResp, error) {
	product, err := s.productSvc.GetByID(ctx, uint(req.GetId()))
	if err != nil {
		return nil, err
	}
	return &catalogv1.GetProductResp{
		Id:              uint32(product.ID),
		Name:            product.Name,
		ShortDesc:       product.ShortDesc,
		FullDesc:        product.FullDesc,
		Sku:             product.SKU,
		Price:           product.Price,
		OldPrice:        product.OldPrice,
		Cost:            product.Cost,
		Stock:           int32(product.Stock),
		CategoryId:      uint32(product.CategoryID),
		ManufacturerId:  uint32(product.ManufacturerID),
		IsPublished:     product.IsPublished,
		CreatedAt:       product.CreatedAt,
		UpdatedAt:       product.UpdatedAt,
	}, nil
}

// ListProducts 获取商品列表 —— 供 trade-service 批量验证库存、content-service 展示列表。
func (s *CatalogRPCServer) ListProducts(ctx context.Context, req *catalogv1.ListProductsReq) (*catalogv1.ListProductsResp, error) {
	products, total, err := s.productSvc.List(ctx, int(req.GetPage()), int(req.GetSize()),
		uint(req.GetCategoryId()), uint(req.GetManufacturerId()), req.GetKeyword())
	if err != nil {
		return nil, err
	}
	items := make([]*catalogv1.GetProductResp, 0, len(products))
	for _, p := range products {
		items = append(items, &catalogv1.GetProductResp{
			Id:              uint32(p.ID),
			Name:            p.Name,
			ShortDesc:       p.ShortDesc,
			FullDesc:        p.FullDesc,
			Sku:             p.SKU,
			Price:           p.Price,
			OldPrice:        p.OldPrice,
			Cost:            p.Cost,
			Stock:           int32(p.Stock),
			CategoryId:      uint32(p.CategoryID),
			ManufacturerId:  uint32(p.ManufacturerID),
			IsPublished:     p.IsPublished,
			CreatedAt:       p.CreatedAt,
			UpdatedAt:       p.UpdatedAt,
		})
	}
	return &catalogv1.ListProductsResp{Items: items, Total: total}, nil
}

// GetCategory 根据 ID 获取分类信息 —— 供 content-service 商品分类展示。
func (s *CatalogRPCServer) GetCategory(ctx context.Context, req *catalogv1.GetCategoryReq) (*catalogv1.GetCategoryResp, error) {
	category, err := s.productSvc.GetCategoryByID(ctx, uint(req.GetId()))
	if err != nil {
		return nil, err
	}
	return &catalogv1.GetCategoryResp{
		Id:          uint32(category.ID),
		Name:        category.Name,
		Description: category.Description,
		ParentId:    uint32(category.ParentID),
		SortOrder:   int32(category.SortOrder),
		IsPublished: category.IsPublished,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}, nil
}

// ListCountries 获取国家列表 —— 供 trade-service 结算时选择配送国家。
func (s *CatalogRPCServer) ListCountries(ctx context.Context, req *catalogv1.ListCountriesReq) (*catalogv1.ListCountriesResp, error) {
	countries, total, err := s.directorySvc.ListCountries(ctx, int(req.GetPage()), int(req.GetSize()))
	if err != nil {
		return nil, err
	}
	items := make([]*catalogv1.CountryItem, 0, len(countries))
	for _, c := range countries {
		items = append(items, &catalogv1.CountryItem{
			Id:               uint32(c.ID),
			Name:             c.Name,
			IsoCode2:         c.IsoCode2,
			IsoCode3:         c.IsoCode3,
			AddressFormat:    c.AddressFormat,
			PostcodeRequired: c.PostcodeRequired,
			CreatedAt:        c.CreatedAt,
			UpdatedAt:        c.UpdatedAt,
		})
	}
	return &catalogv1.ListCountriesResp{Items: items, Total: total}, nil
}

// RegisterCatalogService 注册商品 gRPC 服务到 gRPC Server。
func RegisterCatalogService(server *grpc.Server, productSvc *service.ProductService, directorySvc *service.DirectoryService) {
	catalogv1.RegisterCatalogRPCServer(server, NewCatalogRPCServer(productSvc, directorySvc))
}