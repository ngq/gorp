package runtime

import "context"

const (
	RootKey = "framework.root"
	HostKey = "framework.host"
)

type Root interface {
	BasePath() string
	StoragePath() string
	RuntimePath() string
	LogPath() string
	ConfigPath() string
	TempPath() string
}

type Host interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Shutdown(ctx context.Context) error
	RegisterService(name string, service Hostable) error
	Services() []string
}

type Hostable interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Lifecycle interface {
	OnStarting(ctx context.Context) error
	OnStarted(ctx context.Context) error
	OnStopping(ctx context.Context) error
	OnStopped(ctx context.Context) error
}
