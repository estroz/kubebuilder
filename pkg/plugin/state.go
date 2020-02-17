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

package plugin

import (
	"os"
	"strings"

	"github.com/spf13/afero"

	"sigs.k8s.io/kubebuilder/internal/config"
)

type Metadata struct {
	Path string `json:"path"`
	info os.FileInfo
}

type File struct {
	Metadata
	Blob []byte `json:"blob"`
}

type State struct {
	Fs    afero.Fs
	Files map[string]File
}

func (u *State) Update() error {
	if u.Files == nil {
		u.Files = map[string]File{}
	}
	return afero.Walk(u.Fs, ".", u.updateWalkFunc)
}

var ignorePathPrefixes = []string{
	".git",
	config.DefaultPath,
}

func (u *State) updateWalkFunc(path string, info os.FileInfo, err error) error {
	if err != nil || info == nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	for _, prefixe := range ignorePathPrefixes {
		if strings.HasPrefix(path, prefixe) {
			return nil
		}
	}
	data, rerr := afero.ReadFile(u.Fs, path)
	if rerr != nil {
		return rerr
	}
	u.Files[path] = File{
		Metadata: Metadata{
			Path: path,
			info: info,
		},
		Blob: data,
	}
	return nil
}
