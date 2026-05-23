// Package service 定义 content-service 的 gRPC RPC 接口。
// gorp proto from-service 命令通过解析此文件中的接口定义来生成 .proto 文件。
// content-service 当前暂无其他服务通过 gRPC 调用的需求，
// 但为统一架构风格保留接口定义，后续可扩展博客/新闻等内容查询。
package service

import "context"

// ContentRPC 内容服务 gRPC 接口 —— 定义其他服务可能需要跨服务调用的方法。
// 当前仅定义 GetBlog 供其他服务获取博客内容（如 admin 展示关联内容）。
type ContentRPC interface {
	// GetBlog 根据 ID 获取博客内容
	GetBlog(ctx context.Context, req *GetBlogReq) (*GetBlogResp, error)
}

// ======================== gRPC 请求/响应类型 ========================

// GetBlogReq 获取博客请求
type GetBlogReq struct {
	ID uint64 `json:"id" remark:"博客ID"`
}

// GetBlogResp 获取博客响应
type GetBlogResp struct {
	ID        uint64 `json:"id" remark:"博客ID"`
	Title     string `json:"title" remark:"标题"`
	Content   string `json:"content" remark:"正文"`
	Author    string `json:"author" remark:"作者"`
	Status    string `json:"status" remark:"状态"`
	Tags      string `json:"tags" remark:"标签"`
	CreatedAt string `json:"created_at" remark:"创建时间"`
	UpdatedAt string `json:"updated_at" remark:"更新时间"`
}