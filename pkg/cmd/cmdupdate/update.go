package cmdupdate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	"github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
	"github.com/openshift-eng/openshift-tests-extension/pkg/flags"
)

const metadataDirectory = ".openshift-tests-extension"

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

			// Create the metadata directory if it doesn't exist
			if err := os.MkdirAll(metadataDirectory, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", metadataDirectory, err)
			}

			// Define the file path
			metadataPath := filepath.Join(metadataDirectory, fmt.Sprintf("%s.json", ext.Component.Identifier()))

			// Read existing specs
			var oldSpecs extensiontests.ExtensionTestSpecs
			source, err := os.Open(metadataPath)
			if err != nil {
				return fmt.Errorf("failed to open file: %s: %+w", metadataPath, err)
			}
			if err := json.NewDecoder(source).Decode(&oldSpecs); err != nil {
				return fmt.Errorf("failed to decode file: %s: %+w", metadataPath, err)
			}
			allOldNames := oldSpecs.Names()

			newSpecs := ext.GetSpecs()
			allNewNames := newSpecs.Names()

			diff := findMissingNames(allOldNames, allNewNames)
			var missing []string
			if len(diff) > 0 {
				for _, name := range diff {
					res, err := newSpecs.Filter([]string{fmt.Sprintf(`other_names.exists(n, n == "%s")`, name)})
					if err != nil {
						return err
					}
					if len(res) == 0 {
						missing = append(missing, name)
					}
				}
			}

			if len(missing) > 0 {
				fmt.Fprintf(os.Stderr, "Missing Tests:\n")
				for _, name := range missing {
					fmt.Fprintf(os.Stdout, "  * %s\n", name)
				}
				fmt.Fprintf(os.Stderr, "\n")

				return fmt.Errorf("missing tests, if you've renamed tests you must add their names to OtherNames, " +
					"or mark them obsolete")
			}

			// no missing tests, write the results
			data, err := json.MarshalIndent(newSpecs, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal specs to JSON: %w", err)
			}

			// Write the JSON data to the file
			if err := os.WriteFile(metadataPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", metadataPath, err)
			}

			return nil
		},
	}
	componentFlags.BindFlags(cmd.Flags())
	return cmd
}

// Function to find old names not present in new names
func findMissingNames(allOldNames, allNewNames []string) []string {
	nameExists := make(map[string]bool)
	for _, newName := range allNewNames {
		nameExists[newName] = true
	}

	var missingNames []string
	for _, oldName := range allOldNames {
		if !nameExists[oldName] {
			missingNames = append(missingNames, oldName)
		}
	}

	return missingNames
}
