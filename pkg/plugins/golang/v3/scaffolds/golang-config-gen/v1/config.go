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

package v1

import "github.com/spf13/pflag"

// Config is this plugin's config.
type Config struct {
	WithKustomize bool `json:"withKustomize,omitempty"`
}

func (c *Config) BindInitFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&c.WithKustomize, "with-kustomize", false,
		"configure KubebuilderConfigGen as a kustomize transformer")
}
