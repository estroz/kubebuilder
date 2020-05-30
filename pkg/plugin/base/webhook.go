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

package base

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/pflag"

	"sigs.k8s.io/kubebuilder/internal/cmdutil"
	"sigs.k8s.io/kubebuilder/pkg/model/config"
	"sigs.k8s.io/kubebuilder/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/pkg/plugin"
	"sigs.k8s.io/kubebuilder/pkg/scaffold"
)

// CreateWebhookPlugin scaffolds a Go webhook definition and manifests for
// a resource.
type CreateWebhookPlugin struct {
	Config *config.Config
	// For help text.
	commandName string

	resource   *resource.Options
	defaulting bool
	validation bool
	conversion bool
}

var (
	_ plugin.CreateWebhook = &CreateWebhookPlugin{}
	_ cmdutil.RunOptions   = &CreateWebhookPlugin{}
)

func (p *CreateWebhookPlugin) UpdateContext(ctx *plugin.Context) {
	ctx.Description = `Scaffold a webhook for an API resource. You can choose to scaffold defaulting,
validating and (or) conversion webhooks.
`
	ctx.Examples = fmt.Sprintf(`  # Create defaulting and validating webhooks for CRD of group crew, version v1
  # and kind FirstMate.
  %s create webhook --group crew --version v1 --kind FirstMate --defaulting --programmatic-validation

  # Create conversion webhook for CRD of group crew, version v1 and kind FirstMate.
  %s create webhook --group crew --version v1 --kind FirstMate --conversion
`,
		ctx.CommandName, ctx.CommandName)

	p.commandName = ctx.CommandName
}

func (p *CreateWebhookPlugin) BindFlags(fs *pflag.FlagSet) {
	p.resource = &resource.Options{}
	fs.StringVar(&p.resource.Group, "group", "", "resource Group")
	fs.StringVar(&p.resource.Version, "version", "", "resource Version")
	fs.StringVar(&p.resource.Kind, "kind", "", "resource Kind")
	fs.StringVar(&p.resource.Plural, "resource", "", "resource Resource")

	fs.BoolVar(&p.defaulting, "defaulting", false,
		"if set, scaffold the defaulting webhook")
	fs.BoolVar(&p.validation, "programmatic-validation", false,
		"if set, scaffold the validating webhook")
	fs.BoolVar(&p.conversion, "conversion", false,
		"if set, scaffold the conversion webhook")
}

func (p *CreateWebhookPlugin) InjectConfig(c *config.Config) {
	p.Config = c
}

func (p *CreateWebhookPlugin) Run() error {
	return cmdutil.Run(p)
}

func (p *CreateWebhookPlugin) Validate() error {
	if err := p.resource.Validate(); err != nil {
		return err
	}

	if !p.defaulting && !p.validation && !p.conversion {
		return fmt.Errorf("%s create webhook requires at least one of --defaulting,"+
			" --programmatic-validation and --conversion to be true", p.commandName)
	}

	return nil
}

func (p *CreateWebhookPlugin) GetScaffolder() (scaffold.Scaffolder, error) {
	// Load the boilerplate
	bp, err := ioutil.ReadFile(filepath.Join("hack", "boilerplate.go.txt")) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("unable to load boilerplate: %v", err)
	}

	// Create the actual resource from the resource options
	res := p.resource.NewResource(p.Config, false)
	return scaffold.NewWebhookScaffolder(p.Config, string(bp), res, p.defaulting, p.validation, p.conversion), nil
}

func (p *CreateWebhookPlugin) PostScaffold() error {
	return nil
}
