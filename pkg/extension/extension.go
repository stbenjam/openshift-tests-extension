package extension

import (
	"fmt"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
	"github.com/openshift-eng/openshift-tests-extension/pkg/version"
)

func NewExtension(product, kind, name string) *Extension {
	return &Extension{
		APIVersion: CurrentExtensionVersion,
		Source: Source{
			Commit:       version.CommitFromGit,
			BuildDate:    version.BuildDate,
			GitTreeState: version.GitTreeState,
		},
		Component: Component{
			Product: product,
			Kind:    kind,
			Name:    name,
		},
	}
}

func (e *Extension) AddSuite(suite Suite) *Extension {
	if e.Suites == nil {
		e.Suites = []Suite{suite}
	} else {
		e.Suites = append(e.Suites, suite)
	}

	return e
}

func (e *Extension) FindSpecByName(name string) (*extensiontests.ExtensionTestSpec, error) {
	for i := range e.specs {
		if e.specs[i].Name == name {
			return e.specs[i], nil
		}
	}

	return nil, fmt.Errorf("spec not found: %s", name)
}
