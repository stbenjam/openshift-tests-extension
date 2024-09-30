package extension

type Registry struct {
	extensions map[string]*Extension
}

func NewRegistry() *Registry {
	var r Registry
	return &r
}

func (r *Registry) Get(name string) *Extension {
	return r.extensions[name]
}

func (r *Registry) Register(extension *Extension) {
	if r.extensions == nil {
		r.extensions = make(map[string]*Extension)
	}

	r.extensions[extension.Component.Name] = extension
}

func (r *Registry) Deregister(name string) {
	delete(r.extensions, name)
}
