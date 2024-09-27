package ginkgo

import (
	"fmt"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"github.com/openshift-eng/openshift-tests-extension/pkg/testspec"
)

func Suite(name string) ginkgo.Labels {
	return ginkgo.Label(fmt.Sprintf("Suite:%s", name))
}

func Informing() ginkgo.Labels {
	return ginkgo.Label(fmt.Sprintf("Lifecycle:%s", testspec.LifecycleInforming))
}

func Blocking() ginkgo.Labels {
	return ginkgo.Label(fmt.Sprintf("Lifecycle:%s", testspec.LifecycleBlocking))
}

func GetLifecycle(labels ginkgo.Labels) testspec.Lifecycle {
	for _, label := range labels {
		res := strings.Split(label, ":")
		if len(res) != 2 || !strings.EqualFold(res[0], "lifecycle") {
			continue
		}
		return testspec.MustLifecycle(res[1]) // this panics if unsupported lifecycle is used
	}

	return testspec.LifecycleBlocking
}
