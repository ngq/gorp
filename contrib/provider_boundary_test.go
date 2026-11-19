package contrib

import (
	"testing"

	logzap "github.com/ngq/gorp/contrib/log/zap"
	"github.com/stretchr/testify/require"
)

func TestContribLogBackendAvailability(t *testing.T) {
	logger, err := logzap.New("info", "console")
	require.NoError(t, err)
	require.NotNil(t, logger)
}
