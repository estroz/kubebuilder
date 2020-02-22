/*
Copyright 2018 The Kubernetes Authors.

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

package plugin

import (
	"testing"
)

func TestNamesEqual(t *testing.T) {

	tests := []struct {
		name1, name2 string
		wantEqual    bool
	}{
		{"go", "go", true},
		{"go" + defaultNameSuffix, "go", true},
		{"go" + defaultNameSuffix, "go.other.domain", false},
		{"go", "helm", false},
		{"go", "helm" + defaultNameSuffix, false},
	}

	for _, test := range tests {
		equals := DefaultNamesEqual(test.name1, test.name2)
		if equals {
			// got error, but the name isn't invalid
			if !test.wantEqual {
				t.Errorf("plugin name equality check failed: want unequal names %q and %q, but were equal",
					test.name1, test.name2)
			}
		} else {
			// got no error, but the name is invalid
			if test.wantEqual {
				t.Errorf("plugin name equality check failed: want equal names %q and %q, but were unequal",
					test.name1, test.name2)
			}
		}
	}
}
