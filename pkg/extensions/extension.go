package extensions

var (
	// CommitFromGit is a constant representing the source version that
	// generated this build. It should be set during build via -ldflags.
	CommitFromGit string
	// BuildDate in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
	BuildDate string
	// GitTreeState has the state of git tree, either "clean" or "dirty"
	GitTreeState string
)

var DefaultExtension = Extension{
	APIVersion: "v1",
	Source: Source{
		Commit:       CommitFromGit,
		BuildDate:    BuildDate,
		GitTreeState: GitTreeState,
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
