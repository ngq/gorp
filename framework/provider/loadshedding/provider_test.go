package loadshedding

import (
	"context"
	"runtime"
	"sync"
	"testing"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// TestSemaphoreLoadShedder_AllowAndDone 验证基本的 Allow/Done 流程。
func TestSemaphoreLoadShedder_AllowAndDone(t *testing.T) {
	cfg := resiliencecontract.LoadSheddingConfig{
		Enabled:        true,
		MaxConcurrency: 2,
	}
	shedder := newSemaphoreLoadShedder(cfg)

	ctx := context.Background()

	// 前两个请求应该成功
	if err := shedder.Allow(ctx, "test"); err != nil {
		t.Fatalf("expected first Allow to succeed, got %v", err)
	}
	if err := shedder.Allow(ctx, "test"); err != nil {
		t.Fatalf("expected second Allow to succeed, got %v", err)
	}

	// 第三个请求应该被丢弃
	if err := shedder.Allow(ctx, "test"); err == nil {
		t.Fatal("expected third Allow to be shedded, got nil error")
	}

	// 释放一个槽位
	shedder.Done(ctx, "test", nil)

	// 现在应该能再次获取
	if err := shedder.Allow(ctx, "test"); err != nil {
		t.Fatalf("expected Allow after Done to succeed, got %v", err)
	}
	shedder.Done(ctx, "test", nil)
	shedder.Done(ctx, "test", nil)
}

// TestSemaphoreLoadShedder_DefaultMaxConcurrency 验证默认并发数为 GOMAXPROCS * 100。
func TestSemaphoreLoadShedder_DefaultMaxConcurrency(t *testing.T) {
	cfg := resiliencecontract.LoadSheddingConfig{
		Enabled:        true,
		MaxConcurrency: 0, // 使用默认值
	}
	shedder := newSemaphoreLoadShedder(cfg)

	expected := runtime.GOMAXPROCS(0) * 100
	if shedder.defaultMaxCon != expected {
		t.Fatalf("expected defaultMaxCon=%d, got %d", expected, shedder.defaultMaxCon)
	}
}

// TestSemaphoreLoadShedder_ResourcePolicies 验证按资源粒度的独立策略。
func TestSemaphoreLoadShedder_ResourcePolicies(t *testing.T) {
	cfg := resiliencecontract.LoadSheddingConfig{
		Enabled:        true,
		MaxConcurrency: 100,
		ResourcePolicies: map[string]resiliencecontract.LoadSheddingPolicy{
			"limited": {Enabled: true, MaxConcurrency: 1},
		},
	}
	shedder := newSemaphoreLoadShedder(cfg)
	ctx := context.Background()

	// "limited" 资源只有 1 个并发槽位
	if err := shedder.Allow(ctx, "limited"); err != nil {
		t.Fatalf("expected first Allow for limited to succeed, got %v", err)
	}
	if err := shedder.Allow(ctx, "limited"); err == nil {
		t.Fatal("expected second Allow for limited to be shedded, got nil error")
	}
	shedder.Done(ctx, "limited", nil)

	// "default" 资源使用默认并发数 100
	if err := shedder.Allow(ctx, "default"); err != nil {
		t.Fatalf("expected Allow for default resource to succeed, got %v", err)
	}
	shedder.Done(ctx, "default", nil)
}

// TestSemaphoreLoadShedder_DifferentResourcesIndependent 验证不同资源的信号量独立。
func TestSemaphoreLoadShedder_DifferentResourcesIndependent(t *testing.T) {
	cfg := resiliencecontract.LoadSheddingConfig{
		Enabled:        true,
		MaxConcurrency: 1,
	}
	shedder := newSemaphoreLoadShedder(cfg)
	ctx := context.Background()

	// 占用 resource-a 的唯一槽位
	if err := shedder.Allow(ctx, "resource-a"); err != nil {
		t.Fatalf("expected Allow for resource-a to succeed, got %v", err)
	}

	// resource-b 应该有自己的独立信号量
	if err := shedder.Allow(ctx, "resource-b"); err != nil {
		t.Fatalf("expected Allow for resource-b to succeed (independent semaphore), got %v", err)
	}

	// resource-a 已满，再请求应被拒绝
	if err := shedder.Allow(ctx, "resource-a"); err == nil {
		t.Fatal("expected second Allow for resource-a to be shedded")
	}

	shedder.Done(ctx, "resource-a", nil)
	shedder.Done(ctx, "resource-b", nil)
}

// TestSemaphoreLoadShedder_ConcurrentAccess 验证并发安全性。
func TestSemaphoreLoadShedder_ConcurrentAccess(t *testing.T) {
	cfg := resiliencecontract.LoadSheddingConfig{
		Enabled:        true,
		MaxConcurrency: 10,
	}
	shedder := newSemaphoreLoadShedder(cfg)
	ctx := context.Background()

	var wg sync.WaitGroup
	successCount := 0
	shedCount := 0
	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := shedder.Allow(ctx, "concurrent"); err != nil {
				mu.Lock()
				shedCount++
				mu.Unlock()
				return
			}
			mu.Lock()
			successCount++
			mu.Unlock()
			shedder.Done(ctx, "concurrent", nil)
		}()
	}
	wg.Wait()

	// 所有请求要么成功要么被丢弃，总数应该等于 100
	if successCount+shedCount != 100 {
		t.Fatalf("expected success+shed=100, got success=%d shed=%d", successCount, shedCount)
	}

	// 至少有一些成功的（10 个槽位，100 个并发请求）
	if successCount == 0 {
		t.Fatal("expected at least some successful acquires")
	}
}

// TestProvider_Interface 验证 Provider 实现了 ServiceProvider 接口。
func TestProvider_Interface(t *testing.T) {
	p := NewProvider()
	if p.Name() != "loadshedding.semaphore" {
		t.Fatalf("expected name loadshedding.semaphore, got %s", p.Name())
	}
	if !p.IsDefer() {
		t.Fatal("expected IsDefer to be true")
	}
	provides := p.Provides()
	if len(provides) != 1 || provides[0] != resiliencecontract.LoadShedderKey {
		t.Fatalf("expected Provides to return [%s], got %v", resiliencecontract.LoadShedderKey, provides)
	}
}

// TestErrLoadShedded 验证过载保护错误可被正确识别。
func TestErrLoadShedded(t *testing.T) {
	if ErrLoadShedded == nil {
		t.Fatal("expected ErrLoadShedded to be non-nil")
	}
	// 应该是一个 resilience AppError
	appErr := resiliencecontract.FromError(ErrLoadShedded)
	if appErr == nil {
		t.Fatal("expected ErrLoadShedded to be an AppError")
	}
	if appErr.GetStatus().Code != resiliencecontract.ErrorCodeServiceUnavailable {
		t.Fatalf("expected code %d, got %d", resiliencecontract.ErrorCodeServiceUnavailable, appErr.GetStatus().Code)
	}
}
