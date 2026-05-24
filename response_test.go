package gorp

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/container"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	httpmiddleware "github.com/ngq/gorp/framework/http/middleware"
	"github.com/stretchr/testify/require"
)

type responseCapture struct {
	status int
	body   any
	header http.Header
}

type responderSpy struct {
	calls []string
}

type customResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Result  any    `json:"result,omitempty"`
}

type customResponder struct{}

func (s *responderSpy) Success(transportcontract.Context, any) {
	s.calls = append(s.calls, "success")
}

func (s *responderSpy) SuccessWithMessage(transportcontract.Context, string, any) {
	s.calls = append(s.calls, "success_with_message")
}

func (s *responderSpy) SuccessWithStatus(transportcontract.Context, int, any) {
	s.calls = append(s.calls, "success_with_status")
}

func (s *responderSpy) Error(transportcontract.Context, error) {
	s.calls = append(s.calls, "error")
}

func (s *responderSpy) BadRequest(transportcontract.Context, string) {
	s.calls = append(s.calls, "bad_request")
}

func (s *responderSpy) InternalError(transportcontract.Context, string) {
	s.calls = append(s.calls, "internal_error")
}

func (customResponder) Success(c transportcontract.Context, data any) {
	c.JSON(http.StatusOK, customResponse{
		Success: true,
		Message: "ok",
		Result:  data,
	})
}

func (customResponder) SuccessWithMessage(c transportcontract.Context, message string, data any) {
	c.JSON(http.StatusOK, customResponse{
		Success: true,
		Message: message,
		Result:  data,
	})
}

func (customResponder) SuccessWithStatus(c transportcontract.Context, status int, data any) {
	c.JSON(status, customResponse{
		Success: true,
		Message: "ok",
		Result:  data,
	})
}

func (customResponder) Error(c transportcontract.Context, err error) {
	c.JSON(http.StatusBadRequest, customResponse{
		Success: false,
		Message: err.Error(),
	})
}

func (customResponder) BadRequest(c transportcontract.Context, message string) {
	c.JSON(http.StatusBadRequest, customResponse{
		Success: false,
		Message: message,
	})
}

func (customResponder) InternalError(c transportcontract.Context, message string) {
	c.JSON(http.StatusInternalServerError, customResponse{
		Success: false,
		Message: message,
	})
}

// testResponseContext implements Context for testing
type testResponseContext struct {
	gin      *gin.Context
	captured *responseCapture
}

func (c *testResponseContext) Deadline() (deadline time.Time, ok bool) {
	return c.gin.Request.Context().Deadline()
}

func (c *testResponseContext) Done() <-chan struct{} {
	return c.gin.Request.Context().Done()
}

func (c *testResponseContext) Err() error {
	return c.gin.Request.Context().Err()
}

func (c *testResponseContext) Value(key any) any {
	return c.gin.Request.Context().Value(key)
}

func (c *testResponseContext) Context() context.Context {
	return c.gin.Request.Context()
}

func (c *testResponseContext) Request() *http.Request {
	return c.gin.Request
}

func (c *testResponseContext) Response() http.ResponseWriter {
	return c.gin.Writer
}

func (c *testResponseContext) Param(key string) string {
	return c.gin.Param(key)
}

func (c *testResponseContext) Query(key string) string {
	return c.gin.Query(key)
}

func (c *testResponseContext) DefaultQuery(key, defaultValue string) string {
	return c.gin.DefaultQuery(key, defaultValue)
}

func (c *testResponseContext) DefaultIntQuery(key string, defaultValue int) int {
	return defaultValue
}

func (c *testResponseContext) Int64Param(key string) (int64, error) {
	return 0, nil
}

func (c *testResponseContext) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	return nil, nil, http.ErrNoCookie
}

func (c *testResponseContext) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	return nil
}

func (c *testResponseContext) GetHeader(key string) string {
	return c.gin.GetHeader(key)
}

func (c *testResponseContext) SetHeader(key, value string) {
	c.gin.Header(key, value)
}

func (c *testResponseContext) Bind(obj any) error {
	return c.gin.ShouldBind(obj)
}

func (c *testResponseContext) BindJSON(obj any) error {
	return c.gin.ShouldBindJSON(obj)
}

