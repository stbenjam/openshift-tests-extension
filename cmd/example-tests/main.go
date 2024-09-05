package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Long: "OpenShift Tests External Binary Example",
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
	testOpt := NewTestOptions(os.Stdout, os.Stderr)

	cmd := &cobra.Command{
		Use:          "run-test NAME",
		Short:        "Run a single test by name",
		Long:         "Execute a single test.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return testOpt.Run(args)
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
			tests := testsForSuite()
			sort.Slice(tests, func(i, j int) bool { return tests[i].Name < tests[j].Name })
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
