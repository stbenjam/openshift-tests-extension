package cmdinfo

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	"github.com/openshift-eng/openshift-tests-extension/pkg/flags"
)

func NewCommand(registry *extension.Registry) *cobra.Command {
	componentFlags := flags.NewComponentFlags()

	cmd := &cobra.Command{
		Use:          "info",
		Short:        "Info displays available information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ext := registry.Get(componentFlags.Component)
			if ext == nil {
				return fmt.Errorf("couldn't find the component %q", componentFlags.Component)
			}

			info, err := json.MarshalIndent(ext, "", "    ")
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stdout, "%s\n", string(info))
			return nil
		},
	}
	componentFlags.BindFlags(cmd.Flags())
	return cmd
}