func (c *testResponseContext) BindQuery(obj any) error {
	return c.gin.ShouldBindQuery(obj)
}

func (c *testResponseContext) JSON(status int, body any) {
	c.captured.status = status
	c.captured.body = body
}

func (c *testResponseContext) String(status int, body string) {
	c.captured.status = status
	c.captured.body = body
}

func (c *testResponseContext) XML(status int, body any) {
	c.captured.status = status
	c.captured.body = body
}

func (c *testResponseContext) Data(status int, contentType string, body []byte) {
	c.captured.status = status
	c.captured.header.Set("Content-Type", contentType)
	c.captured.body = body
}

func (c *testResponseContext) Redirect(status int, location string) {
	c.captured.status = status
	c.captured.header.Set("Location", location)
}

func (c *testResponseContext) Status(code int) {
	c.captured.status = code
}

func (c *testResponseContext) RoutePath() string {
	return c.gin.FullPath()
}

func (c *testResponseContext) ResponseStatus() int {
	return c.captured.status
}

func (c *testResponseContext) Get(key string) (any, bool) {
	return c.gin.Get(key)
}

func (c *testResponseContext) Set(key string, value any) {
	c.gin.Set(key, value)
}

func (c *testResponseContext) Abort(status int) {
	c.gin.AbortWithStatus(status)
}

func (c *testResponseContext) AbortWithJSON(status int, body any) {
	c.captured.status = status
	c.captured.body = body
	c.gin.Abort()
}

func (c *testResponseContext) IsAborted() bool {
	return c.gin.IsAborted()
}

func (c *testResponseContext) Next() {
	c.gin.Next()
}

func newResponseTestContext(t *testing.T, ctx context.Context) (Context, *responseCapture) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	ginCtx.Request = req

	captured := &responseCapture{header: make(http.Header)}
	testCtx := &testResponseContext{gin: ginCtx, captured: captured}
	return testCtx, captured
}

func TestSuccessFallsBackToDefaultResponder(t *testing.T) {
	ctx, captured := newResponseTestContext(t, context.Background())

	Success(ctx, map[string]any{"ok": true})

	require.Equal(t, http.StatusOK, captured.status)
	resp, ok := captured.body.(httpmiddleware.Response)
	require.True(t, ok)
	require.Equal(t, 0, resp.Code)
	require.Equal(t, "success", resp.Message)
	require.Equal(t, map[string]any{"ok": true}, resp.Data)
}

func TestErrorUsesBusinessResponderFromContainer(t *testing.T) {
	c := container.New()
	spy := &responderSpy{}
	c.Bind(transportcontract.HTTPResponderKey, func(runtimecontract.Container) (any, error) {
		return spy, nil
	}, true)

	ctx, _ := newResponseTestContext(t, supportcontract.NewContainerContext(context.Background(), c))

	Error(ctx, errors.New("boom"))

	require.Equal(t, []string{"error"}, spy.calls)
}

func TestSuccessWithMessageUsesBusinessResponderFromContainer(t *testing.T) {
	c := container.New()
	spy := &responderSpy{}
	c.Bind(transportcontract.HTTPResponderKey, func(runtimecontract.Container) (any, error) {
		return spy, nil
	}, true)

	ctx, _ := newResponseTestContext(t, supportcontract.NewContainerContext(context.Background(), c))

	SuccessWithMessage(ctx, "done", map[string]any{"id": 1})

	require.Equal(t, []string{"success_with_message"}, spy.calls)
}

func TestSuccessCanUseCustomBusinessResponseShape(t *testing.T) {
	c := container.New()
	c.Bind(transportcontract.HTTPResponderKey, func(runtimecontract.Container) (any, error) {
		return customResponder{}, nil
	}, true)

	ctx, captured := newResponseTestContext(t, supportcontract.NewContainerContext(context.Background(), c))

	Success(ctx, map[string]any{"id": 7})

	require.Equal(t, http.StatusOK, captured.status)
	resp, ok := captured.body.(customResponse)
	require.True(t, ok)
	require.True(t, resp.Success)
	require.Equal(t, "ok", resp.Message)
	require.Equal(t, map[string]any{"id": 7}, resp.Result)
}
