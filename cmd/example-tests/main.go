package main

import (
	"fmt"
	"os"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd/cmdinfo"
	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd/cmdlist"
	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd/cmdrun"
	"github.com/openshift-eng/openshift-tests-extension/pkg/extensions"

	// If using ginkgo, import your tests here
	_ "github.com/openshift-eng/openshift-tests-extension/test/example"
)

// TODO:remove this
const suite = "OpenShift Tests Extension"

func main() {
	// Extension registry
	registry := extensions.NewRegistry()

	// Default extension -- more than one are possible and selectable with the "--component" flag
	ext := extensions.NewExtension("openshift", "payload", "example-tests")
	ext.AddSuite(extensions.Suite{Name: "openshift/conformance/parallel"})
	ext.AddSuite(extensions.Suite{Name: "example/tests", Parents: []string{"openshift/conformance/parallel"}})

	// If using Gingko, build test specs automatically
	_, err := ext.BuildExtensionTestSpecsFromOpenShiftGinkgoSuite()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	// If not using gingko build the test specs manually
	// TODO:example

	registry.Register(ext)

	root := &cobra.Command{
		Long: "OpenShift Tests Extension Example",
	}

	// TODO: Wire up extension stuff
	root.AddCommand(
		cmdrun.NewCommand(suite),
		cmdlist.NewCommand(),
		cmdinfo.NewCommand(),
	)

	gomega.RegisterFailHandler(ginkgo.Fail)

	if err := func() error {
		return root.Execute()
	}(); err != nil {
		if ex, ok := err.(ExitError); ok {
			fmt.Fprintf(os.Stderr, "Ginkgo exit error %d: %v\n", ex.Code, err)
			os.Exit(ex.Code)
		}
		os.Exit(1)
	}
}
