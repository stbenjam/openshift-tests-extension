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
	g "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"

	// Import your tests here
	_ "github.com/openshift-eng/openshift-tests-extension/test/example"
)

// TODO:remove this
const suite = "OpenShift Tests Extension"

func main() {

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

func newRunTestCommand() *cobra.Command {
	testOpt := g.NewTestOptions(os.Stdout, os.Stderr)

	cmd := &cobra.Command{
		Use:          "run-test NAME",
		Short:        "RunTest a single test by name",
		Long:         "Execute a single test.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return testOpt.RunTest(args, suite)
		},
	}
	return cmd
}
