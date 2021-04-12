/*
Copyright 2021 The Kubernetes Authors.

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

package scaffolds

import (
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"
)

var _ plugins.Scaffolder = &webhookScaffolder{}

type webhookScaffolder struct {
	config   config.Config
	resource *resource.Resource

	// fs is the filesystem that will be used by the scaffolder
	fs machinery.Filesystem
}

// NewWebhookScaffolder returns a new Scaffolder for v2 webhook creation operations
func NewWebhookScaffolder(config config.Config, resource *resource.Resource) plugins.Scaffolder {
	return &webhookScaffolder{
		config:   config,
		resource: resource,
	}
}

func (s *webhookScaffolder) InjectFS(fs machinery.Filesystem) { s.fs = fs }

func (s *webhookScaffolder) Scaffold() error {

	// Delete stuff from go/v3

	return nil
}
