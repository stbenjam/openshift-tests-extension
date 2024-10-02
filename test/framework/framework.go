package framework

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These tests have to be run from the root of the repository, as ginkgo seems to mess up
// the CWD and runtime Caller doesn't get the right path.  FIXME?
const binary = "./example-tests"

var _ = Describe("[sig-testing] openshift-tests-extension", Ordered, Label("framework"), func() {
	It("should be run from the main directory", func() {
		_, err := os.Stat("Makefile")
		Expect(err).ShouldNot(HaveOccurred(), "Expected to be run from the root directory")
	})

	It("should run `make` command successfully", func() {
		cmd := exec.Command("make", "example-tests")

		// Run the command
		err := cmd.Run()
		Expect(err).ShouldNot(HaveOccurred(), "Expected `make` to run successfully")
	})

	It("should have the example-tests binary", func() {
		_, err := os.Stat(binary)
		Expect(err).ShouldNot(HaveOccurred(), "Expected `example-tests` binary to exist")
	})

})
