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

package cli // nolint:dupl

import (
	"fmt"

	"github.com/spf13/cobra"

	"sigs.k8s.io/kubebuilder/pkg/plugin"
)

func (c *cli) newCreateAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Scaffold a Kubernetes API",
		Long: `Scaffold a Kubernetes API.
`,
		RunE: errCmdFunc(
			fmt.Errorf("api subcommand requires an existing project"),
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
	c.bindCreateAPI(cmd)
	return cmd
}

func (c cli) bindCreateAPI(cmd *cobra.Command) {
	versionedPlugins, err := c.getVersionedPlugins()
	if err != nil {
		cmdErr(cmd, err)
		return
	}
	var getter plugin.CreateAPIPluginGetter
	var hasGetter bool
	for _, p := range versionedPlugins {
		tmpGetter, isGetter := p.(plugin.CreateAPIPluginGetter)
		if isGetter {
			if hasGetter {
				err := fmt.Errorf("duplicate API creation plugins for project version %q: %s, %s",
					c.projectVersion, getter.Name(), p.Name())
				cmdErr(cmd, err)
				return
			}
			hasGetter = true
			getter = tmpGetter
		}
	}
	if !hasGetter {
		err := fmt.Errorf("project version %q does not support an API creation plugin",
			c.projectVersion)
		cmdErr(cmd, err)
		return
	}

	createAPI := getter.GetCreateAPIPlugin()
	createAPI.BindFlags(cmd.Flags())
	// TODO: inject defaults.
	ctx := plugin.Context{
		CommandName: c.commandName,
	}
	createAPI.UpdateContext(&ctx)
	cmd.Long = ctx.Description
	cmd.Example = ctx.Examples
	cmd.RunE = runECmdFunc(createAPI, fmt.Sprintf("failed to create api for project with version %q", c.projectVersion))
}
