package extensions

import (
	"github.com/openshift-eng/openshift-tests-extension/pkg/api"
	"github.com/openshift-eng/openshift-tests-extension/pkg/extensions/standard"
)

const DefaultExtension = "standard"

type Registry struct {
	extensions map[string]*api.Extension
}

func NewExtensionRegistry() *Registry {
	var r Registry
	r.Register(DefaultExtension, standard.Extension)
	return &r
}

func (r *Registry) Get(name string) *api.Extension {
	return r.extensions[name]
}

func (r *Registry) Register(name string, extension api.Extension) {
	if r.extensions == nil {
		r.extensions = make(map[string]*api.Extension)
	}

	r.extensions[name] = &extension
}

func (r *Registry) Deregister(name string) {
	delete(r.extensions, name)
}
