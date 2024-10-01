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
	opts := struct {
		componentFlags *flags.ComponentFlags
		nameFlags      *flags.NamesFlags
		outputFlags    *flags.OutputFlags
	}{
		componentFlags: flags.NewComponentFlags(),
		nameFlags:      flags.NewNamesFlags(),
		outputFlags:    flags.NewOutputFlags(),
	}

	cmd := &cobra.Command{
		Use:          "run-test [-n NAME...] [NAME]",
		Short:        "Runs tests by name",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ext := registry.Get(opts.componentFlags.Component)
			if ext == nil {
				return fmt.Errorf("component not found: %s", opts.componentFlags.Component)
			}
			if len(args) > 1 {
				return fmt.Errorf("use --names to specify more than one test")
			}
			opts.nameFlags.Names = append(opts.nameFlags.Names, args...)
			if len(opts.nameFlags.Names) == 0 {
				return fmt.Errorf("must specify at least one test")
			}

			specs, err := ext.FindSpecsByName(opts.nameFlags.Names...)
			if err != nil {
				return err
			}

			w, err := extensiontests.NewResultWriter(os.Stdout, extensiontests.ResultFormat(opts.outputFlags.Output))
			if err != nil {
				return err
			}
			defer w.Flush()

			return specs.Run(w)
		},
	}
	opts.componentFlags.BindFlags(cmd.Flags())
	opts.nameFlags.BindFlags(cmd.Flags())
	opts.outputFlags.BindFlags(cmd.Flags())

	return cmd
}
