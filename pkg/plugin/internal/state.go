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

package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"

	"sigs.k8s.io/kubebuilder/pkg/plugin"
)

func NewStateFromFSWalkFunc(state plugin.State, fs afero.Fs) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return err
		}
		fmt.Println("path:", path)
		data, rerr := afero.ReadFile(fs, path)
		if rerr != nil {
			return fmt.Errorf("error adding file to state: %v", rerr)
		}
		file := plugin.File{}
		file.Blob = data
		file.Path = path
		file.Info = info
		return state.Add(file)
	}
}
