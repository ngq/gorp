package container

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestContainer_IsBindRecognizesDeferredKeyBeforeLoad(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := &deferredProvider{
		name:   "deferred-user",
		keys:   []string{"user.repo"},
		loaded: &loaded,
		booted: &booted,
		value:  "ok",
	}

	require.NoError(t, c.RegisterProvider(p))
	require.True(t, c.IsBind("user.repo"))
	require.Equal(t, 0, loaded)
	require.Equal(t, 0, booted)

	v, err := c.Make("user.repo")
	require.NoError(t, err)
	require.Equal(t, "ok", v)
	require.Equal(t, 1, loaded)
	require.Equal(t, 1, booted)
}

func TestContainer_MakeLoadsDeferredProviderOnlyOnce(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := &deferredProvider{
		name:   "deferred-once",
		keys:   []string{"cache.client"},
		loaded: &loaded,
		booted: &booted,
		value:  "cache",
	}

	require.NoError(t, c.RegisterProvider(p))

	_, err := c.Make("cache.client")
	require.NoError(t, err)
	_, err = c.Make("cache.client")
	require.NoError(t, err)

	require.Equal(t, 1, loaded)
	require.Equal(t, 1, booted)
}

func TestContainer_MustMakePanicsForUnknownKey(t *testing.T) {
	c := New()
	require.Panics(t, func() {
		_ = c.MustMake("missing.key")
	})
}

func TestContainer_BindNonSingletonCreatesFreshInstance(t *testing.T) {
	c := New()
	count := 0
	c.Bind("transient.counter", func(contract.Container) (any, error) {
		count++
		return count, nil
	}, false)

	v1, err := c.Make("transient.counter")
	require.NoError(t, err)
	v2, err := c.Make("transient.counter")
	require.NoError(t, err)

	require.Equal(t, 1, v1)
	require.Equal(t, 2, v2)
}
