package runtime

import "context"

// AppService 是 framework 面向业务项目的最小应用服务标记接口。
type AppService interface{}

// UnitOfWork 表示轻量事务边界。
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
