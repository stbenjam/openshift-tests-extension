package cmdupdate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	"github.com/openshift-eng/openshift-tests-extension/pkg/flags"
)

const metadataDirectory = ".openshift-tests-extension"

// NewUpdateCommand adds an "update" command used to generate and verify the metadata we keep track of. This should
// be a black box to end users, i.e. we can add more criteria later they'll consume when revendoring.  For now,
// we prevent a test to be renamed without updating other names, or a test to be deleted.
func NewUpdateCommand(registry *extension.Registry) *cobra.Command {
	componentFlags := flags.NewComponentFlags()

	cmd := &cobra.Command{
		Use:          "update",
		Short:        "Update test metadata",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ext := registry.Get(componentFlags.Component)
			if ext == nil {
				return fmt.Errorf("couldn't find the component %q", componentFlags.Component)
			}

			if err := os.MkdirAll(metadataDirectory, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", metadataDirectory, err)
			}

			metadataPath := filepath.Join(metadataDirectory, fmt.Sprintf("%s.json", strings.ReplaceAll(ext.Component.Identifier(), ":", "_")))
			newSpecs := ext.GetSpecs()
			data, err := json.MarshalIndent(newSpecs, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal specs to JSON: %w", err)
			}

			if err := os.WriteFile(metadataPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", metadataPath, err)
			}

			fmt.Printf("successfully updated metadata\n")
			return nil
		},
	}
	componentFlags.BindFlags(cmd.Flags())
	return cmd
}
