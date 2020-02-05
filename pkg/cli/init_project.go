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
	"log"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"sigs.k8s.io/kubebuilder/pkg/cli/internal"
	"sigs.k8s.io/kubebuilder/pkg/plugin"
)

func (c cli) newInitProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new project",
		Long: `Initialize a new project.

For further help about a specific project version, set --project-version.
`,
		Run:     func(cmd *cobra.Command, args []string) {},
		Example: c.getInitHelpExamples(),
	}

	// Register --project-version on the dynamically created command
	// so that it shows up in help and does not cause a parse error.
	cmd.Flags().String("project-version", c.defProjVersion,
		fmt.Sprintf("project version, (%s)", strings.Join(c.getAvailableProjectVersions(), ", ")))

	// Pre-parse the project version and help flags so that we can
	// dynamically bind to a plugin's init implementation (or not).
	projectVersion, isHelpOnly := c.getBaseFlags()

	// If only the help flag was set, return the command as is.
	if isHelpOnly {
		return cmd
	}

	// Lookup the plugin for projectVersion and bind it to the command.
	c.bindInit(cmd, projectVersion)
	return cmd
}

func (c cli) getInitHelpExamples() string {
	vs := c.getAvailableProjectVersions()
	var sb strings.Builder

	for _, v := range vs {
		rendered := fmt.Sprintf(`  # Help for initializing a project with version %s
  %s init --project-version=%s -h

`, v, c.commandName, v)
		sb.WriteString(rendered)
	}
	return strings.TrimSuffix(sb.String(), "\n\n")
}

func (c cli) getAvailableProjectVersions() (projectVersions []string) {
	for _, ps := range c.plugins {
		for _, p := range ps {
			// Only return project versions from non-deprecated plugins
			if _, ok := p.(plugin.Deprecated); ok {
				continue
			}
			if _, ok := p.(plugin.Init); !ok {
				continue
			}
			projectVersions = append(projectVersions, strconv.Quote(p.Version()))
		}
	}
	return projectVersions
}

func (c cli) bindInit(cmd *cobra.Command, projectVersion string) {
	ps, ok := c.plugins[projectVersion]
	if !ok {
		log.Fatal(fmt.Errorf("unknown project version %q", projectVersion))
	}
	var getter plugin.InitPluginGetter
	var hasGetter bool
	for _, p := range ps {
		tmpGetter, isGetter := p.(plugin.InitPluginGetter)
		if isGetter {
			if hasGetter {
				log.Fatalf("duplicate initialization plugins for project version %q: %s, %s",
					projectVersion, getter.Name(), p.Name())
			}
			hasGetter = true
			getter = tmpGetter
		}
	}
	if !hasGetter {
		log.Fatalf("project version %q does not support a project initialization plugin",
			projectVersion)
	}

	ip := getter.GetInitPlugin()
	ip.BindFlags(cmd.Flags())
	// TODO: inject defaults.
	ctx := plugin.Context{
		CommandName: c.commandName,
	}
	ip.UpdateContext(&ctx)
	ip.SetVersion(projectVersion)
	cmd.Long = ctx.Description
	cmd.Example = ctx.Examples
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if internal.Configured() {
			return fmt.Errorf("failed to initialize project because project is already initialized")
		}
		if err := ip.Run(); err != nil {
			return fmt.Errorf("failed to initialize project with version %q: %v",
				projectVersion, err)
		}
		return nil
	}
}
