package grpc_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ngq/gorp/framework"
	bootstrapgrpc "github.com/ngq/gorp/framework/bootstrap/grpc"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	"github.com/stretchr/testify/require"
)

func TestNewGRPCServiceRuntime(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "app.yaml"), []byte("app:\n  name: test-grpc-service\n"), 0644))

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer func() { _ = os.Chdir(origDir) }()

	opts := bootstrapgrpc.GRPCServiceOptions{
		GovernanceMode: "mono",
	}
	rt, err := bootstrapgrpc.NewGRPCServiceRuntime("test-grpc-service", opts)
	require.NoError(t, err)
	require.NotNil(t, rt)
	require.Equal(t, "test-grpc-service", rt.ServiceName)
	require.NotNil(t, rt.App)
	require.NotNil(t, rt.Container)
	require.NotNil(t, rt.Logger)
	require.NotNil(t, rt.Config)
}

func TestStartGRPCServerWhenNotBound(t *testing.T) {
	app := framework.NewApplication()
	c := app.Container()
	logger := rtLoggerStub{}

	srv, err := bootstrapgrpc.StartGRPCServer(c, logger)
	require.NoError(t, err)
	require.Nil(t, srv)
}

type rtLoggerStub struct{}

func (rtLoggerStub) Debug(msg string, fields ...observabilitycontract.Field) {}
func (rtLoggerStub) Info(msg string, fields ...observabilitycontract.Field)  {}
func (rtLoggerStub) Warn(msg string, fields ...observabilitycontract.Field)  {}
func (rtLoggerStub) Error(msg string, fields ...observabilitycontract.Field) {}
func (rtLoggerStub) With(fields ...observabilitycontract.Field) observabilitycontract.Logger {
	return rtLoggerStub{}
}
