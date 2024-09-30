package cmdlist

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	"github.com/openshift-eng/openshift-tests-extension/pkg/flags"
)

func NewCommand(registry *extension.Registry) *cobra.Command {
	var listOpts struct {
		all        bool
		envFlags   *flags.EnvironmentFlags
		suiteFlags *flags.SuiteFlags
	}
	listOpts.envFlags = flags.NewEnvironmentFlags()
	listOpts.suiteFlags = flags.NewSuiteFlags()

	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List available tests",
		Long:         "List the available tests in this binary.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			/*tests := g.ListTests()

			if !listOpts.all {
				// Filter by env flags
				envSet := sets.New[string](listOpts.envFlags.Environments...)
				tests = ginkgofilter.FilterTestCasesByEnvironment(tests, envSet)
			}

			if suite := listOpts.suiteFlags.Suite; suite != "" {
				tests = ginkgofilter.FilterTestCasesBySuite(tests, suite)
			}*/

			data, err := json.MarshalIndent(registry.Get("default").GetSpecs(), "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s\n", data)
			return nil
		},
	}
	cmd.Flags().BoolVar(&listOpts.all, "all", false, "Show all tests, with no environment filtering")
	listOpts.envFlags.BindFlags(cmd.Flags())
	listOpts.suiteFlags.BindFlags(cmd.Flags())

	return cmd
}
