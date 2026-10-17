package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

func TestTimeout_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, Timeout(500*time.Millisecond))
	router.GET("/fast", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "fast"})
	})

	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["msg"] != "fast" {
		t.Fatalf("expected msg fast, got %v", body["msg"])
	}
}

func TestTimeout_Exceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, Timeout(50*time.Millisecond))
	router.GET("/slow", func(c *gin.Context) {
		time.Sleep(150 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"msg": "slow"})
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", recorder.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["message"] != "request timeout" {
		t.Fatalf("expected request timeout, got %v", body["message"])
	}
}

func TestTimeout_PanicRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, RecoveryMiddleware(), Timeout(500*time.Millisecond))
	router.GET("/panic", func(c *gin.Context) {
		panic("timeout worker exploded")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 from recovery middleware, got %d", recorder.Code)
	}
}

func TestTimeoutMiddleware_GinForm(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(TimeoutMiddleware(50 * time.Millisecond))
	router.GET("/slow-gin", func(c *gin.Context) {
		time.Sleep(150 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"msg": "done"})
	})
	router.GET("/fast-gin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "fast-done"})
	})

	// Fast request
	reqFast := httptest.NewRequest(http.MethodGet, "/fast-gin", nil)
	recFast := httptest.NewRecorder()
	router.ServeHTTP(recFast, reqFast)
	if recFast.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recFast.Code)
	}

	// Slow request
	reqSlow := httptest.NewRequest(http.MethodGet, "/slow-gin", nil)
	recSlow := httptest.NewRecorder()
	router.ServeHTTP(recSlow, reqSlow)
	if recSlow.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", recSlow.Code)
	}
}

func TestTimeoutMiddlewareWithHandler_Custom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TimeoutMiddlewareWithHandler(50*time.Millisecond, func(c *gin.Context) {
		c.JSON(http.StatusGatewayTimeout, gin.H{"custom": "timeout occurred"})
	}))
	router.GET("/slow-custom", func(c *gin.Context) {
		time.Sleep(150 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"msg": "done"})
	})

	req := httptest.NewRequest(http.MethodGet, "/slow-custom", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", recorder.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["custom"] != "timeout occurred" {
		t.Fatalf("expected custom timeout occurred, got %v", body["custom"])
	}
}

func TestTimeout_ConcurrentRaceSafety(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router,
		func(next transportcontract.Handler) transportcontract.Handler {
			return func(c transportcontract.Context) {
				c.Set("thread_safe_key", "value")
				next(c)
			}
		},
		Timeout(200*time.Millisecond),
	)
	router.GET("/concurrent", func(c *gin.Context) {
		val, _ := c.Get("thread_safe_key")
		if val != "value" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "corrupted context key"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/concurrent", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("concurrent request failed with status %d", rec.Code)
			}
		}()
	}
	wg.Wait()
}

func TestRequestTimeout_And_IsRequestCanceled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TimeoutMiddleware(300 * time.Millisecond))

	var recordedTimeout time.Duration
	var canceled bool
	router.GET("/budget", func(c *gin.Context) {
		recordedTimeout = RequestTimeout(c)
		canceled = IsRequestCanceled(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/budget", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if recordedTimeout <= 0 || recordedTimeout > 300*time.Millisecond {
		t.Fatalf("expected budget between 0 and 300ms, got %v", recordedTimeout)
	}
	if canceled {
		t.Fatalf("expected context not to be canceled yet")
	}
}
