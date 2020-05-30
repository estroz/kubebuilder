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

	"github.com/blang/semver"

	"sigs.k8s.io/kubebuilder/pkg/plugin"
)

// resolvePluginsByKey finds a plugin for pluginKey if it exactly matches
// some form of a known plugin's key. Those forms can be a:
// - Fully qualified key: "go.kubebuilder.io/v2.0.0"
// - Short key: "go/v2.0.0"
// - Fully qualified name: "go.kubebuilder.io"
// - Short name: "go"
// Some of these keys may conflict, ex. the fully-qualified and short names of
// "go.kubebuilder.io/v1.0.0" and "go.kubebuilder.io/v2.0.0" have ambiguous
// unversioned names "go.kubernetes.io" and "go". If pluginKey is ambiguous
// or does not match any known plugin's key, an error is returned.
//
// Note: resolvePluginsByKey returns a slice so initialize() can generalize
// setting default plugins if no pluginKey is set.
func resolvePluginsByKey(versionedPlugins []plugin.Base, pluginKey string) (resolved []plugin.Base, err error) {
	name, version := plugin.SplitKey(pluginKey)
	shortName := plugin.GetShortName(name)

	if version == "" {
		// Case: if plugin key has no version, check all plugin names.
		resolved = versionedPlugins
	} else {
		// Case: if plugin key has version, filter by version.
		resolved = findPluginsMatchingVersion(versionedPlugins, version)
	}

	if len(resolved) == 0 {
		return nil, fmt.Errorf("ambiguous plugin version %q: no versions match", version)
	}

	if name == shortName {
		// Case: if plugin name is short, find matching short names.
		resolved = findPluginsMatchingShortName(resolved, shortName)
	} else {
		// Case: if plugin name is fully-qualified, match only fully-qualified names.
		resolved = findPluginsMatchingName(resolved, name)
	}

	if len(resolved) == 0 {
		return nil, fmt.Errorf("ambiguous plugin name %q: no names match", name)
	}

	return resolved, nil
}

func findPluginsMatchingName(ps []plugin.Base, name string) (equal []plugin.Base) {
	for _, p := range ps {
		if p.Name() == name {
			equal = append(equal, p)
		}
	}
	return equal
}

func findPluginsMatchingShortName(ps []plugin.Base, shortName string) (equal []plugin.Base) {
	for _, p := range ps {
		if plugin.GetShortName(p.Name()) == shortName {
			equal = append(equal, p)
		}
	}
	return equal
}

func findPluginsMatchingVersion(ps []plugin.Base, version string) []plugin.Base {
	// Assume versions have been validated already.
	sv := must(semver.ParseTolerant(version))

	var equal, matchingMajor []plugin.Base
	for _, p := range ps {
		pv := must(semver.ParseTolerant(p.Version()))
		if sv.Major == pv.Major {
			if sv.Minor == pv.Minor {
				equal = append(equal, p)
			} else {
				matchingMajor = append(matchingMajor, p)
			}
		}
	}

	if len(equal) != 0 {
		return equal
	}
	return matchingMajor
}

func must(v semver.Version, err error) semver.Version {
	if err != nil {
		panic(err)
	}
	return v
}
