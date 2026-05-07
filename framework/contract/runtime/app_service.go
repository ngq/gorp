// Application scenarios:
// - Define minimal business-facing runtime contracts shared across framework layers.
// - Expose the smallest unit-of-work abstraction without binding to a concrete ORM or transaction engine.
// - Keep application service semantics lightweight and portable.
//
// 适用场景：
// - 定义框架各层共享的最小业务运行时契约。
// - 在不绑定具体 ORM 或事务引擎的情况下暴露最小工作单元抽象。
// - 让应用服务语义保持轻量和可移植。
package runtime

import "context"

// AppService marks a framework-level business service.
//
// AppService 标记框架层业务服务。
type AppService interface{}

// UnitOfWork defines a lightweight transactional boundary.
//
// UnitOfWork 定义轻量事务边界。
type UnitOfWork interface {
	// Do executes the callback within the unit-of-work boundary.
	//
	// Do 在工作单元边界内执行回调。
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
