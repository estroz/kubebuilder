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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/kubebuilder/pkg/plugin"
)

var _ = Describe("resolvePluginsByKey", func() {
	plugins := makePluginsForKeys(
		"foo.example.com/v1.0",
		"bar.example.com/v1.0",
		"baz.example.com/v1.0",
		"foo.kubebuilder.io/v1.0",
		"foo.kubebuilder.io/v2.0",
		"bar.kubebuilder.io/v1.0",
		"bar.kubebuilder.io/v2.0",
	)

	It("should check key correctly", func() {

		By("Resolving foo.example.com/v1.0")
		resolvedPlugins, err := resolvePluginsByKey(plugins, "foo.example.com/v1.0")
		Expect(err).NotTo(HaveOccurred())
		Expect(getPluginKeys(resolvedPlugins...)).To(Equal([]string{"foo.example.com/v1.0"}))

		By("Resolving foo.example.com")
		resolvedPlugins, err = resolvePluginsByKey(plugins, "foo.example.com")
		Expect(err).NotTo(HaveOccurred())
		Expect(getPluginKeys(resolvedPlugins...)).To(Equal([]string{"foo.example.com/v1.0"}))

		By("Resolving baz")
		resolvedPlugins, err = resolvePluginsByKey(plugins, "baz")
		Expect(err).NotTo(HaveOccurred())
		Expect(getPluginKeys(resolvedPlugins...)).To(Equal([]string{"baz.example.com/v1.0"}))

		By("Resolving foo/v2.0")
		resolvedPlugins, err = resolvePluginsByKey(plugins, "foo/v2.0")
		Expect(err).NotTo(HaveOccurred())
		Expect(getPluginKeys(resolvedPlugins...)).To(Equal([]string{"foo.kubebuilder.io/v2.0"}))

		By("Resolving foo.kubebuilder.io")
		resolvedPlugins, err = resolvePluginsByKey(plugins, "foo.kubebuilder.io")
		Expect(err).NotTo(HaveOccurred())
		Expect(getPluginKeys(resolvedPlugins...)).To(Equal([]string{
			"foo.kubebuilder.io/v1.0",
			"foo.kubebuilder.io/v2.0",
		}))

		By("Resolving foo/v1.0")
		resolvedPlugins, err = resolvePluginsByKey(plugins, "foo/v1.0")
		Expect(err).NotTo(HaveOccurred())
		Expect(getPluginKeys(resolvedPlugins...)).To(Equal([]string{
			"foo.example.com/v1.0",
			"foo.kubebuilder.io/v1.0",
		}))

		By("Resolving foo")
		resolvedPlugins, err = resolvePluginsByKey(plugins, "foo")
		Expect(err).NotTo(HaveOccurred())
		Expect(getPluginKeys(resolvedPlugins...)).To(Equal([]string{
			"foo.example.com/v1.0",
			"foo.kubebuilder.io/v1.0",
			"foo.kubebuilder.io/v2.0",
		}))

		By("Resolving blah")
		_, err = resolvePluginsByKey(plugins, "blah")
		Expect(err).To(MatchError(`ambiguous plugin name "blah": no names match`))

		By("Resolving foo.example.com/v2.0")
		_, err = resolvePluginsByKey(plugins, "foo.example.com/v2.0")
		Expect(err).To(MatchError(`ambiguous plugin name "foo.example.com": no names match`))

		By("Resolving foo/v3.0")
		_, err = resolvePluginsByKey(plugins, "foo/v3.0")
		Expect(err).To(MatchError(`ambiguous plugin version "v3.0": no versions match`))

		By("Resolving foo.example.com/v3.0")
		_, err = resolvePluginsByKey(plugins, "foo.example.com/v3.0")
		Expect(err).To(MatchError(`ambiguous plugin version "v3.0": no versions match`))
	})
})

type mockPlugin struct {
	name, version string
}

func (p mockPlugin) Name() string                     { return p.name }
func (p mockPlugin) Version() string                  { return p.version }
func (mockPlugin) SupportedProjectVersions() []string { return []string{"2"} }

func makeBasePlugin(name, version string) plugin.Base {
	return mockPlugin{name, version}
}

func makePluginsForKeys(keys ...string) (plugins []plugin.Base) {
	for _, key := range keys {
		plugins = append(plugins, makeBasePlugin(plugin.SplitKey(key)))
	}
	return
}

func getPluginKeys(plugins ...plugin.Base) (keys []string) {
	for _, p := range plugins {
		keys = append(keys, plugin.KeyFor(p))
	}
	return
}
