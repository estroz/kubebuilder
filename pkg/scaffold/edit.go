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

	"sigs.k8s.io/kubebuilder/internal/config"
)

type editScaffolder struct {
	EditOptions
	config *config.Config
}

type EditOptions struct {
	Fs         afero.Fs
	Multigroup bool
}

func NewEditScaffolder(config *config.Config, opts EditOptions) Scaffolder {
	if opts.Fs == nil {
		opts.Fs = afero.NewOsFs()
	}
	return &editScaffolder{
		EditOptions: opts,
		config:      config,
	}
}

func (s *editScaffolder) Scaffold() error {
	s.config.MultiGroup = s.Multigroup

	return s.config.Save()
}
