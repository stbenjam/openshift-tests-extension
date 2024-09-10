package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/spf13/cobra"

	g "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"

	// Import your tests here
	_ "github.com/openshift-eng/openshift-tests-extension/test/example"
)

const suite = "OpenShift Tests Extension"

func main() {
	root := &cobra.Command{
		Long: "OpenShift Tests Extension Example",
	}

	root.AddCommand(
		newRunTestCommand(),
		newListTestsCommand(),
	)

	gomega.RegisterFailHandler(ginkgo.Fail)

	if err := func() error {
		return root.Execute()
	}(); err != nil {
		if ex, ok := err.(ExitError); ok {
			fmt.Fprintf(os.Stderr, "Ginkgo exit error %d: %v\n", ex.Code, err)
			os.Exit(ex.Code)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
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

func newListTestsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List available tests",
		Long:         "List the available tests in this binary.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			tests := g.ListTests()
			data, err := json.Marshal(tests)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s\n", data)
			return nil
		},
	}

	return cmd
}
