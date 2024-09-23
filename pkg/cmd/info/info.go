package info

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extensions"
	"github.com/openshift-eng/openshift-tests-extension/pkg/flags"
)

func NewInfoCommand() *cobra.Command {
	componentFlags := flags.NewComponentFlags()

	cmd := &cobra.Command{
		Use:          "info",
		Short:        "Info displays available information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := extensions.NewExtensionRegistry()
			extension := registry.Get(componentFlags.Component)
			if extension != nil {
				return fmt.Errorf("couldn't find the named extension %q", extension)
			}

			info, err := json.MarshalIndent(extension, "", "    ")
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stdout, string(info))
			return nil
		},
	}
	componentFlags.BindFlags(cmd.Flags())
	return cmd
}
