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

package config

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"sigs.k8s.io/yaml"
)

const (
	// Scaffolding versions
	Version1 = "1"
	Version2 = "2"
)

// Config is the unmarshalled representation of the configuration file
type Config struct {
	// Version is the project version, defaults to "1" (backwards compatibility)
	Version string `json:"version,omitempty" mapstructure:"version"`

	// Domain is the domain associated with the project and used for API groups
	Domain string `json:"domain,omitempty" mapstructure:"domain"`

	// Repo is the go package name of the project root
	Repo string `json:"repo,omitempty" mapstructure:"repo"`

	// Resources tracks scaffolded resources in the project
	// This info is tracked only in project with version 2
	Resources []GVK `json:"resources,omitempty" mapstructure:"resources"`

	// Multigroup tracks if the project has more than one group
	MultiGroup bool `json:"multigroup,omitempty" mapstructure:"multigroup"`

	// ExtraFields is an arbitrary YAML blob that can be used by non-kubebuilder
	// plugins for plugin-specific configure.
	ExtraFields map[string]interface{} `json:"-" mapstructure:",remain"`
}

// IsV1 returns true if it is a v1 project
func (config Config) IsV1() bool {
	return config.Version == Version1
}

// IsV2 returns true if it is a v2 project
func (config Config) IsV2() bool {
	return config.Version == Version2
}

// ResourceGroups returns unique groups of scaffolded resources in the project
func (config Config) ResourceGroups() []string {
	groupSet := map[string]struct{}{}
	for _, r := range config.Resources {
		groupSet[r.Group] = struct{}{}
	}

	groups := make([]string, 0, len(groupSet))
	for g := range groupSet {
		groups = append(groups, g)
	}

	return groups
}

// HasResource returns true if API resource is already tracked
// NOTE: this works only for v2, since in v1 resources are not tracked
func (config Config) HasResource(target GVK) bool {
	// Short-circuit v1
	if config.IsV1() {
		return false
	}

	// Return true if the target resource is found in the tracked resources
	for _, r := range config.Resources {
		if r.isEqualTo(target) {
			return true
		}
	}

	// Return false otherwise
	return false
}

// AddResource appends the provided resource to the tracked ones
// It returns if the configuration was modified
// NOTE: this works only for v2, since in v1 resources are not tracked
func (config *Config) AddResource(gvk GVK) bool {
	// Short-circuit v1
	if config.IsV1() {
		return false
	}

	// No-op if the resource was already tracked, return false
	if config.HasResource(gvk) {
		return false
	}

	// Append the resource to the tracked ones, return true
	config.Resources = append(config.Resources, gvk)
	return true
}

// GVK contains information about scaffolded resources
type GVK struct {
	Group   string `json:"group,omitempty" mapstructure:"group"`
	Version string `json:"version,omitempty" mapstructure:"version"`
	Kind    string `json:"kind,omitempty" mapstructure:"kind"`
}

// isEqualTo compares it with another resource
func (r GVK) isEqualTo(other GVK) bool {
	return r.Group == other.Group &&
		r.Version == other.Version &&
		r.Kind == other.Kind
}

func (c Config) Marshal() ([]byte, error) {
	content, err := yaml.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("error marshalling project configuration: %v", err)
	}
	// Empty config strings are "{}" due to the map field.
	if strings.TrimSpace(string(content)) == "{}" {
		content = []byte{}
	}
	// Append extra fields since they're ignored when marshalling, unless the
	// project is v1 which does not support extra fields.
	if !c.IsV1() && len(c.ExtraFields) != 0 {
		extraFieldBytes, err := yaml.Marshal(c.ExtraFields)
		if err != nil {
			return nil, fmt.Errorf("error marshalling project configuration extra fields: %v", err)
		}
		content = append(content, extraFieldBytes...)
	}
	return content, nil
}

func Unmarshal(in []byte, out *Config) error {
	var fields map[string]interface{}
	if err := yaml.Unmarshal(in, &fields); err != nil {
		return fmt.Errorf("error unmarshalling project configuration: %v", err)
	}
	if err := mapstructure.Decode(fields, out); err != nil {
		return fmt.Errorf("error decoding project configuration: %v", err)
	}
	// v1 projects do not support extra fields.
	if out.IsV1() {
		out.ExtraFields = nil
	}
	return nil
}

// EncodeExtraFields encodes extraFieldsObj in c. This method is intended to
// be used for custom configuration objects.
func (c *Config) EncodeExtraFields(extraFieldsObj interface{}) error {
	// Short-circuit v1
	if c.IsV1() {
		return fmt.Errorf("v1 project configs do not have extra fields")
	}
	// Shouldn't expect random objects to have mapstructure tags.
	b, err := yaml.Marshal(extraFieldsObj)
	if err != nil {
		return fmt.Errorf("failed to convert %T object to bytes: %s", extraFieldsObj, err)
	}
	var fields map[string]interface{}
	if err := yaml.Unmarshal(b, &fields); err != nil {
		return err
	}
	if err := mapstructure.Decode(fields, c); err != nil {
		return fmt.Errorf("failed to decode %T object: %s", extraFieldsObj, err)
	}
	return nil
}

// DecodeExtraFields decodes extra fields stored in c into extraFieldsObj. This
// method is intended to be used for custom configuration objects.
func (c Config) DecodeExtraFields(extraFieldsObj interface{}) error {
	// Short-circuit v1
	if c.IsV1() {
		return fmt.Errorf("v1 project configs do not have extra fields")
	}
	if c.ExtraFields == nil {
		c.ExtraFields = make(map[string]interface{})
	}
	b, err := yaml.Marshal(c.ExtraFields)
	if err != nil {
		return fmt.Errorf("failed to convert extra fields object to bytes: %s", err)
	}
	if err := yaml.Unmarshal(b, extraFieldsObj); err != nil {
		return fmt.Errorf("failed to unmarshal extra fields object: %s", err)
	}
	return nil
}
