package cmdupdate

import (
	"fmt"
	"os"

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

			metadata, err := ext.NewMetadataFromDisk(metadataDirectory)
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			missing, err := metadata.FindRemovedTestsWithoutRename()
			if err != nil && len(missing) > 0 {
				fmt.Fprintf(os.Stderr, "Missing Tests:\n")
				for _, name := range missing {
					fmt.Fprintf(os.Stdout, "  * %s\n", name)
				}
				fmt.Fprintf(os.Stderr, "\n")

				return fmt.Errorf("missing tests, if you've renamed tests you must add their names to OtherNames, " +
					"or mark them obsolete")
			}

			return metadata.WriteToDisk()
		},
	}
	componentFlags.BindFlags(cmd.Flags())
	return cmd
}
