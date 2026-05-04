package transport

const HTTPResponderKey = "framework.http.responder"

type HTTPResponder interface {
	Success(HTTPContext, any)
	SuccessWithMessage(HTTPContext, string, any)
	SuccessWithStatus(HTTPContext, int, any)
	Error(HTTPContext, error)
	BadRequest(HTTPContext, string)
	InternalError(HTTPContext, string)
}
