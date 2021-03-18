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

	"github.com/spf13/afero"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/file"
	cgtemplates "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3/scaffolds/golang-config-gen/v1/scaffolds/internal/templates"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3/scaffolds/golang-config-gen/v1/scaffolds/internal/templates/config/configgen"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3/scaffolds/golang-config-gen/v1/scaffolds/internal/templates/config/kdefault"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3/scaffolds/internal/templates"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3/scaffolds/internal/templates/hack"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/internal/cmdutil"
)

const (
	// ControllerRuntimeVersion is the kubernetes-sigs/controller-runtime version to be used in the project
	ControllerRuntimeVersion = "v0.8.3"
	// ControllerToolsVersion is the kubernetes-sigs/controller-tools version to be used in the project
	ControllerToolsVersion = "v0.5.0"
	// KustomizeVersion is the kubernetes-sigs/kustomize version to be used in the project
	KustomizeVersion = "v4.0.5"
)

var _ cmdutil.Scaffolder = &initScaffolder{}

type initScaffolder struct {
	config          config.Config
	boilerplatePath string
	license         string
	owner           string

	// fs is the filesystem that will be used by the scaffolder
	fs machinery.Filesystem

	// Scaffold files with kustomize as the config-gen invoker, not kubebuilder.
	withKustomize bool
}

// NewInitScaffolder returns a new Scaffolder for project initialization operations
func NewInitScaffolder(config config.Config, license, owner string, withKustomize bool) cmdutil.Scaffolder {
	return &initScaffolder{
		config:          config,
		boilerplatePath: hack.DefaultBoilerplatePath,
		license:         license,
		owner:           owner,
		withKustomize:   withKustomize,
	}
}

// InjectFS implements cmdutil.Scaffolder
func (s *initScaffolder) InjectFS(fs machinery.Filesystem) {
	s.fs = fs
}

// Scaffold implements cmdutil.Scaffolder
func (s *initScaffolder) Scaffold() error {
	fmt.Println("Writing scaffold for you to edit...")

	// Initialize the machinery.Scaffold that will write the boilerplate file to disk
	// The boilerplate file needs to be scaffolded as a separate step as it is going to
	// be used by the rest of the files, even those scaffolded in this command call.
	scaffold := machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
	)

	bpFile := &hack.Boilerplate{
		License: s.license,
		Owner:   s.owner,
	}
	bpFile.Path = s.boilerplatePath
	if err := scaffold.Execute(bpFile); err != nil {
		return err
	}

	boilerplate, err := afero.ReadFile(s.fs.FS, s.boilerplatePath)
	if err != nil {
		return err
	}

	// Initialize the machinery.Scaffold that will write the files to disk
	scaffold = machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
		machinery.WithBoilerplate(string(boilerplate)),
	)

	builders := []file.Builder{
		&configgen.ConfigGen{WithKustomize: s.withKustomize},
		&configgen.ControllerManagerConfig{WithKustomize: s.withKustomize},
		&templates.Main{},
		&templates.GoMod{ControllerRuntimeVersion: ControllerRuntimeVersion},
		&templates.GitIgnore{},
		&cgtemplates.Makefile{
			WithKustomize:            s.withKustomize,
			BoilerplatePath:          s.boilerplatePath,
			ControllerToolsVersion:   ControllerToolsVersion,
			KustomizeVersion:         KustomizeVersion,
			ControllerRuntimeVersion: ControllerRuntimeVersion,
		},
		&templates.Dockerfile{},
		&templates.DockerIgnore{},
	}

	if s.withKustomize {
		builders = append(builders,
			&configgen.Kustomization{},
			&kdefault.Kustomization{},
		)
	}

	return scaffold.Execute(builders...)
}
