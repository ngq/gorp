package integration_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ngq/gorp/framework/contract/integration"
	"github.com/stretchr/testify/require"
)

type OrderCreatedPayload struct {
	OrderID string `json:"order_id"`
	Amount  float64 `json:"amount"`
}

func TestNewCloudEvent_ValidSpec(t *testing.T) {
	ctx := context.WithValue(context.Background(), "trace_id", "trace-9999")
	ctx = context.WithValue(ctx, "request_id", "req-8888")

	payload := OrderCreatedPayload{OrderID: "1001", Amount: 99.9}
	evt := integration.NewCloudEvent(ctx, "com.gorp.order.created", "urn:gorp:service:order", payload)

	require.Equal(t, "1.0", evt.SpecVersion)
	require.NotEmpty(t, evt.ID)
	require.Equal(t, "com.gorp.order.created", evt.Type)
	require.Equal(t, "urn:gorp:service:order", evt.Source)
	require.Equal(t, "application/json", evt.DataContentType)
	require.Equal(t, payload, evt.Data)

	// Check Trace Context extensions
	require.Equal(t, "trace-9999", evt.Extensions["trace_id"])
	require.Equal(t, "req-8888", evt.Extensions["request_id"])

	// Check JSON serialization
	jsonBytes, err := json.Marshal(evt)
	require.NoError(t, err)
	require.Contains(t, string(jsonBytes), `"specversion":"1.0"`)
	require.Contains(t, string(jsonBytes), `"order_id":"1001"`)
}
