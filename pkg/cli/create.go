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

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"sigs.k8s.io/kubebuilder/pkg/cli/internal"
	"sigs.k8s.io/kubebuilder/pkg/plugin"
)

func (c cli) newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Scaffold a Kubernetes API or webhook",
		Long:  `Scaffold a Kubernetes API or webhook.`,
	}
	cmd.AddCommand(c.newCreateAPICmd())

	if !internal.ConfiguredAndV1() {
		cmd.AddCommand(c.newCreateWebhookCmd())
	}

	return cmd
}

func (c cli) newCreateAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Scaffold a Kubernetes API",
		Long: `Scaffold a Kubernetes API.
`,
		RunE: errCmdFunc(
			fmt.Errorf("api subcommand requires an existing project"),
		),
	}

	projectVersion, found := getProjectVersion()
	if !found {
		msg := `For project-specific information, run this command in the root directory of a
project.
`
		cmd.Long = fmt.Sprintf("%s\n%s", cmd.Long, msg)
		return cmd
	}

	// Lookup the plugin for projectVersion and bind it to the command.
	c.bindCreateAPI(cmd, projectVersion)
	return cmd
}

func (c cli) bindCreateAPI(cmd *cobra.Command, projectVersion string) { // nolint:dupl
	ps, ok := c.plugins[projectVersion]
	if !ok {
		err := fmt.Errorf("unknown project version %q", projectVersion)
		cmdErr(cmd, err)
		return
	}
	var getter plugin.CreateAPIPluginGetter
	var hasGetter bool
	for _, p := range ps {
		tmpGetter, isGetter := p.(plugin.CreateAPIPluginGetter)
		if isGetter {
			if hasGetter {
				err := fmt.Errorf("duplicate API creation plugins for project version %q: %s, %s",
					projectVersion, getter.Name(), p.Name())
				cmdErr(cmd, err)
				return
			}
			hasGetter = true
			getter = tmpGetter
		}
	}
	if !hasGetter {
		err := fmt.Errorf("project version %q does not support an API creation plugin",
			projectVersion)
		cmdErr(cmd, err)
		return
	}

	cap := getter.GetCreateAPIPlugin()
	cap.BindFlags(cmd.Flags())
	// TODO: inject defaults.
	ctx := plugin.Context{
		CommandName: c.commandName,
	}
	cap.UpdateContext(&ctx)
	cmd.Long = ctx.Description
	cmd.Example = ctx.Examples
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := cap.Run(); err != nil {
			return fmt.Errorf("failed to create api for project with version %q: %v",
				projectVersion, err)
		}
		return nil
	}
}
