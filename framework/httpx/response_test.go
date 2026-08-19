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

func TestResponse_WithGinContext_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	httpx.OK(c, map[string]string{"user": "alice"})

	require.Equal(t, http.StatusOK, w.Code)
	var resp httpx.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, httpx.CodeSuccess, resp.Code)
	require.Equal(t, "success", resp.Message)
}

func TestResponse_WithAppError_Fail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	appErr := resiliencecontract.NewError(http.StatusNotFound, resiliencecontract.ErrorReasonNotFound, "user not found")
	httpx.Fail(c, appErr)

	require.Equal(t, http.StatusNotFound, w.Code)
	var resp httpx.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.Code)
	require.Equal(t, "user not found", resp.Message)
}

func TestResponse_WithGenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	httpx.Error(c, errors.New("db connection lost"))

	require.Equal(t, http.StatusInternalServerError, w.Code)
	var resp httpx.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, httpx.CodeInternalError, resp.Code)
	require.Equal(t, "db connection lost", resp.Message)
}
