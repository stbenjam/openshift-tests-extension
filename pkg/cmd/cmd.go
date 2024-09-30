package cmd

import (
	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd/cmdinfo"
	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd/cmdlist"
	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd/cmdrun"
	"github.com/openshift-eng/openshift-tests-extension/pkg/extension"
)

func DefaultExtensionCommands(registry *extension.Registry) []*cobra.Command {
	return []*cobra.Command{
		cmdrun.NewCommand(registry),
		cmdlist.NewCommand(registry),
		cmdinfo.NewCommand(registry),
	}
}
