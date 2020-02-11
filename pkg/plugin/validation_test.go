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

func TestCmpNames(t *testing.T) {

	tests := []struct {
		name1, name2 string
		wantEqual    bool
	}{
		{"go", "go", true},
		{"go.kubernetes.io", "go", true},
		{"go.kubernetes.io", "go.other.domain", false},
		{"go", "helm", false},
		{"go", "helm.kubernetes.io", false},
	}

	for _, test := range tests {
		val := CmpNames(test.name1, test.name2)
		if val == 0 {
			// got error, but the name isn't invalid
			if !test.wantEqual {
				t.Errorf("plugin name cmp failed: want equal names %q and %q, got %d", test.name1, test.name2, val)
			}
		} else {
			// got no error, but the name is invalid
			if test.wantEqual {
				t.Errorf("plugin name cmp failed: want unequal names %q and %q but were equal", test.name1, test.name2)
			}
		}
	}

}
