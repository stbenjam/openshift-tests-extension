package cmdlist

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	g "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List available tests",
		Long:         "List the available tests in this binary.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			tests := g.ListTests()
			data, err := json.Marshal(tests)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s\n", data)
			return nil
		},
	}

	return cmd
}
