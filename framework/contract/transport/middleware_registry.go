// Application scenarios:
// - Define the middleware registry contract for proto-annotation-driven automatic middleware mounting.
// - Allow users to register named middleware instances (e.g., "auth", "logging") and
//   have the generated route code automatically look them up by name.
// - Decouple proto declaration from middleware implementation: proto says "auth: true",
//   the registry maps "auth" to the actual middleware instance.
//
// 适用场景：
// - 定义 proto 注解驱动的自动中间件挂载所需的注册表契约。
// - 允许用户注册具名中间件实例（如 "auth"、"logging"），
//   生成的路由代码按名称自动查找对应中间件。
// - 解耦 proto 声明与中间件实现：proto 声明 "auth: true"，
//   注册表将 "auth" 映射到实际的中间件实例。
package transport

// MiddlewareRegistryKey is the container key for the middleware registry.
//
// MiddlewareRegistryKey 是中间件注册表的容器键。
const MiddlewareRegistryKey = "framework.http.middleware_registry"

// MiddlewareRegistry defines the contract for named middleware registration and lookup.
// Proto annotations (gorp.auth, gorp.middleware) reference middleware by name;
// the registry maps names to concrete HTTPMiddleware instances.
//
// MiddlewareRegistry 定义具名中间件注册与查找契约。
// Proto 注解（gorp.auth、gorp.middleware）通过名称引用中间件；
// 注册表将名称映射到具体的 HTTPMiddleware 实例。
//
// 中文说明：
// - 用户在应用启动时注册中间件：registry.Register("auth", jwtAuthMiddleware)
// - proto 声明 (gorp.auth) = { required: true }，生成代码调用 registry.Lookup("auth")
// - 注册表解耦了 proto 声明与具体中间件实现，同一套 proto 可在不同项目中使用不同中间件
type MiddlewareRegistry interface {
	// Register adds a named middleware to the registry.
	// If a middleware with the same name already exists, it is replaced.
	// Common names: "auth", "authz", "logging", "ratelimit", "cors".
	//
	// Register 向注册表添加具名中间件。
	// 如果同名中间件已存在，则替换。
	// 常用名称："auth"、"authz"、"logging"、"ratelimit"、"cors"。
	Register(name string, middleware HTTPMiddleware)

	// Lookup retrieves a middleware by name.
	// Returns the middleware and true if found, nil and false otherwise.
	//
	// Lookup 按名称查找中间件。
	// 找到时返回中间件和 true，否则返回 nil 和 false。
	Lookup(name string) (HTTPMiddleware, bool)

	// LookupAll retrieves multiple middleware by names.
	// Returns found middleware in order; silently skips unknown names.
	//
	// LookupAll 按名称列表查找多个中间件。
	// 按顺序返回找到的中间件；静默跳过未知名称。
	LookupAll(names []string) []HTTPMiddleware

	// Names returns all registered middleware names.
	//
	// Names 返回所有已注册的中间件名称。
	Names() []string
}
