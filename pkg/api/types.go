package api

// Extension represents an extension to openshift-tests.
type Extension struct {
	APIVersion string    `json:"apiVersion"`
	Source     Source    `json:"source"`
	Component  Component `json:"component"`

	// Suites that the extension wants to advertise/participate in.
	Suites []Suite `json:"suites"`
}

// Source contains the details of the commit and source URL.
type Source struct {
	// The commit from which this binary was compiled.
	Commit string `json:"commit"`
	// The git repository (if known) that this extension was built from.
	SourceURL string `json:"source_url"`
}

// Component represents the component the binary acts on.
type Component struct {
	// The product this component is part of.
	Product string `json:"product"`
	// The type of the component.
	Type string `json:"type"`
	// The name of the component.
	Name string `json:"name"`
}

// Suite represents additional suites the extension wants to advertise.
type Suite struct {
	// The name of the suite.
	Name string `json:"name"`
	// Parent suites this suite is part of.
	Parents []string `json:"parents"`
}
