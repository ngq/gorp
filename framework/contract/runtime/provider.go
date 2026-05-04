package runtime

type ServiceProvider interface {
	Name() string
	Register(c Container) error
	Boot(c Container) error
	IsDefer() bool
	Provides() []string
}
