package framework

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	e "github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
)

// suiteTimeout is the maximum time allowed for running the test suite.
const suiteTimeout = 5 * time.Minute

// These tests exercise the full OTE pipeline with a binary that deliberately
// contaminates stdout with realistic non-JSON output (klog lines, Ginkgo
// reporter text, etc.) during run-test subprocess execution. This verifies
// that extractJSON correctly handles stdout contamination end-to-end.
//
// The example-tests binary is already built by the framework.go Ordered
// container, so no separate build step is needed here.
var _ = Describe("[sig-testing] example-tests stdout contamination", Ordered, Label("framework"), func() {
	Context("run-suite", func() {
		var result e.ExtensionTestResults
		var output []byte
		var cmdErr error

		BeforeEach(func() {
			ctx, cancel := context.WithTimeout(context.Background(), suiteTimeout)
			defer cancel()

			cmd := exec.CommandContext(ctx, binary, "run-suite", "example/fast")

			// Capture both stdout and stderr
			output, cmdErr = cmd.Output()

			// Expect command to exit with a non-zero status (exit code 1 for failed tests)
			var exitErr *exec.ExitError
			ok := errors.As(cmdErr, &exitErr)
			Expect(ok).To(BeTrue(), "Expected command to exit with a non-zero status")
			Expect(exitErr.ExitCode()).To(Equal(1), "Expected exit code 1")

			// Unmarshal the JSON output into the predefined ExtensionTestResults type.
			// This verifies that extractJSON correctly handled the non-JSON stdout
			// contamination from klog/Ginkgo output in the run-test subprocesses.
			err := json.Unmarshal(output, &result)
			Expect(err).ShouldNot(HaveOccurred(), "Expected JSON output to unmarshal into ExtensionTestResults despite subprocess stdout contamination")
		})

		It("should contain a test that passed despite stdout contamination", func() {
			var foundPassed bool
			for _, test := range result {
				if test.Result == "passed" {
					foundPassed = true
					break
				}
			}
			Expect(foundPassed).To(BeTrue(), "Expected at least one test to have passed despite stdout contamination")
		})

		It("should contain a test that failed despite stdout contamination", func() {
			var foundFailed bool
			for _, test := range result {
				if test.Result == "failed" {
					foundFailed = true
					break
				}
			}
			Expect(foundFailed).To(BeTrue(), "Expected at least one test to have failed despite stdout contamination")
		})

		It("should have the correct error for panicking test despite stdout contamination", func() {
			var found bool
			for _, test := range result {
				if test.Name == "[sig-testing] openshift-tests-extension should support panicking tests" && test.Result == "failed" {
					found = true
					Expect(test.Error).To(ContainSubstring("Test Panicked: oh no"), "Expected error to contain 'Test Panicked: oh no'")
					break
				}
			}
			Expect(found).To(BeTrue(), "Expected the panicking test result to be present")
		})

		It("should handle pending tests correctly despite stdout contamination", func() {
			var foundPending bool
			for _, test := range result {
				if test.Name == "[sig-testing] openshift-tests-extension should support pending tests" && test.Result == "skipped" {
					foundPending = true
					break
				}
			}
			Expect(foundPending).To(BeTrue(), "Expected pending test (XIt) to be reported as skipped despite stdout contamination")
		})
	})
})
