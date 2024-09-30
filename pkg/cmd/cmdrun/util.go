package cmdrun

import (
	"fmt"
	"time"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
)

func runSpec(spec *extensiontests.ExtensionTestSpec) *extensiontests.ExtensionTestResult {
	startTime := time.Now()
	res := spec.Run()
	duration := time.Since(startTime)
	endTime := startTime.Add(duration)
	if res == nil {
		// this shouldn't happen
		panic(fmt.Sprintf("test produced no result: %s", spec.Name))
	}

	res.Lifecycle = spec.Lifecycle

	// If the runner doesn't populate this info, we should set it
	if res.StartTime == nil {
		res.StartTime = &startTime
	}
	if res.EndTime == nil {
		res.EndTime = &endTime
	}
	if res.Duration == 0 {
		res.Duration = duration.Milliseconds()
	}

	return res
}
