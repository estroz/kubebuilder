/*
Copyright 2019 The Kubernetes Authors.

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

package cli // nolint:dupl

import (
	"fmt"

	"github.com/spf13/cobra"

	"sigs.k8s.io/kubebuilder/pkg/plugin"
)

func (c cli) newCreateWebhookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "Scaffold a webhook for an API resource",
		Long: `Scaffold a webhook for an API resource.
`,
		RunE: errCmdFunc(
			fmt.Errorf("webhook subcommand requires an existing project"),
		),
	}

	if !c.configured {
		msg := `For project-specific information, run this command in the root directory of a
project.
`
		cmd.Long = fmt.Sprintf("%s\n%s", cmd.Long, msg)
		return cmd
	}

	// Lookup the plugin for projectVersion and bind it to the command.
	c.bindCreateWebhook(cmd)
	return cmd
}

func (c cli) bindCreateWebhook(cmd *cobra.Command) {
	versionedPlugins, err := c.getVersionedPlugins()
	if err != nil {
		cmdErr(cmd, err)
		return
	}
	var getter plugin.CreateWebhookPluginGetter
	var hasGetter bool
	for _, p := range versionedPlugins {
		tmpGetter, isGetter := p.(plugin.CreateWebhookPluginGetter)
		if isGetter {
			if hasGetter {
				err := fmt.Errorf("duplicate webhook creation plugins for project version %q: %s, %s",
					c.projectVersion, getter.Name(), p.Name())
				cmdErr(cmd, err)
				return
			}
			hasGetter = true
			getter = tmpGetter
		}
	}
	if !hasGetter {
		err := fmt.Errorf("project version %q does not support a webhook creation plugin",
			c.projectVersion)
		cmdErr(cmd, err)
		return
	}

	createWebhook := getter.GetCreateWebhookPlugin()
	createWebhook.BindFlags(cmd.Flags())
	// TODO: inject defaults.
	ctx := plugin.Context{
		CommandName: c.commandName,
	}
	createWebhook.UpdateContext(&ctx)
	cmd.Long = ctx.Description
	cmd.Example = ctx.Examples
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if !c.configured {
			return fmt.Errorf("failed to create webhook because project is not initialized")
		}
		if err := createWebhook.Run(); err != nil {
			return fmt.Errorf("failed to create webhook for project with version %q: %v",
				c.projectVersion, err)
		}
		return nil
	}
}
