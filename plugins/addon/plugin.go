package addon

import (
	"github.com/spf13/pflag"
	"sigs.k8s.io/kubebuilder/pkg/model"
	"sigs.k8s.io/kubebuilder/pkg/plugin"
)

var (
	_ plugin.Base              = &Plugin{}
	_ plugin.GenericSubcommand = &Plugin{}
)

type Plugin struct{}

func (Plugin) Name() string                       { return "addon" }
func (Plugin) Version() string                    { return "v2.0.0" }
func (Plugin) SupportedProjectVersions() []string { return []string{"2"} }
func (Plugin) UpdateContext(*plugin.Context)      {}
func (Plugin) BindFlags(*pflag.FlagSet)           {}
func (Plugin) PostRun() error                     { return nil }

func (p *Plugin) Run(u *model.Universe) error {
	functions := []PluginFunc{
		ExampleManifest,
		ExampleChannel,
		ReplaceController,
		ReplaceTypes,
	}

	for _, fn := range functions {
		if err := fn(u); err != nil {
			return err
		}
	}

	return nil
}
