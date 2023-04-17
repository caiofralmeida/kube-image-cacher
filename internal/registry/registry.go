package registry

type Registry struct {
	Provider, URL string
}

func New(provider, url string) Registry {
	return Registry{provider, url}
}
