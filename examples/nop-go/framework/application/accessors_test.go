// Package application_test provides unit tests for the application service identity accessors.
//
// 适用场景：
// - 验证 Application Service 从 context 中提取身份、安全主体和租户信息的 accessor 行为。
package application

import (
	"context"
	"testing"

	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

// TestServiceIdentityContextHelpers verifies that service identity can be
// stored in and retrieved from a context using the application helpers.
//
// TestServiceIdentityContextHelpers 验证服务标识可以通过应用程序辅助函数
// 存储到 context 中并从中检索。
func TestServiceIdentityContextHelpers(t *testing.T) {
	identity := &securitycontract.ServiceIdentity{
		ServiceID:   "svc-1",
		ServiceName: "demo",
	}
	ctx := WithServiceIdentity(context.Background(), identity)

	got, ok := FromServiceIdentity(ctx)
	if !ok {
		t.Fatal("expected service identity from context")
	}
	if got.ServiceID != "svc-1" || got.ServiceName != "demo" {
		t.Fatalf("unexpected service identity: %#v", got)
	}
}
