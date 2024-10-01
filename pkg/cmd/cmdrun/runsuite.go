package cmdrun

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
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
			ext := registry.Get(runOpts.componentFlags.Component)
			if ext == nil {
				return fmt.Errorf("component not found: %s", runOpts.componentFlags.Component)
			}
			if len(args) != 1 {
				return fmt.Errorf("must specify one suite name")
			}

			w, err := extensiontests.NewResultWriter(os.Stdout, extensiontests.ResultFormat(runOpts.outputFlags.Output))
			if err != nil {
				return err
			}
			defer w.Flush()

			suite, err := ext.GetSuite(args[0])
			if err != nil {
				return errors.Wrapf(err, "couldn't find suite: %s", args[0])
			}

			specs, err := ext.GetSpecs().Filter(suite.Qualifiers)
			if err != nil {
				return errors.Wrap(err, "couldn't filter specs")
			}

			return specs.Run(w)
		},
	}
	runOpts.componentFlags.BindFlags(cmd.Flags())
	runOpts.outputFlags.BindFlags(cmd.Flags())

	return cmd
}
