// Application scenarios:
// - Define the framework-wide HTTP responder contract used by middleware and handlers.
// - Separate transport-layer response writing from business handler logic.
// - Keep success/error response shaping replaceable across providers and applications.
//
// 适用场景：
// - 定义中间件和处理器共同使用的框架级 HTTP responder 契约。
// - 将 transport 层响应写出逻辑与业务 handler 解耦。
// - 让成功/失败响应整形能力可以在 provider 和应用之间替换。
package transport

// HTTPResponderKey is the container key for the HTTP responder capability.
//
// HTTPResponderKey 是 HTTP responder 能力的容器键。
const HTTPResponderKey = "framework.http.responder"

// HTTPResponder defines the response-writing contract used by the framework.
//
// HTTPResponder 定义框架使用的响应写出契约。
type HTTPResponder interface {
	Success(Context, any)
	SuccessWithMessage(Context, string, any)
	SuccessWithStatus(Context, int, any)
	Error(Context, error)
	BadRequest(Context, string)
	InternalError(Context, string)
}
