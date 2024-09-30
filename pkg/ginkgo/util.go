package ginkgo

import (
	"fmt"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/sets"

	ext "github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
)

func BuildExtensionTestSpecsFromOpenShiftGinkgoSuite() (ext.ExtensionTestSpecs, error) {
	var tests []*ext.ExtensionTestSpec

	if !ginkgo.GetSuite().InPhaseBuildTree() {
		if err := ginkgo.GetSuite().BuildTree(); err != nil {
			return nil, errors.Wrapf(err, "couldn't build ginkgo tree")
		}
	}

	ginkgo.GetSuite().WalkTests(func(name string, spec types.TestSpec) {
		testCase := &ext.ExtensionTestSpec{
			Name:      spec.Text(),
			Labels:    sets.New[string](spec.Labels()...),
			Lifecycle: GetLifecycle(spec.Labels()),
			Run: func() *ext.ExtensionTestResult {
				return RunSpec(spec)
			},
		}
		tests = append(tests, testCase)
	})

	return tests, nil
}

func Informing() ginkgo.Labels {
	return ginkgo.Label(fmt.Sprintf("Lifecycle:%s", ext.LifecycleInforming))
}

func Blocking() ginkgo.Labels {
	return ginkgo.Label(fmt.Sprintf("Lifecycle:%s", ext.LifecycleBlocking))
}

func GetLifecycle(labels ginkgo.Labels) ext.Lifecycle {
	for _, label := range labels {
		res := strings.Split(label, ":")
		if len(res) != 2 || !strings.EqualFold(res[0], "lifecycle") {
			continue
		}
		return MustLifecycle(res[1]) // this panics if unsupported lifecycle is used
	}

	return ext.LifecycleBlocking
}

func MustLifecycle(l string) ext.Lifecycle {
	switch ext.Lifecycle(l) {
	case ext.LifecycleInforming, ext.LifecycleBlocking:
		return ext.Lifecycle(l)
	default:
		panic(fmt.Sprintf("unknown test lifecycle: %s", l))
	}
}
