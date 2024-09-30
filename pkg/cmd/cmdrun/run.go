package cmdrun

import (
	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extension"
)

func NewCommand(registry *extension.Registry) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "run-test NAME",
		Short:        "RunTest a single test by name",
		Long:         "Execute a single test.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}
