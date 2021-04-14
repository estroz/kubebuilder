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

package templates

import (
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Inserter = &MakefileUpdater{}

// MakefileUpdater scaffolds a file that defines project management CLI commands
type MakefileUpdater struct {
	// Kustomize version to use in the project
	KustomizeVersion string
}

// GetPath implements machinery.Builder
func (*MakefileUpdater) GetPath() string {
	return "Makefile"
}

// GetIfExistsAction implements machinery.Builder
func (*MakefileUpdater) GetIfExistsAction() machinery.IfExistsAction {
	return machinery.OverwriteFile
}

const (
	deploymentMarker = "deployment"
	toolsMarker      = "tools"
)

// GetMarkers implements machinery.Inserter
func (f *MakefileUpdater) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(f.GetPath(), deploymentMarker),
		machinery.NewMarkerFor(f.GetPath(), toolsMarker),
	}
}

// GetCodeFragments implements machinery.Inserter
func (f *MakefileUpdater) GetCodeFragments() machinery.CodeFragmentsMap {
	return machinery.CodeFragmentsMap{
		machinery.NewMarkerFor(f.GetPath(), deploymentMarker): []string{deploymentFragment},
		machinery.NewMarkerFor(f.GetPath(), toolsMarker):      []string{fmt.Sprintf(toolsFragment, f.KustomizeVersion)},
	}
}

//nolint:lll
const deploymentFragment = `install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

`

const toolsFragment = `KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@%s)

`
