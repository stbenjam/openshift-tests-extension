package extension

import (
	"github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
)

const CurrentExtensionVersion = "v1"

// Extension represents an extension to openshift-tests.
type Extension struct {
	APIVersion string    `json:"apiVersion"`
	Source     Source    `json:"source"`
	Component  Component `json:"component"`

	// Suites that the extension wants to advertise/participate in.
	Suites []Suite `json:"suites"`

	// Private data
	specs []*extensiontests.ExtensionTestSpec
}

func (e *Extension) GetSpecs() extensiontests.ExtensionTestSpecs {
	return e.specs
}

func (e *Extension) AddSpecs(specs extensiontests.ExtensionTestSpecs) {
	specs.Walk(func(spec *extensiontests.ExtensionTestSpec) {
		spec.Source = e.Component.Identifier()
	})

	e.specs = append(e.specs, specs...)
}

// Source contains the details of the commit and source URL.
type Source struct {
	// Commit from which this binary was compiled.
	Commit string `json:"commit"`
	// BuildDate ISO8601 string of when the binary was built
	BuildDate string `json:"build_date"`
	// GitTreeState lets you know the status of the git tree (clean/dirty)
	GitTreeState string `json:"git_tree_state"`
	// SourceURL contains the url of the git repository (if known) that this extension was built from.
	SourceURL string `json:"source_url,omitempty"`
}

// Component represents the component the binary acts on.
type Component struct {
	// The product this component is part of.
	Product string `json:"product"`
	// The type of the component.
	Kind string `json:"type"`
	// The name of the component.
	Name string `json:"name"`
}

// Suite represents additional suites the extension wants to advertise.
type Suite struct {
	// The name of the suite.
	Name string `json:"name"`
	// Parent suites this suite is part of.
	Parents []string `json:"parents,omitempty"`
	// Qualifiers are CEL expressions that are OR'd together for test selection that are members of the suite.
	Qualifiers []string `json:"qualifiers,omitempty"`
}
