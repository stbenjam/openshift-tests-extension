package cmdrun

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	"github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
	"github.com/openshift-eng/openshift-tests-extension/pkg/flags"
)

func NewCommand(registry *extension.Registry) *cobra.Command {
	var runOpts struct {
		componentFlags *flags.ComponentFlags
		nameFlags      *flags.NamesFlags
	}
	runOpts.componentFlags = flags.NewComponentFlags()
	runOpts.nameFlags = flags.NewNamesFlags()

	cmd := &cobra.Command{
		Use:          "run-test NAME",
		Short:        "RunTest a single test by name",
		Long:         "Execute a single test.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ext := registry.Get(runOpts.componentFlags.Component)
			if ext == nil {
				return fmt.Errorf("component not found: %s", runOpts.componentFlags.Component)
			}
			if len(args) > 1 {
				return fmt.Errorf("use --names to specify more than one test")
			}
			runOpts.nameFlags.Names = append(runOpts.nameFlags.Names, args...)
			if len(runOpts.nameFlags.Names) == 0 {
				return fmt.Errorf("must specify at least one test")
			}

			var results extensiontests.ExtensionTestResults
			var specs extensiontests.ExtensionTestSpecs
			var notFound []string
			for _, name := range runOpts.nameFlags.Names {
				spec, err := ext.FindSpecByName(name)
				if err != nil {
					notFound = append(notFound, name)
					continue
				}
				specs = append(specs, spec)
			}
			if len(notFound) > 0 {
				return fmt.Errorf("some tests couldn't be found: \n\t* %s", strings.Join(notFound, "\n\t* "))
			}

			// Run each test
			for _, spec := range specs {
				startTime := time.Now()
				res := spec.Run()
				duration := time.Since(startTime)
				endTime := startTime.Add(duration)
				if res == nil {
					// this shouldn't happen
					panic(fmt.Sprintf("test produced no result: %s", spec.Name))
				}

				res.Lifecycle = spec.Lifecycle

				// If the runner doesn't populate this info, we should set it
				if res.StartTime == nil {
					res.StartTime = &startTime
				}
				if res.EndTime == nil {
					res.EndTime = &endTime
				}
				if res.Duration == 0 {
					res.Duration = duration.Milliseconds()
				}

				results = append(results, res)
			}

			res, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				return errors.Wrap(err, "couldn't marshal results")
			}

			fmt.Println(string(res))
			return nil
		},
	}
	runOpts.componentFlags.BindFlags(cmd.Flags())
	runOpts.nameFlags.BindFlags(cmd.Flags())

	return cmd
}
