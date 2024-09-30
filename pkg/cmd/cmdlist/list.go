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
		componentFlags *flags.ComponentFlags
		suiteFlags     *flags.SuiteFlags
	}
	listOpts.suiteFlags = flags.NewSuiteFlags() //FIXME: filter on this
	listOpts.componentFlags = flags.NewComponentFlags()

	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List available tests",
		Long:         "List the available tests in this binary.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ext := registry.Get(listOpts.componentFlags.Component)
			if ext == nil {
				return fmt.Errorf("component not found: %s", listOpts.componentFlags.Component)
			}

			data, err := json.MarshalIndent(ext.GetSpecs(), "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s\n", data)
			return nil
		},
	}
	listOpts.suiteFlags.BindFlags(cmd.Flags())

	return cmd
}
