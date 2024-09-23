package cmdrun

import (
	"os"

	"github.com/spf13/cobra"

	g "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
)

func NewCommand(suite string) *cobra.Command {
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
