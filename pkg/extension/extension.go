package extension

import (
	"fmt"

	et "github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
	"github.com/openshift-eng/openshift-tests-extension/pkg/version"
)

func NewExtension(product, kind, name string) *Extension {
	return &Extension{
		APIVersion: CurrentExtensionAPIVersion,
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

func (e *Extension) GetSuite(name string) (*Suite, error) {
	var suite *Suite

	for _, s := range e.Suites {
		if s.Name == name {
			suite = &s
			break
		}
	}

	if suite == nil {
		return nil, fmt.Errorf("no such suite: %s", name)
	}

	return suite, nil
}

func (e *Extension) GetSpecs() et.ExtensionTestSpecs {
	return e.specs
}

func (e *Extension) AddSpecs(specs et.ExtensionTestSpecs) {
	specs.Walk(func(spec *et.ExtensionTestSpec) {
		spec.Source = e.Component.Identifier()
	})

	e.specs = append(e.specs, specs...)
}

// IgnoreObsoleteTests allows removal of a test.
func (e *Extension) IgnoreObsoleteTests(testNames ...string) {
	e.obsoleteTests = append(e.obsoleteTests, testNames...)
}

// AddGlobalSuite adds a suite whose qualifiers will apply to all tests,
// not just this one.  Allowing a developer to create a composed suite of
// tests from many sources.
func (e *Extension) AddGlobalSuite(suite Suite) *Extension {
	if e.Suites == nil {
		e.Suites = []Suite{suite}
	} else {
		e.Suites = append(e.Suites, suite)
	}

	return e
}

// AddSuite adds a suite whose qualifiers will only apply to tests present
// in its own extension.
func (e *Extension) AddSuite(suite Suite) *Extension {
	expr := fmt.Sprintf("source == %q", e.Component.Identifier())
	for i := range suite.Qualifiers {
		suite.Qualifiers[i] = fmt.Sprintf("(%s) && (%s)", expr, suite.Qualifiers[i])
	}
	e.AddGlobalSuite(suite)
	return e
}

func (e *Component) Identifier() string {
	return fmt.Sprintf("%s:%s:%s", e.Product, e.Kind, e.Name)
}
