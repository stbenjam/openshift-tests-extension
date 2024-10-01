package cmdrun

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	"github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
	"github.com/openshift-eng/openshift-tests-extension/pkg/flags"
)

func NewRunSuiteCommand(registry *extension.Registry) *cobra.Command {
	var runOpts struct {
		componentFlags *flags.ComponentFlags
		outputFlags    *flags.OutputFlags
	}
	runOpts.componentFlags = flags.NewComponentFlags()
	runOpts.outputFlags = flags.NewOutputFlags()

	cmd := &cobra.Command{
		Use:          "run-suite NAME",
		Short:        "Run a group of tests by suite",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var results extensiontests.ExtensionTestResults
			ext := registry.Get(runOpts.componentFlags.Component)
			if ext == nil {
				return fmt.Errorf("component not found: %s", runOpts.componentFlags.Component)
			}
			if len(args) != 1 {
				return fmt.Errorf("must specify one suite name")
			}
			var foundSuite *extension.Suite
			for _, suite := range ext.Suites {
				if suite.Name == args[0] {
					foundSuite = &suite
				}
			}
			if foundSuite == nil {
				return fmt.Errorf("couldn't find suite: %s", args[0])
			}

			// Filter for suite
			specs := ext.GetSpecs()
			if len(foundSuite.Qualifiers) > 0 {
				specs = specs.MustFilter(foundSuite.Qualifiers)
			}

			// Run specs
			w, err := extensiontests.NewResultWriter(os.Stdout, extensiontests.ResultFormat(runOpts.outputFlags.Output))
			if err != nil {
				return err
			}
			for _, spec := range specs {
				res := runSpec(spec)
				w.Write(res)
				results = append(results, res)
			}
			w.Flush()

			if failed := results.CheckOverallResult(); failed != nil {
				os.Exit(1) // exit 1 without letting cobra print the error and pollute our output
			}

			return nil
		},
	}
	runOpts.componentFlags.BindFlags(cmd.Flags())
	runOpts.outputFlags.BindFlags(cmd.Flags())

	return cmd
}
