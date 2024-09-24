package example

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	g "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
)

// FIXME(stbenjam): ginkgo doesn't allow "/" in label names, so it's hard to use this for our existing suite names,
// maybe convert "." to "/" always?
var _ = Describe("Simple Tests", g.Suite("openshift.conformance.parallel"), func() {
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

	It("should only run on AWS with annotation [Include:Platform:AWS]", g.Suite("example.extension"), func() {
		Expect(true).To(BeTrue())
	})
})
