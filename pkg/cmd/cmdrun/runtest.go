package cmdrun

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	"github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
	"github.com/openshift-eng/openshift-tests-extension/pkg/flags"
)

func NewRunTestCommand(registry *extension.Registry) *cobra.Command {
	var runOpts struct {
		componentFlags *flags.ComponentFlags
		nameFlags      *flags.NamesFlags
		outputFlags    *flags.OutputFlags
	}
	runOpts.componentFlags = flags.NewComponentFlags()
	runOpts.nameFlags = flags.NewNamesFlags()
	runOpts.outputFlags = flags.NewOutputFlags()

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

			specs, err := ext.FindSpecsByName(runOpts.nameFlags.Names...)
			if err != nil {
				return err
			}

			w, err := extensiontests.NewResultWriter(os.Stdout, extensiontests.ResultFormat(runOpts.outputFlags.Output))
			if err != nil {
				return err
			}
			defer w.Flush()

			return specs.Run(w)
		},
	}
	runOpts.componentFlags.BindFlags(cmd.Flags())
	runOpts.nameFlags.BindFlags(cmd.Flags())
	runOpts.outputFlags.BindFlags(cmd.Flags())

	return cmd
}
