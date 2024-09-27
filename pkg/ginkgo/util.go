package ginkgo

import (
	"fmt"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	"github.com/pkg/errors"

	testspec2 "github.com/openshift-eng/openshift-tests-extension/pkg/extensions/testspec"
)

func BuildExtensionTestSpecsFromOpenShiftGinkgoSuite() ([]*testspec2.ExtensionTestSpec, error) {
	var tests []*testspec2.ExtensionTestSpec

	if !ginkgo.GetSuite().InPhaseBuildTree() {
		if err := ginkgo.GetSuite().BuildTree(); err != nil {
			return nil, errors.Wrapf(err, "couldn't build ginkgo tree")
		}
	}

	ginkgo.GetSuite().WalkTests(func(name string, spec types.TestSpec) {
		testCase := &testspec2.ExtensionTestSpec{
			Name:      spec.Text(),
			Labels:    spec.Labels(),
			Lifecycle: GetLifecycle(spec.Labels()),
		}
		tests = append(tests, testCase)
	})

	return tests, nil
}

func Suite(name string) ginkgo.Labels {
	return ginkgo.Label(fmt.Sprintf("Suite:%s", name))
}

func Informing() ginkgo.Labels {
	return ginkgo.Label(fmt.Sprintf("Lifecycle:%s", testspec2.LifecycleInforming))
}

func Blocking() ginkgo.Labels {
	return ginkgo.Label(fmt.Sprintf("Lifecycle:%s", testspec2.LifecycleBlocking))
}

func GetLifecycle(labels ginkgo.Labels) testspec2.Lifecycle {
	for _, label := range labels {
		res := strings.Split(label, ":")
		if len(res) != 2 || !strings.EqualFold(res[0], "lifecycle") {
			continue
		}
		return testspec2.MustLifecycle(res[1]) // this panics if unsupported lifecycle is used
	}

	return testspec2.LifecycleBlocking
}
