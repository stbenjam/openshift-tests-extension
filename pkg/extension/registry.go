package extension

import "fmt"

const DefaultExtension = "default"

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
		// first extension is default
		r.extensions[DefaultExtension] = extension
	}

	r.extensions[fmt.Sprintf("%s:%s:%s", extension.Component.Product, extension.Component.Kind, extension.Component.Name)] = extension
}

func (r *Registry) Deregister(name string) {
	delete(r.extensions, name)
}
