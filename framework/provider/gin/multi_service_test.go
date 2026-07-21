package gin

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ngq/gorp/framework/container"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/ngq/gorp/framework/http/serverconfig"
	"github.com/ngq/gorp/framework/provider/host"
)

type testLogger struct{}

func (testLogger) Debug(string, ...observabilitycontract.Field) {}
func (testLogger) Info(string, ...observabilitycontract.Field)  {}
func (testLogger) Warn(string, ...observabilitycontract.Field)  {}
func (testLogger) Error(string, ...observabilitycontract.Field) {}
func (testLogger) With(...observabilitycontract.Field) observabilitycontract.Logger {
	return testLogger{}
}

func newTestService(t *testing.T, name, addr string) *service {
	t.Helper()
	c := container.New()
	c.Bind(observabilitycontract.LogKey, func(runtimecontract.Container) (any, error) {
		return testLogger{}, nil
	}, true)
	t.Cleanup(func() { _ = c.Destroy() })
	return newService(c, serverconfig.Service{Name: name, Enabled: true, Addr: addr}).(*service)
}

func TestHTTPServiceExposesGinStyleRoutingAndIsolatesRouteTrees(t *testing.T) {
	admin := newTestService(t, "admin", ":0")
	wecom := newTestService(t, "wecom", ":0")

	v1 := admin.Group("/api/v1")
	v1.GET("/users", func(c transportcontract.Context) { c.Status(http.StatusNoContent) })
	wecom.POST("/callback", func(c transportcontract.Context) { c.Status(http.StatusAccepted) })

	adminEngine, _ := NativeEngine(admin)
	wecomEngine, _ := NativeEngine(wecom)

	assertStatus(t, adminEngine, http.MethodGet, "/api/v1/users", http.StatusNoContent)
	assertStatus(t, adminEngine, http.MethodPost, "/callback", http.StatusNotFound)
	assertStatus(t, wecomEngine, http.MethodPost, "/callback", http.StatusAccepted)
	assertStatus(t, wecomEngine, http.MethodGet, "/api/v1/users", http.StatusNotFound)
}

func TestEveryHTTPServiceHasRecoveryAndIndependentMiddleware(t *testing.T) {
	admin := newTestService(t, "admin", ":0")
	wecom := newTestService(t, "wecom", ":0")
	admin.Use(func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			c.SetHeader("X-Service", "admin")
			next(c)
		}
	})
	admin.GET("/panic", func(transportcontract.Context) { panic("admin panic") })
	wecom.GET("/panic", func(transportcontract.Context) { panic("wecom panic") })

	adminEngine, _ := NativeEngine(admin)
	wecomEngine, _ := NativeEngine(wecom)
	adminRecorder := httptest.NewRecorder()
	adminEngine.ServeHTTP(adminRecorder, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if adminRecorder.Code != http.StatusInternalServerError || adminRecorder.Header().Get("X-Service") != "admin" {
		t.Fatalf("admin middleware/recovery mismatch: status=%d headers=%v", adminRecorder.Code, adminRecorder.Header())
	}
	wecomRecorder := httptest.NewRecorder()
	wecomEngine.ServeHTTP(wecomRecorder, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if wecomRecorder.Code != http.StatusInternalServerError || wecomRecorder.Header().Get("X-Service") != "" {
		t.Fatalf("wecom middleware/recovery mismatch: status=%d headers=%v", wecomRecorder.Code, wecomRecorder.Header())
	}
}

func TestRegistrySupportsSharedAndServiceSpecificMiddleware(t *testing.T) {
	admin := newTestService(t, "admin", ":0")
	wecom := newTestService(t, "wecom", ":0")
	registry := newHTTPRegistry([]transportcontract.HTTPServiceEntry{
		{Name: "admin", Service: admin, Enabled: true},
		{Name: "wecom", Service: wecom, Enabled: true},
	})
	registry.Use(func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			c.SetHeader("X-Shared", "true")
			next(c)
		}
	})
	admin.Use(func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			c.SetHeader("X-Admin", "true")
			next(c)
		}
	})
	admin.GET("/ping", func(c transportcontract.Context) { c.Status(http.StatusNoContent) })
	wecom.GET("/ping", func(c transportcontract.Context) { c.Status(http.StatusNoContent) })

	adminEngine, _ := NativeEngine(admin)
	wecomEngine, _ := NativeEngine(wecom)
	adminRecorder := httptest.NewRecorder()
	adminEngine.ServeHTTP(adminRecorder, httptest.NewRequest(http.MethodGet, "/ping", nil))
	wecomRecorder := httptest.NewRecorder()
	wecomEngine.ServeHTTP(wecomRecorder, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if adminRecorder.Header().Get("X-Shared") != "true" || wecomRecorder.Header().Get("X-Shared") != "true" {
		t.Fatal("shared middleware must run on every HTTP service")
	}
	if adminRecorder.Header().Get("X-Admin") != "true" || wecomRecorder.Header().Get("X-Admin") != "" {
		t.Fatal("service-specific middleware must remain isolated")
	}
}

func TestTwoHTTPServicesListenAtTheSameTime(t *testing.T) {
	admin := newTestService(t, "admin", "127.0.0.1:0")
	wecom := newTestService(t, "wecom", "127.0.0.1:0")
	admin.GET("/admin", func(c transportcontract.Context) { c.Status(http.StatusNoContent) })
	wecom.GET("/wecom", func(c transportcontract.Context) { c.Status(http.StatusAccepted) })
	if err := admin.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Stop(context.Background()) })
	if err := wecom.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = wecom.Stop(context.Background()) })

	adminAddr := "http://" + admin.listener.Addr().String()
	wecomAddr := "http://" + wecom.listener.Addr().String()
	assertRemoteStatus(t, adminAddr+"/admin", http.StatusNoContent)
	assertRemoteStatus(t, adminAddr+"/wecom", http.StatusNotFound)
	assertRemoteStatus(t, wecomAddr+"/wecom", http.StatusAccepted)
	assertRemoteStatus(t, wecomAddr+"/admin", http.StatusNotFound)
}

func TestHTTPServiceStartReturnsPortConflictAndHostRollsBack(t *testing.T) {
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := probe.Addr().String()
	_ = probe.Close()

	first := newTestService(t, "admin", addr)
	second := newTestService(t, "wecom", addr)
	h := host.NewDefaultHost(nil)
	_ = h.RegisterService("http.admin", host.NewHTTPService("http.admin", first))
	_ = h.RegisterService("http.wecom", host.NewHTTPService("http.wecom", second))

	if err := h.Start(context.Background()); err == nil {
		t.Fatal("expected second service to fail binding the occupied address")
	}
	if first.listener != nil {
		t.Fatal("first service listener must be closed during startup rollback")
	}
}

func assertStatus(t *testing.T, handler http.Handler, method, path string, want int) {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(method, path, nil))
	if recorder.Code != want {
		t.Fatalf("%s %s: expected %d, got %d", method, path, want, recorder.Code)
	}
}

func assertRemoteStatus(t *testing.T, url string, want int) {
	t.Helper()
	response, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != want {
		t.Fatalf("GET %s: expected %d, got %d", url, want, response.StatusCode)
	}
}
