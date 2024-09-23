package flags

import (
	"github.com/spf13/pflag"
)

// EnvironmentFlags contains information for specifying multiple environments.
type EnvironmentFlags struct {
	Environments []string
}

func NewEnvironmentFlags() *EnvironmentFlags {
	return &EnvironmentFlags{
		Environments: []string{},
	}
}

func (f *EnvironmentFlags) BindFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&f.Environments,
		"env",
		f.Environments,
		"specify one or more environments (use --env multiple times for more than one)")
}
