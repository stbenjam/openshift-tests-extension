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
		suiteFlags     *flags.SuiteFlags
	}
	runOpts.componentFlags = flags.NewComponentFlags()
	runOpts.suiteFlags = flags.NewSuiteFlags()

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
			if runOpts.suiteFlags.Suite == "" {
				cmd.Help()
				return fmt.Errorf("please specify a suite")
			}
			var foundSuite *extension.Suite
			for _, suite := range ext.Suites {
				if suite.Name == runOpts.suiteFlags.Suite {
					foundSuite = &suite
				}
			}
			if foundSuite == nil {
				return fmt.Errorf("couldn't find suite: %s", runOpts.suiteFlags.Suite)
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

			j, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal results: %v", err)
			}
			fmt.Println(string(j))

			return nil
		},
	}
	runOpts.componentFlags.BindFlags(cmd.Flags())
	runOpts.suiteFlags.BindFlags(cmd.Flags())

	return cmd
}
