package extensions

import "github.com/openshift-eng/openshift-tests-extension/pkg/version"

var DefaultExtension = Extension{
	APIVersion: "v1",
	Source: Source{
		Commit:       version.CommitFromGit,
		BuildDate:    version.BuildDate,
		GitTreeState: version.GitTreeState,
		SourceURL:    "https://github.com/openshift-eng/openshift-tests-extension",
	},
	Component: Component{
		Product: "openshift",
		Type:    "payload",
		Name:    "example-tests",
	},
	Suites: []Suite{
		// Includes tests that are part of openshift/conformance/parallel
		{
			Name: "openshift/conformance/parallel",
		},
		// Adds a new suite called, "example/extension"
		{
			Name: "example/extension",
		},
	},
}
