/*
Copyright 2018 The Kubernetes Authors.

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
	ctx := c.newAPIContext()
	cmd := &cobra.Command{
		Use:     "api",
		Short:   "Scaffold a Kubernetes API",
		Long:    ctx.Description,
		Example: ctx.Examples,
		RunE: errCmdFunc(
			fmt.Errorf("api subcommand requires an existing project"),
		),
	}

	// Lookup the plugin for projectVersion and bind it to the command.
	c.bindCreateAPI(ctx, cmd)
	return cmd
}

func (c cli) newAPIContext() plugin.Context {
	ctx := plugin.Context{
		CommandName: c.commandName,
		Description: `Scaffold a Kubernetes API.
`,
	}
	if !c.configured {
		ctx.Description = fmt.Sprintf("%s\n%s", ctx.Description, runInProjectRootMsg)
	}
	return ctx
}

func (c cli) bindCreateAPI(ctx plugin.Context, cmd *cobra.Command) {
	versionedPlugins, err := c.getVersionedPlugins()
	if err != nil {
		cmdErr(cmd, err)
		return
	}
	// TODO(estroz): infer which plugin should be used by config layout
	var getter plugin.CreateAPIPluginGetter
	var foundGetter bool
	for _, p := range versionedPlugins {
		getter, foundGetter = p.(plugin.CreateAPIPluginGetter)
		if foundGetter {
			break
		}
	}
	if !foundGetter {
		cmdErr(cmd, c.noGetterErr(versionedPlugins))
		return
	}

	createAPI := getter.GetCreateAPIPlugin()
	createAPI.BindFlags(cmd.Flags())
	createAPI.UpdateContext(&ctx)
	cmd.Long = ctx.Description
	cmd.Example = ctx.Examples
	cmd.RunE = runECmdFunc(createAPI, fmt.Sprintf("failed to create api for project with version %q", c.projectVersion))
}
