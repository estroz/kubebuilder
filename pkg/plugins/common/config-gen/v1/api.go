/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"errors"
	"fmt"

	"github.com/spf13/pflag"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/common/config-gen/v1/scaffolds"
)

var _ plugin.CreateAPISubcommand = &createAPISubcommand{}

type createAPISubcommand struct {
	config config.Config

	resource *resource.Resource
}

func (p *createAPISubcommand) UpdateMetadata(cliMeta plugin.CLIMetadata, subcmdMeta *plugin.SubcommandMetadata) {
	subcmdMeta.Description = `Scaffold a Kubernetes API by writing a Resource definition and/or a Controller.

If information about whether the resource and controller should be scaffolded
was not explicitly provided, it will prompt the user if they should be.

After the scaffold is written, the dependencies will be updated and
make generate will be run.
`
	subcmdMeta.Examples = fmt.Sprintf(`  # Create a frigates API with Group: ship, Version: v1beta1 and Kind: Frigate
  %[1]s create api --group ship --version v1beta1 --kind Frigate

  # Edit the API Scheme
  nano api/v1beta1/frigate_types.go

  # Edit the Controller
  nano controllers/frigate/frigate_controller.go

  # Edit the Controller Test
  nano controllers/frigate/frigate_controller_test.go

  # Install CRDs into the Kubernetes cluster using kubectl apply
  make install

  # Regenerate code and run against the Kubernetes cluster configured by ~/.kube/config
  make run
`, cliMeta.CommandName)
}

func (p *createAPISubcommand) BindFlags(fs *pflag.FlagSet) {}

func (p *createAPISubcommand) InjectConfig(c config.Config) error {
	p.config = c
	return nil
}

func (p *createAPISubcommand) InjectResource(res *resource.Resource) error {
	p.resource = res
	return nil
}

func (p *createAPISubcommand) Scaffold(fs machinery.Filesystem) error {
	cfg := Config{}
	if err := p.config.DecodePluginConfig(pluginName, &cfg); err != nil {
		keyNotFoundErr := config.PluginKeyNotFoundError{}
		if !errors.As(err, &keyNotFoundErr) {
			return err
		}
	}

	scaffolder := scaffolds.NewAPIScaffolder(p.config, p.resource, cfg.WithKustomize)
	scaffolder.InjectFS(fs)
	return scaffolder.Scaffold()
}
