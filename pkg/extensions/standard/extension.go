package standard

import "github.com/openshift-eng/openshift-tests-extension/pkg/api"

var Extension = api.Extension{
	APIVersion: "1.0",
	Source: api.Source{
		SourceURL: "https://github.com/openshift-eng/openshift-tests-extension",
	},
	Component: api.Component{
		Product: "openshift",
		Type:    "payload",
		Name:    "example-tests",
	},
	Suites: []api.Suite{
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
