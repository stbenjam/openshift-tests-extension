package cmdlist

import (
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
		outputFlags    *flags.OutputFlags
	}
	listOpts.suiteFlags = flags.NewSuiteFlags() //FIXME: filter on this
	listOpts.componentFlags = flags.NewComponentFlags()
	listOpts.outputFlags = flags.NewOutputFlags()

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

			// Find suite, if specified
			var foundSuite *extension.Suite
			var err error
			if listOpts.suiteFlags.Suite != "" {
				foundSuite, err = ext.GetSuite(listOpts.suiteFlags.Suite)
				if err != nil {
					return err
				}
			}

			// Filter for suite
			specs := ext.GetSpecs()
			if foundSuite != nil {
				specs, err = specs.Filter(foundSuite.Qualifiers)
				if err != nil {
					return err
				}
			}

			data, err := listOpts.outputFlags.Marshal(specs)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s\n", data)
			return nil
		},
	}
	listOpts.suiteFlags.BindFlags(listTestsCmd.Flags())
	listOpts.componentFlags.BindFlags(listTestsCmd.Flags())
	listOpts.outputFlags.BindFlags(listTestsCmd.Flags())

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
	listOpts.outputFlags.BindFlags(listCmd.Flags())
	listCmd.AddCommand(listTestsCmd, listComponentsCmd)

	return listCmd
}
