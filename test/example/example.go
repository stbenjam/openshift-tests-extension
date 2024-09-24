package example

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Simple Tests", func() {
	It("should print 'Hello, OpenShift!'", func() {
		fmt.Println("Hello, OpenShift!")
		Expect(true).To(BeTrue()) // This ensures the test passes
	})

	It("should fail and print a sad face", func() {
		fmt.Println(":(")
		Expect(true).To(BeFalse()) // This makes the test fail
	})

	It("should filter test results by label", Labels([]string{"Skipped:Platform:AWS"}), func() {
		Expect(true).To(BeTrue())
	})

	It("should only run on AWS with annotation [Include:Platform:AWS]", func() {
		Expect(true).To(BeTrue())
	})
})
