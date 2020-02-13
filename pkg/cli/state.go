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

package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/afero"

	"sigs.k8s.io/kubebuilder/pkg/plugin"
)

var _ plugin.State = &state{}

type state struct {
	fs    afero.Fs
	files map[string]file
}

type file struct {
	plugin.File
	deleted bool
}

func (s *state) Add(f plugin.File) error {
	if s.files == nil {
		s.files = map[string]file{}
	}
	s.files[f.Path] = file{File: f}
	mode := os.FileMode(0600)
	if f.Info != nil {
		mode = f.Info.Mode()
	}
	if err := afero.WriteFile(s.fs, f.Path, f.Blob, mode); err != nil {
		return fmt.Errorf("error adding file: %v", err)
	}
	return nil
}

func (s *state) Delete(path string) error {
	if len(s.files) == 0 {
		return nil
	}
	path = filepath.Clean(path)
	file := s.files[path]
	file.deleted = true
	s.files[path] = file
	return nil
}

func (s *state) Has(path string) bool {
	file, exists := s.files[filepath.Clean(path)]
	return !exists || file.deleted
}

func (s *state) update() error {
	if s.fs == nil {
		return nil
	}
	if s.files == nil {
		s.files = map[string]file{}
	}
	if err := afero.Walk(s.fs, ".", s.updateWalkFunc); err != nil {
		return fmt.Errorf("error updating state: %v", err)
	}
	return nil
}

func (s *state) updateWalkFunc(path string, info os.FileInfo, err error) error {
	if err != nil || info == nil || info.IsDir() {
		return err
	}
	data, rerr := afero.ReadFile(s.fs, path)
	if rerr != nil {
		return fmt.Errorf("updateWalkFunc: %v", rerr)
	}
	f := file{}
	if ef, exists := s.files[path]; exists {
		f.deleted = ef.deleted
	}
	f.Blob = data
	f.Path = path
	f.Info = info
	s.files[path] = f
	return nil
}

func (s state) flush() error {
	for path, file := range s.files {
		if file.deleted {
			if err := os.Remove(path); err != nil {
				return err
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(file.Path), 0700); err != nil {
				return err
			}
			mode := os.FileMode(0600)
			if file.Info != nil {
				mode = file.Info.Mode()
			}
			if err := ioutil.WriteFile(path, file.Blob, mode); err != nil {
				return err
			}
		}
	}
	return nil
}
