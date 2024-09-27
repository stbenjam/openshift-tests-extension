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

	// Import your tests here
	_ "github.com/openshift-eng/openshift-tests-extension/test/example"
)

// TODO:remove this
const suite = "OpenShift Tests Extension"

func main() {
	// Extension registry
	registry := extensions.NewRegistry()

	// Default extension
	ext := extensions.NewExtension("openshift", "payload", "example-tests")
	ext.AddSuite(extensions.Suite{Name: "openshift/conformance/parallel"})
	ext.AddSuite(extensions.Suite{Name: "example/tests"})
	registry.Register(ext)

	root := &cobra.Command{
		Long: "OpenShift Tests Extension Example",
	}

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
