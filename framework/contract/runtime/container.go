package runtime

const ContainerKey = "framework.container"

type Container interface {
	Bind(key string, factory Factory, singleton bool)
	IsBind(key string) bool
	Make(key string) (any, error)
	MustMake(key string) any
	RegisterProvider(p ServiceProvider) error
	RegisterProviders(providers ...ServiceProvider) error
}

type Factory func(Container) (any, error)
