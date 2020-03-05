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

	"sigs.k8s.io/kubebuilder/pkg/plugin"
)

// cmdErr updates a cobra command to output error information when executed
// or used with the help flag.
func cmdErr(cmd *cobra.Command, err error) {
	cmd.Long = fmt.Sprintf("%s\nNote: %v", cmd.Long, err)
	cmd.RunE = errCmdFunc(err)
}

// errCmdFunc returns a cobra RunE function that returns the provided error
func errCmdFunc(err error) func(*cobra.Command, []string) error {
	return func(*cobra.Command, []string) error {
		return err
	}
}

// runECmdFunc returns a cobra RunE function that runs gsub and returns its value.
func runECmdFunc(gsub plugin.GenericSubcommand, msg string) func(*cobra.Command, []string) error { // nolint:interfacer
	return func(*cobra.Command, []string) error {
		if err := gsub.Run(); err != nil {
			return fmt.Errorf("%s: %v", msg, err)
		}
		return nil
	}
}

// noGetterErr returns a nicely formatted error if a plugin Getter
// (ex. InitPluginGetter) does not exist in the set of plugins versionedPlugins.
func (c cli) noGetterErr(versionedPlugins []plugin.Base) error {
	keys := []string{}

	if len(versionedPlugins) == 0 && len(c.cliPluginKeys) != 0 {
		for name, version := range c.cliPluginKeys {
			keys = append(keys, plugin.Key(name, version))
		}
		return fmt.Errorf("no plugins found for CLI plugin keys %+q, likely a naming discrepancy", keys)
	}

	for _, p := range versionedPlugins {
		keys = append(keys, plugin.KeyFor(p))
	}
	return fmt.Errorf("no plugin found for project version %q, possible plugins: %+q", c.projectVersion, keys)
}
