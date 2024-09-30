package cmdrun

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	"github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
	"github.com/openshift-eng/openshift-tests-extension/pkg/flags"
)

func NewRunSuiteCommand(registry *extension.Registry) *cobra.Command {
	var runOpts struct {
		componentFlags *flags.ComponentFlags
	}
	runOpts.componentFlags = flags.NewComponentFlags()

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
				cmd.Help()
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

			// Runs pecs
			for _, spec := range specs {
				res := runSpec(spec)
				results = append(results, res)
			}

			j, err := json.Marshal(results)
			if err != nil {
				return fmt.Errorf("failed to marshal results: %v", err)
			}
			fmt.Println(string(j))

			return results.CheckOverallResult()
		},
	}
	runOpts.componentFlags.BindFlags(cmd.Flags())

	return cmd
}
