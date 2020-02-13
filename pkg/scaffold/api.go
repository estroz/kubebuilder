/*
Copyright 2019 The Kubernetes Authors.

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
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"sigs.k8s.io/kubebuilder/internal/config"
	"sigs.k8s.io/kubebuilder/pkg/model"
	"sigs.k8s.io/kubebuilder/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/pkg/scaffold/input"
	controllerv1 "sigs.k8s.io/kubebuilder/pkg/scaffold/v1/controller"
	crdv1 "sigs.k8s.io/kubebuilder/pkg/scaffold/v1/crd"
	scaffoldv2 "sigs.k8s.io/kubebuilder/pkg/scaffold/v2"
	controllerv2 "sigs.k8s.io/kubebuilder/pkg/scaffold/v2/controller"
	crdv2 "sigs.k8s.io/kubebuilder/pkg/scaffold/v2/crd"
)

// apiScaffolder contains configuration for generating scaffolding for Go type
// representing the API and controller that implements the behavior for the API.
type apiScaffolder struct {
	APIOptions
	config *config.Config
}

type APIOptions struct {
	Fs       afero.Fs
	Resource *resource.Resource
	// Plugins is the list of plugins we should allow to transform our generated scaffolding
	Plugins []Plugin
	// DoResource indicates whether to scaffold API Resource or not
	DoResource bool
	// DoController indicates whether to scaffold controller files or not
	DoController bool
}

func NewAPIScaffolder(config *config.Config, opts APIOptions) Scaffolder {
	if opts.Fs == nil {
		opts.Fs = afero.NewOsFs()
	}
	return &apiScaffolder{
		APIOptions: opts,
		config:     config,
	}
}

func (s *apiScaffolder) Scaffold() error {
	fmt.Println("Writing scaffold for you to edit...")

	switch {
	case s.config.IsV1():
		return s.scaffoldV1()
	case s.config.IsV2():
		return s.scaffoldV2()
	default:
		return fmt.Errorf("unknown project version %v", s.config.Version)
	}
}

func (s *apiScaffolder) buildUniverse() (*model.Universe, error) {
	return model.NewUniverse(
		model.WithConfig(&s.config.Config),
		// TODO: missing model.WithBoilerplate[From], needs boilerplate or path
		model.WithResource(s.Resource),
	)
}

func (s *apiScaffolder) scaffoldV1() error {
	if s.DoResource {
		fmt.Println(filepath.Join("pkg", "apis", s.Resource.GroupPackageName, s.Resource.Version,
			fmt.Sprintf("%s_types.go", strings.ToLower(s.Resource.Kind))))
		fmt.Println(filepath.Join("pkg", "apis", s.Resource.GroupPackageName, s.Resource.Version,
			fmt.Sprintf("%s_types_test.go", strings.ToLower(s.Resource.Kind))))

		universe, err := s.buildUniverse()
		if err != nil {
			return fmt.Errorf("error building API scaffold: %v", err)
		}

		if err := (&Scaffold{}).Execute(
			universe,
			input.Options{Fs: s.Fs},
			&crdv1.Register{Resource: s.Resource},
			&crdv1.Types{Resource: s.Resource},
			&crdv1.VersionSuiteTest{Resource: s.Resource},
			&crdv1.TypesTest{Resource: s.Resource},
			&crdv1.Doc{Resource: s.Resource},
			&crdv1.Group{Resource: s.Resource},
			&crdv1.AddToScheme{Resource: s.Resource},
			&crdv1.CRDSample{Resource: s.Resource},
		); err != nil {
			return fmt.Errorf("error scaffolding APIs: %v", err)
		}
	} else {
		// disable generation of example reconcile body if not scaffolding resource
		// because this could result in a fork-bomb of k8s resources where watching a
		// deployment, replicaset etc. results in generating deployment which
		// end up generating replicaset, pod etc recursively.
		s.Resource.CreateExampleReconcileBody = false
	}

	if s.DoController {
		fmt.Println(filepath.Join("pkg", "controller", strings.ToLower(s.Resource.Kind),
			fmt.Sprintf("%s_controller.go", strings.ToLower(s.Resource.Kind))))
		fmt.Println(filepath.Join("pkg", "controller", strings.ToLower(s.Resource.Kind),
			fmt.Sprintf("%s_controller_test.go", strings.ToLower(s.Resource.Kind))))

		universe, err := s.buildUniverse()
		if err != nil {
			return fmt.Errorf("error building controller scaffold: %v", err)
		}

		if err := (&Scaffold{}).Execute(
			universe,
			input.Options{Fs: s.Fs},
			&controllerv1.Controller{Resource: s.Resource},
			&controllerv1.AddController{Resource: s.Resource},
			&controllerv1.Test{Resource: s.Resource},
			&controllerv1.SuiteTest{Resource: s.Resource},
		); err != nil {
			return fmt.Errorf("error scaffolding controller: %v", err)
		}
	}

	return nil
}

func (s *apiScaffolder) scaffoldV2() error {
	if s.DoResource {
		// Only save the resource in the config file if it didn't exist
		if s.config.AddResource(s.Resource.GVK()) {
			if err := s.config.Save(); err != nil {
				return fmt.Errorf("error updating project file with resource information : %v", err)
			}
		}

		if s.config.MultiGroup {
			fmt.Println(filepath.Join("apis", s.Resource.Group, s.Resource.Version,
				fmt.Sprintf("%s_types.go", strings.ToLower(s.Resource.Kind))))
		} else {
			fmt.Println(filepath.Join("api", s.Resource.Version,
				fmt.Sprintf("%s_types.go", strings.ToLower(s.Resource.Kind))))
		}

		universe, err := s.buildUniverse()
		if err != nil {
			return fmt.Errorf("error building API scaffold: %v", err)
		}

		if err := (&Scaffold{Plugins: s.Plugins}).Execute(
			universe,
			input.Options{Fs: s.Fs},
			&scaffoldv2.Types{Resource: s.Resource},
			&scaffoldv2.Group{Resource: s.Resource},
			&scaffoldv2.CRDSample{Resource: s.Resource},
			&scaffoldv2.CRDEditorRole{Resource: s.Resource},
			&scaffoldv2.CRDViewerRole{Resource: s.Resource},
			&crdv2.EnableWebhookPatch{Resource: s.Resource},
			&crdv2.EnableCAInjectionPatch{Resource: s.Resource},
		); err != nil {
			return fmt.Errorf("error scaffolding APIs: %v", err)
		}

		universe, err = s.buildUniverse()
		if err != nil {
			return fmt.Errorf("error building kustomization scaffold: %v", err)
		}

		kustomizationFile := &crdv2.Kustomization{Resource: s.Resource}
		if err := (&Scaffold{}).Execute(
			universe,
			input.Options{Fs: s.Fs},
			kustomizationFile,
			&crdv2.KustomizeConfig{},
		); err != nil {
			return fmt.Errorf("error scaffolding kustomization: %v", err)
		}

		if err := kustomizationFile.Update(); err != nil {
			return fmt.Errorf("error updating kustomization.yaml: %v", err)
		}

	} else {
		// disable generation of example reconcile body if not scaffolding resource
		// because this could result in a fork-bomb of k8s resources where watching a
		// deployment, replicaset etc. results in generating deployment which
		// end up generating replicaset, pod etc recursively.
		s.Resource.CreateExampleReconcileBody = false
	}

	if s.DoController {
		if s.config.MultiGroup {
			fmt.Println(filepath.Join("controllers", s.Resource.Group,
				fmt.Sprintf("%s_controller.go", strings.ToLower(s.Resource.Kind))))
		} else {
			fmt.Println(filepath.Join("controllers",
				fmt.Sprintf("%s_controller.go", strings.ToLower(s.Resource.Kind))))
		}

		universe, err := s.buildUniverse()
		if err != nil {
			return fmt.Errorf("error building controller scaffold: %v", err)
		}

		suiteTestFile := &controllerv2.SuiteTest{Resource: s.Resource}
		if err := (&Scaffold{Plugins: s.Plugins}).Execute(
			universe,
			input.Options{Fs: s.Fs},
			suiteTestFile,
			&controllerv2.Controller{Resource: s.Resource},
		); err != nil {
			return fmt.Errorf("error scaffolding controller: %v", err)
		}

		if err := suiteTestFile.Update(); err != nil {
			return fmt.Errorf("error updating suite_test.go under controllers pkg: %v", err)
		}
	}

	if err := (&scaffoldv2.Main{}).Update(
		&scaffoldv2.MainUpdateOptions{
			Fs:             s.Fs,
			Config:         &s.config.Config,
			WireResource:   s.DoResource,
			WireController: s.DoController,
			Resource:       s.Resource,
		},
	); err != nil {
		return fmt.Errorf("error updating main.go: %v", err)
	}

	return nil
}
