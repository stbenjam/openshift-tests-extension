package extensions

type Registry struct {
	extensions map[string]*Extension
}

func NewExtensionRegistry() *Registry {
	var r Registry
	r.Register("default", DefaultExtension)
	return &r
}

func (r *Registry) Get(name string) *Extension {
	return r.extensions[name]
}

func (r *Registry) Register(name string, extension Extension) {
	if r.extensions == nil {
		r.extensions = make(map[string]*Extension)
	}

	r.extensions[name] = &extension
}

func (r *Registry) Deregister(name string) {
	delete(r.extensions, name)
}
