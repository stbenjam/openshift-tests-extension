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
	envFlags := flags.NewEnvironmentFlags()

	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List available tests",
		Long:         "List the available tests in this binary.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			tests := g.ListTests()

			// Filter by env flags
			envSet := sets.New[string](envFlags.Environments...)
			tests = ginkgofilter.FilterTestCases(tests, envSet)

			data, err := json.Marshal(tests)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s\n", data)
			return nil
		},
	}
	envFlags.BindFlags(cmd.Flags())

	return cmd
}
