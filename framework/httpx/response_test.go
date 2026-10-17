package httpx_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	"github.com/ngq/gorp/framework/httpx"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestResponse_WithGinContext_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	httpx.OK(c, map[string]string{"user": "alice"})

	require.Equal(t, http.StatusOK, w.Code)
	var resp httpx.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, httpx.CodeOK, resp.Code)
	require.Equal(t, httpx.CodeSuccess, resp.Code)
	require.Equal(t, "success", resp.Message)
	require.Empty(t, resp.Reason)
}

func TestResponse_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	errObj := httpx.NotFound("lesson.not_found", "课程不存在")
	httpx.Error(c, errObj)

	require.Equal(t, http.StatusNotFound, w.Code)
	var resp httpx.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, httpx.CodeNotFound, resp.Code)
	require.Equal(t, httpx.Code(5), resp.Code)
	require.Equal(t, "lesson.not_found", resp.Reason)
	require.Equal(t, "课程不存在", resp.Message)
}

func TestResponse_ResourceExhausted(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	errObj := httpx.ResourceExhausted("payment.balance", "字数不足")
	httpx.Error(c, errObj)

	require.Equal(t, http.StatusTooManyRequests, w.Code)
	var resp httpx.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, httpx.CodeResourceExhausted, resp.Code)
	require.Equal(t, httpx.Code(8), resp.Code)
	require.Equal(t, "payment.balance", resp.Reason)
	require.Equal(t, "字数不足", resp.Message)
}

func TestResponse_FailedPrecondition(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	errObj := httpx.FailedPrecondition("user.phone_unbound", "请先绑定手机")
	httpx.Error(c, errObj)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var resp httpx.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, httpx.CodeFailedPrecondition, resp.Code)
	require.Equal(t, httpx.Code(9), resp.Code)
	require.Equal(t, "user.phone_unbound", resp.Reason)
	require.Equal(t, "请先绑定手机", resp.Message)
}

func TestResponse_UnauthenticatedAndPermissionDenied(t *testing.T) {
	// Unauthenticated -> 401, Code 16
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	httpx.Error(c1, httpx.Unauthenticated("auth.missing", "未登录"))
	require.Equal(t, http.StatusUnauthorized, w1.Code)
	var resp1 httpx.Response
	require.NoError(t, json.Unmarshal(w1.Body.Bytes(), &resp1))
	require.Equal(t, httpx.CodeUnauthenticated, resp1.Code)
	require.Equal(t, httpx.Code(16), resp1.Code)
	require.Equal(t, "auth.missing", resp1.Reason)

	// PermissionDenied -> 403, Code 7
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	httpx.Error(c2, httpx.PermissionDenied("auth.forbidden", "无权访问"))
	require.Equal(t, http.StatusForbidden, w2.Code)
	var resp2 httpx.Response
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &resp2))
	require.Equal(t, httpx.CodePermissionDenied, resp2.Code)
	require.Equal(t, httpx.Code(7), resp2.Code)
	require.Equal(t, "auth.forbidden", resp2.Reason)
}

func TestResponse_WithGenericError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	httpx.Error(c, errors.New("db connection lost"))

	require.Equal(t, http.StatusInternalServerError, w.Code)
	var resp httpx.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, httpx.CodeInternal, resp.Code)
	require.Equal(t, httpx.Code(13), resp.Code)
	require.Equal(t, "db connection lost", resp.Message)
}

func TestResponse_NewBizError_FailFast(t *testing.T) {
	require.PanicsWithValue(t, "httpx: 非法语义码 99，必须使用 Code 常量", func() {
		httpx.NewBizError(httpx.Code(99), "invalid", "message")
	})

	require.PanicsWithValue(t, "httpx: CodeOK 不能构造错误", func() {
		httpx.NewBizError(httpx.CodeOK, "ok", "message")
	})
}

func TestResponse_ProblemJSON(t *testing.T) {
	httpx.UseProblemJSON(true)
	defer httpx.UseProblemJSON(false)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	errObj := httpx.NotFound("lesson.not_found", "课程不存在")
	httpx.Error(c, errObj)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "application/problem+json")

	var prob httpx.Problem
	err := json.Unmarshal(w.Body.Bytes(), &prob)
	require.NoError(t, err)
	require.Equal(t, "lesson.not_found", prob.Type)
	require.Equal(t, "Not Found", prob.Title)
	require.Equal(t, http.StatusNotFound, prob.Status)
	require.Equal(t, "课程不存在", prob.Detail)
	require.Equal(t, "lesson.not_found", prob.Reason)
}

func TestResponse_WithAppError_Fail(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	appErr := resiliencecontract.NewError(http.StatusNotFound, resiliencecontract.ErrorReasonNotFound, "user not found")
	httpx.Fail(c, appErr)

	require.Equal(t, http.StatusNotFound, w.Code)
	var resp httpx.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, httpx.CodeNotFound, resp.Code)
	require.Equal(t, "NOT_FOUND", resp.Reason)
	require.Equal(t, "user not found", resp.Message)
}

func TestResponse_AllCanonicalCodes(t *testing.T) {
	testCases := []struct {
		err        *httpx.BizError
		wantStatus int
		wantCode   httpx.Code
	}{
		{httpx.Cancelled("cancel", "cancelled"), 499, httpx.CodeCancelled},
		{httpx.Unknown("unknown", "unknown error"), 500, httpx.CodeUnknown},
		{httpx.InvalidArgument("arg", "invalid arg"), 400, httpx.CodeInvalidArgument},
		{httpx.DeadlineExceeded("deadline", "timeout"), 504, httpx.CodeDeadlineExceeded},
		{httpx.AlreadyExists("exists", "already exists"), 409, httpx.CodeAlreadyExists},
		{httpx.Aborted("abort", "aborted"), 409, httpx.CodeAborted},
		{httpx.OutOfRange("range", "out of range"), 400, httpx.CodeOutOfRange},
		{httpx.Unimplemented("unimpl", "unimplemented"), 501, httpx.CodeUnimplemented},
		{httpx.Unavailable("unavail", "unavailable"), 503, httpx.CodeUnavailable},
		{httpx.DataLoss("dataloss", "data loss"), 500, httpx.CodeDataLoss},
		{httpx.Internal("internal error"), 500, httpx.CodeInternal},
	}

	for _, tc := range testCases {
		t.Run(tc.err.Reason(), func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			httpx.Error(c, tc.err)
			require.Equal(t, tc.wantStatus, w.Code)
			var resp httpx.Response
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			require.Equal(t, tc.wantCode, resp.Code)
		})
	}
}
