package extensions

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	"github.com/pkg/errors"

	g "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
	"github.com/openshift-eng/openshift-tests-extension/pkg/testspec"
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

func (e *Extension) BuildExtensionTestSpecsFromOpenShiftGinkgoSuite() ([]*testspec.TestSpec, error) {
	var tests []*testspec.TestSpec

	if !ginkgo.GetSuite().InPhaseBuildTree() {
		if err := ginkgo.GetSuite().BuildTree(); err != nil {
			return nil, errors.Wrapf(err, "couldn't build ginkgo tree")
		}
	}

	ginkgo.GetSuite().WalkTests(func(name string, spec types.TestSpec) {
		testCase := &testspec.TestSpec{
			Name:      spec.Text(),
			Labels:    spec.Labels(),
			Lifecycle: g.GetLifecycle(spec.Labels()),
		}
		tests = append(tests, testCase)
	})

	e.specs = tests
	return tests, nil
}
