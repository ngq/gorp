package dtmsdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDTMClient_SAGA(t *testing.T) {
	var submitBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/dtmsvr/newGid":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"gid":"gid-123"}`))
		case "/api/dtmsvr/submit":
			require.Equal(t, http.MethodPost, r.Method)
			require.NoError(t, json.NewDecoder(r.Body).Decode(&submitBody))
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"dtm_result":"SUCCESS"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	cfg := &contract.DTMConfig{
		Enabled:       true,
		Endpoint:      server.URL,
		Timeout:       3,
		RetryCount:    4,
		RetryInterval: 7,
	}
	client, err := NewDTMClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	saga := client.SAGA("test-saga")
	require.NotNil(t, saga)
	saga.Add("/api/action", "/api/compensate", map[string]string{"key": "value"})
	saga.Add("/api/action2", "/api/compensate2", nil)

	err = saga.Submit(context.Background())
	require.NoError(t, err)
	require.NotNil(t, submitBody)
	assert.Equal(t, "gid-123", submitBody["gid"])
	assert.Equal(t, "saga", submitBody["trans_type"])
	assert.Equal(t, float64(7), submitBody["retry_interval"])
	assert.Equal(t, float64(4), submitBody["retry_count"])

	steps, ok := submitBody["steps"].([]any)
	require.True(t, ok)
	require.Len(t, steps, 2)

	payloads, ok := submitBody["payloads"].([]any)
	require.True(t, ok)
	require.Len(t, payloads, 2)
	assert.Equal(t, `{"key":"value"}`, payloads[0])
	assert.Equal(t, `null`, payloads[1])

	tx, err := saga.Build()
	require.NoError(t, err)
	require.NotNil(t, tx)
	assert.Equal(t, "gid-123", tx.GID)
	assert.Len(t, tx.Steps, 2)
}

func TestDTMClient_SAGA_SubmitRequiresSteps(t *testing.T) {
	cfg := &contract.DTMConfig{Enabled: true, Endpoint: "http://localhost:36789"}
	client, err := NewDTMClient(cfg)
	require.NoError(t, err)

	err = client.SAGA("empty").Submit(context.Background())
	assert.EqualError(t, err, "dtm: saga has no steps")
}

func TestDTMClient_TCC(t *testing.T) {
	cfg := &contract.DTMConfig{
		Enabled:  true,
		Endpoint: "http://localhost:36789",
	}
	client, err := NewDTMClient(cfg)
	assert.NoError(t, err)

	tcc := client.TCC("test-tcc")
	assert.NotNil(t, tcc)
	tcc.Add("/api/try", "/api/confirm", "/api/cancel", nil)

	err = tcc.Submit(context.Background())
	assert.ErrorIs(t, err, ErrDTMSDKNotImported)
}

func TestDTMClient_XA(t *testing.T) {
	cfg := &contract.DTMConfig{
		Enabled:  true,
		Endpoint: "http://localhost:36789",
	}
	client, err := NewDTMClient(cfg)
	assert.NoError(t, err)

	xa := client.XA("test-xa")
	assert.NotNil(t, xa)
	xa.Add("/api/xa-action", nil)

	err = xa.Submit(context.Background())
	assert.ErrorIs(t, err, ErrDTMSDKNotImported)
}

func TestDTMClient_Barrier(t *testing.T) {
	cfg := &contract.DTMConfig{
		Enabled:  true,
		Endpoint: "http://localhost:36789",
	}
	client, err := NewDTMClient(cfg)
	assert.NoError(t, err)

	barrier := client.Barrier("saga", "test-gid")
	assert.NotNil(t, barrier)

	executed := false
	err = barrier.Call(context.Background(), func(db any) error {
		executed = true
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestDTMClient_Query(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/dtmsvr/query", r.URL.Path)
		require.Equal(t, "test-gid", r.URL.Query().Get("gid"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"gid":"test-gid",
			"status":"submitted",
			"trans_type":"saga",
			"create_time":1710000000,
			"update_time":1710000001,
			"steps":[
				{"branch_id":"01","status":"prepared","op":"action","url":"/api/order/create"}
			]
		}`))
	}))
	defer server.Close()

	cfg := &contract.DTMConfig{
		Enabled:  true,
		Endpoint: server.URL,
	}
	client, err := NewDTMClient(cfg)
	require.NoError(t, err)

	info, err := client.Query(context.Background(), "test-gid")
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "test-gid", info.GID)
	assert.Equal(t, "submitted", info.Status)
	assert.Equal(t, "saga", info.TransactionType)
	assert.EqualValues(t, 1710000000, info.CreateTime)
	assert.EqualValues(t, 1710000001, info.UpdateTime)
	require.Len(t, info.Steps, 1)
	assert.Equal(t, "01", info.Steps[0].BranchID)
	assert.Equal(t, "prepared", info.Steps[0].Status)
	assert.Equal(t, "action", info.Steps[0].Op)
	assert.Equal(t, "/api/order/create", info.Steps[0].URL)
}

func TestDTMClient_QueryRequiresGID(t *testing.T) {
	cfg := &contract.DTMConfig{Enabled: true, Endpoint: "http://localhost:36789"}
	client, err := NewDTMClient(cfg)
	require.NoError(t, err)

	info, err := client.Query(context.Background(), "")
	assert.Nil(t, info)
	assert.EqualError(t, err, "dtm: gid is required")
}

func TestSAGABuilder_AddBranch(t *testing.T) {
	cfg := &contract.DTMConfig{Enabled: true}
	client, _ := NewDTMClient(cfg)

	saga := client.SAGA("test")
	saga.AddBranch("/action", "/compensate", nil, contract.BranchOptions{
		RetryCount:    3,
		RetryInterval: 5,
		Timeout:       9,
	})

	tx, err := saga.Build()
	assert.NoError(t, err)
	assert.Len(t, tx.Steps, 1)
	assert.EqualValues(t, 3, tx.Steps[0].RetryCount)
	assert.EqualValues(t, 5, tx.Steps[0].RetryInterval)
	assert.EqualValues(t, 9, tx.Steps[0].Timeout)
}

func TestDTMClient_APIBaseURL(t *testing.T) {
	cases := []struct {
		name     string
		endpoint string
		expected string
	}{
		{name: "plain endpoint", endpoint: "http://localhost:36789", expected: "http://localhost:36789/api/dtmsvr"},
		{name: "already api base", endpoint: "http://localhost:36789/api/dtmsvr", expected: "http://localhost:36789/api/dtmsvr"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewDTMClient(&contract.DTMConfig{Endpoint: tc.endpoint, Timeout: 1})
			require.NoError(t, err)
			assert.Equal(t, tc.expected, client.apiBaseURL())
		})
	}
}

func TestProvider_Name(t *testing.T) {
	p := NewProvider()
	assert.Equal(t, "dtm.sdk", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{contract.DTMKey}, p.Provides())
}

func TestErrDTMSDKNotImported(t *testing.T) {
	assert.Contains(t, ErrDTMSDKNotImported.Error(), "official SDK")
	assert.Contains(t, ErrDTMSDKNotImported.Error(), "lightweight framework adapter")
}
