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

package scaffold

import (
	"github.com/spf13/afero"

	"sigs.k8s.io/kubebuilder/pkg/model"
	"sigs.k8s.io/kubebuilder/pkg/model/config"
	"sigs.k8s.io/kubebuilder/pkg/scaffold/input"
	"sigs.k8s.io/kubebuilder/pkg/scaffold/project"
)

type updateScaffolder struct {
	UpdateOptions
	config *config.Config
}

type UpdateOptions struct {
	Fs afero.Fs
}

func NewUpdateScaffolder(config *config.Config, opts UpdateOptions) Scaffolder {
	if opts.Fs == nil {
		opts.Fs = afero.NewOsFs()
	}
	return &updateScaffolder{
		UpdateOptions: opts,
		config:        config,
	}
}

func (s *updateScaffolder) Scaffold() error {
	universe, err := model.NewUniverse(
		model.WithConfig(s.config),
		model.WithoutBoilerplate,
	)
	if err != nil {
		return err
	}

	return (&Scaffold{}).Execute(
		universe,
		input.Options{Fs: s.Fs},
		&project.GopkgToml{},
	)
}
