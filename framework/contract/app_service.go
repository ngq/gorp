package contract

import "context"

// AppService 是 framework 面向业务项目的最小应用服务标记接口。
//
// 中文说明：
// - 轻量 CRUD / app service 模式不强制厚分层；
// - 对简单业务来说，应用服务只需要表达“一个可被业务调用的动作集合”；
// - 后续复杂业务可以在此基础上继续升级到更丰富的 usecase / aggregate / domain service 结构。
type AppService interface{}

// UnitOfWork 表示轻量事务边界。
//
// 中文说明：
// - 用于简单业务在应用服务层表达“在一个事务中完成一组动作”；
// - 不强制业务项目引入完整 DDD 事务模型；
// - 轻量模式下只暴露最小事务包裹接口。
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
