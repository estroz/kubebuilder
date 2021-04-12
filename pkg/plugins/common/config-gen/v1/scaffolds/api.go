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
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/common/config-gen/v1/scaffolds/internal/templates/config/configgen"
)

var _ plugins.Scaffolder = &apiScaffolder{}

// apiScaffolder contains configuration for generating scaffolding for Go type
// representing the API and controller that implements the behavior for the API.
type apiScaffolder struct {
	config   config.Config
	resource *resource.Resource

	// fs is the filesystem that will be used by the scaffolder
	fs machinery.Filesystem

	withKustomize bool
}

// NewAPIScaffolder returns a new Scaffolder for API/controller creation operations
func NewAPIScaffolder(config config.Config, res *resource.Resource, withKustomize bool) plugins.Scaffolder {
	return &apiScaffolder{
		config:        config,
		resource:      res,
		withKustomize: withKustomize,
	}
}

func (s *apiScaffolder) InjectFS(fs machinery.Filesystem) { s.fs = fs }

func (s *apiScaffolder) Scaffold() error {

	// Initialize the machinery.Scaffold that will write the files to disk
	scaffold := machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
		machinery.WithResource(s.resource),
	)

	if s.resource.HasAPI() {
		fmt.Println("Creating config-gen files for you to edit...")

		// TODO(estroz): copy sample file.
		// sample := &samples.CRDSample{Force: s.force}
		// if !s.withKustomize {
		// 	if err := sample.SetTemplateDefaults(); err != nil {
		// 		return err
		// 	}
		// 	sample.Path = strings.TrimPrefix(sample.Path, "config"+string(filepath.Separator))
		// }

		if err := scaffold.Execute(
			// Updates conversion CRD name set.
			&configgen.ConfigGenUpdater{WithKustomize: s.withKustomize},
			// sample,
		); err != nil {
			return fmt.Errorf("error scaffolding APIs: %v", err)
		}
	}

	return nil
}
