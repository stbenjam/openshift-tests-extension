package cmdlist

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	"github.com/openshift-eng/openshift-tests-extension/pkg/flags"
)

func NewListCommand(registry *extension.Registry) *cobra.Command {
	opts := struct {
		componentFlags *flags.ComponentFlags
		suiteFlags     *flags.SuiteFlags
		outputFlags    *flags.OutputFlags
	}{
		suiteFlags:     flags.NewSuiteFlags(),
		componentFlags: flags.NewComponentFlags(),
		outputFlags:    flags.NewOutputFlags(),
	}

	listTestsCmd := &cobra.Command{
		Use:          "tests",
		Short:        "List available tests",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ext := registry.Get(opts.componentFlags.Component)
			if ext == nil {
				return fmt.Errorf("component not found: %s", opts.componentFlags.Component)
			}

			// Find suite, if specified
			var foundSuite *extension.Suite
			var err error
			if opts.suiteFlags.Suite != "" {
				foundSuite, err = ext.GetSuite(opts.suiteFlags.Suite)
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

			data, err := opts.outputFlags.Marshal(specs)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s\n", data)
			return nil
		},
	}
	opts.suiteFlags.BindFlags(listTestsCmd.Flags())
	opts.componentFlags.BindFlags(listTestsCmd.Flags())
	opts.outputFlags.BindFlags(listTestsCmd.Flags())

	listComponentsCmd := &cobra.Command{
		Use:          "components",
		Short:        "List available components",
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
	opts.suiteFlags.BindFlags(listCmd.Flags())
	opts.componentFlags.BindFlags(listCmd.Flags())
	opts.outputFlags.BindFlags(listCmd.Flags())
	listCmd.AddCommand(listTestsCmd, listComponentsCmd)

	return listCmd
}
