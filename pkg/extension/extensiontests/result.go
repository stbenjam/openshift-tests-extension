package extensiontests

import "fmt"

func (results ExtensionTestResults) Walk(walkFn func(*ExtensionTestResult)) {
	for i := range results {
		walkFn(results[i])
	}
}

func (results ExtensionTestResults) CheckOverallResult() error {
	failed := 0

	results.Walk(func(result *ExtensionTestResult) {
		if result.Result == ResultFailed && result.Lifecycle == LifecycleBlocking {
			failed++
		}
	})

	if failed > 0 {
		return fmt.Errorf("%d tests failed", failed)
	}
	return nil
}
