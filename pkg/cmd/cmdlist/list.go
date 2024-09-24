package cmdlist

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/openshift-eng/openshift-tests-extension/pkg/flags"
	g "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
	"github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo/ginkgofilter"
)

func NewCommand() *cobra.Command {
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
			tests := g.ListTests()

			if !listOpts.all {
				// Filter by env flags
				envSet := sets.New[string](listOpts.envFlags.Environments...)
				tests = ginkgofilter.FilterTestCasesByEnvironment(tests, envSet)
			}

			if suite := listOpts.suiteFlags.Suite; suite != "" {
				tests = ginkgofilter.FilterTestCasesBySuite(tests, suite)
			}

			data, err := json.MarshalIndent(tests, "", "  ")
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
