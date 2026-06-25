package framework

import (
	"encoding/json"
	"errors"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	e "github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
)

const noisyBinary = "./example-noisy-tests"

// These tests exercise the full OTE pipeline with a binary that deliberately
// contaminates stdout with realistic non-JSON output (klog lines, Ginkgo
// reporter text, etc.) during run-test subprocess execution. This verifies
// that extractJSON correctly handles stdout contamination end-to-end.
var _ = Describe("[sig-testing] example-noisy-tests", Ordered, Label("framework"), func() {
	It("should build the noisy binary", func() {
		cmd := exec.Command("make", "example-noisy-tests")
		err := cmd.Run()
		Expect(err).ShouldNot(HaveOccurred(), "Expected `make example-noisy-tests` to run successfully")
	})
})

var _ = Describe("[sig-testing] example-noisy-tests run-suite", Label("framework"), func() {
	var result e.ExtensionTestResults
	var output []byte
	var cmdErr error

	BeforeEach(func() {
		cmd := exec.Command(noisyBinary, "run-suite", "example/fast")

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
		for _, test := range result {
			if test.Name == "[sig-testing] openshift-tests-extension should support panicking tests" && test.Result == "failed" {
				Expect(test.Error).To(ContainSubstring("Test Panicked: oh no"), "Expected error to contain 'Test Panicked: oh no'")
				break
			}
		}
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
