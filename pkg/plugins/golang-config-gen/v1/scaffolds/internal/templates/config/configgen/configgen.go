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

package configgen

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/model/file"
)

var _ file.Template = &ConfigGen{}

// ConfigGen scaffolds a KubebuilderConfigGen file.
type ConfigGen struct {
	file.TemplateMixin
	file.ProjectNameMixin
	file.DomainMixin
	file.MultiGroupMixin
	file.ComponentConfigMixin

	// WithKustomize determines whether this file will be used as a kustomize transformer or directly by kubebuilder.
	// It controls the path of this file and files referenced by it.
	WithKustomize bool

	// Manager image tag.
	Image string
}

// SetTemplateDefaults implements file.Template
func (f *ConfigGen) SetTemplateDefaults() error {
	if f.Path == "" {
		if f.WithKustomize {
			f.Path = filepath.Join("config", "configgen", "kubebuilderconfiggen.yaml")
		} else {
			f.Path = "kubebuilderconfiggen.yaml"
		}
	}

	if f.Image == "" {
		f.Image = fmt.Sprintf("%s/%s:v0.1.0", f.Domain, f.ProjectName)
	}

	f.TemplateBody = fmt.Sprintf(configGenTransformerTemplate,
		file.NewMarkerFor(f.Path, crdName),
	)

	return nil
}

var _ file.Inserter = &ConfigGenUpdater{}

// ConfigGenUpdater updates this scaffold with a resource.
type ConfigGenUpdater struct { //nolint:golint
	file.ResourceMixin

	// WithKustomize determines whether this file will be used as a kustomize transformer or directly by kubebuilder.
	// It controls the path of this file and files referenced by it.
	WithKustomize bool

	path string
}

// GetPath implements file.Builder
func (f *ConfigGenUpdater) GetPath() string {
	if f.path != "" {
		return f.path
	}
	s := ConfigGen{WithKustomize: f.WithKustomize}
	_ = s.SetTemplateDefaults()
	f.path = s.Path
	return f.path
}

// GetIfExistsAction implements file.Builder
func (*ConfigGenUpdater) GetIfExistsAction() file.IfExistsAction {
	return file.Overwrite
}

const (
	crdName = "crdName"
)

const (
	whCRDConvCodeFragment = `  #    %s.%s: false
`
)

// GetMarkers implements file.Inserter
func (f *ConfigGenUpdater) GetMarkers() []file.Marker {
	return []file.Marker{
		file.NewMarkerFor(f.GetPath(), crdName),
		file.NewMarkerFor(f.GetPath(), crdName, "#    "),
	}
}

// GetCodeFragments implements file.Inserter
func (f *ConfigGenUpdater) GetCodeFragments() file.CodeFragmentsMap {
	fragments := make(file.CodeFragmentsMap)

	// Only update if resource is set, which is not the case at init.
	if f.Resource != nil {
		// TODO(estroz): read pluralized name from type marker.
		val := fmt.Sprintf(whCRDConvCodeFragment, f.Resource.Plural, f.Resource.QualifiedGroup())
		fragments[file.NewMarkerFor(f.GetPath(), crdName)] = []string{val}
		fragments[file.NewMarkerFor(f.GetPath(), crdName, "#    ")] = []string{val}
	}

	return fragments
}

const configGenTransformerTemplate = `apiVersion: kubebuilder.sigs.k8s.io/v1alpha1
kind: KubebuilderConfigGen

metadata:
  # name of the project.  used in various resource names.
  # required
  name: {{ if not .WithKustomize }}{{ .ProjectName }}-{{ end }}controller-manager

{{- if not .WithKustomize }}
  # namespace for the project
  # optional -- defaults to "${metadata.name}-system"
  namespace: {{ .ProjectName }}-system
{{- end }}

spec:
  # configure how CRDs are generated
  crds:
    # path to go module source directory provided to controller-gen libraries
    # optional -- defaults to '.'
    sourceDirectory: {{ if .WithKustomize }}../.{{ end }}./api{{ if .MultiGroup}}s{{ end }}/...

  # configure how the controller-manager is generated
  controllerManager:
    # image to run
    image: {{ .Image }}

    # if set, use component config for the controller-manager
    # optional -- defaults to "enable: false"
    componentConfig:
      # use component config
      enable: {{ if .ComponentConfig }}true{{ else }}false{{end}}

      # path to component config to put into a ConfigMap
      configFilepath: controller_manager_config.yaml

    # configure how metrics are exposed
    # uncomment to expose metrics configuration
    # optional -- defaults to not generating metrics configuration
    #metrics:
    #  # disable the auth proxy required for scraping metrics
    #  disable: false
    #
    #  # generate prometheus ServiceMonitor resource
    #  enableServiceMonitor: true

  # configure how webhooks are generated
  # uncomment to expose webhook configuration
  # optional -- defaults to not generating webhook configuration
  #webhooks:
  #  # enable will cause webhook config to be generated
  #  enable: true
  #
  #  # configures crds which use conversion webhooks
  #  enableConversion:
  #    # key is the name of the CRD. For example:
  #    # bars.example.my.domain: false
  #    %[1]s
  #
  #  # configures where to get the certificate used for webhooks
  #  # discriminated union
  #  certificateSource:
  #    # type of certificate source
  #    # one of ["certManager", "dev", "manual"] -- defaults to "manual"
  #    # certManager: certmanager is used to manage certificates -- requires CertManager to be installed
  #    # dev: certificate is generated and wired into resources
  #    # manual: no certificate is generated or wired into resources
  #    type: "dev"
  #
  #    # options for a dev certificate -- requires "dev" as the type
  #    devCertificate:
  #      duration: 1h
`
