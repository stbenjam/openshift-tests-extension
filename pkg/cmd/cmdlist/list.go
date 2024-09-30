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

	listTestsCmd := &cobra.Command{
		Use:          "tests",
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
	listOpts.suiteFlags.BindFlags(listTestsCmd.Flags())
	listOpts.componentFlags.BindFlags(listTestsCmd.Flags())

	listComponentsCmd := &cobra.Command{
		Use:          "components",
		Short:        "List available components",
		Long:         "List the available components in this binary.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			registry.Walk(func(e *extension.Extension) {
				fmt.Printf("%s:%s:%s\n", e.Component.Product, e.Component.Kind, e.Component.Name)
			})
			return nil
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list [subcommand]",
		Short: "List items",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTestsCmd.RunE(cmd, args)
		},
	}
	listOpts.suiteFlags.BindFlags(listCmd.Flags())
	listOpts.componentFlags.BindFlags(listCmd.Flags())
	listCmd.AddCommand(listTestsCmd, listComponentsCmd)

	return listCmd
}
