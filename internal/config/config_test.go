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
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"sigs.k8s.io/kubebuilder/pkg/model/config"
)

func TestSaveReadFrom(t *testing.T) {
	cases := []struct {
		description    string
		config         Config
		extraFieldsObj interface{}
		expConfigStr   string
		wantSaveErr    bool
	}{
		{
			description:    "empty config",
			config:         Config{},
			extraFieldsObj: struct{}{},
			expConfigStr:   "",
			wantSaveErr:    true,
		},
		{
			description:    "empty config with path",
			config:         Config{path: DefaultPath},
			extraFieldsObj: struct{}{},
			expConfigStr:   "",
		},
		{
			description: "config version 1",
			config: Config{
				Config: config.Config{
					Version: config.Version1,
					Repo:    "github.com/example/project",
					Domain:  "example.com",
				},
				path: DefaultPath,
			},
			extraFieldsObj: struct{}{},
			expConfigStr: `domain: example.com
repo: github.com/example/project
version: "1"`,
		},
		{
			description: "config version 2 without extra fields",
			config: Config{
				Config: config.Config{
					Version: config.Version2,
					Repo:    "github.com/example/project",
					Domain:  "example.com",
				},
				path: DefaultPath,
			},
			extraFieldsObj: struct{}{},
			expConfigStr: `domain: example.com
repo: github.com/example/project
version: "2"`,
		},
		{
			description: "config version 2 with extra fields as map",
			config: Config{
				Config: config.Config{
					Version: config.Version2,
					Repo:    "github.com/example/project",
					Domain:  "example.com",
				},
				path: DefaultPath,
			},
			extraFieldsObj: map[string]interface{}{
				"foo": "bar",
			},
			expConfigStr: `domain: example.com
repo: github.com/example/project
version: "2"
foo: bar`,
		},
		{
			description: "config version 2 with extra fields as struct",
			config: Config{
				Config: config.Config{
					Version: config.Version2,
					Repo:    "github.com/example/project",
					Domain:  "example.com",
				},
				path: DefaultPath,
			},
			extraFieldsObj: struct {
				Foo string `json:"foo"`
			}{
				Foo: "bar",
			},
			expConfigStr: `domain: example.com
repo: github.com/example/project
version: "2"
foo: bar`,
		},
	}

	for _, c := range cases {
		// Setup
		if !c.config.IsV1() {
			if err := c.config.EncodeExtraFields(c.extraFieldsObj); err != nil {
				t.Fatalf("%s: %s", c.description, err)
			}
		}
		c.config.fs = afero.NewMemMapFs()

		// Test Save
		err := c.config.Save()
		if err != nil {
			if !c.wantSaveErr {
				t.Errorf("%s: expected MarshalExtraFields to succeed, got error: %s", c.description, err)
			}
			continue
		} else if c.wantSaveErr {
			t.Errorf("%s: expected MarshalExtraFields to fail, got no error", c.description)
			continue
		}
		configBytes, err := afero.ReadFile(c.config.fs, c.config.path)
		if err != nil {
			t.Fatalf("%s: %s", c.description, err)
		}
		if c.expConfigStr != strings.TrimSpace(string(configBytes)) {
			t.Errorf("%s: compare saved configs\nexpected:\n%s\n\nreturned:\n%s", c.description, c.expConfigStr, string(configBytes))
		}

		// Test readFrom
		// If a config is empty, readFrom will set version to Version1 so we want
		// empty configs to have equal values.
		if c.config.Version == "" {
			c.config.Version = config.Version1
		}
		cfg, err := readFrom(c.config.fs, c.config.path)
		if err != nil {
			t.Fatalf("%s: %s", c.description, err)
		}
		if !reflect.DeepEqual(c.config.Config, cfg) {
			t.Errorf("%s: compare read configs\nexpected:\n%#v\n\nreturned:\n%#v", c.description, c.config.Config, cfg)
		}
	}
}
