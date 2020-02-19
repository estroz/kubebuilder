/*
Copyright 2019 The Kubernetes Authors.

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
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/imports"
	"sigs.k8s.io/kubebuilder/pkg/model"
	"sigs.k8s.io/kubebuilder/pkg/plugin"
	"sigs.k8s.io/kubebuilder/pkg/scaffold/input"
)

type updateFunc func(path string, contents []byte) ([]byte, error)

// Update updates a file in a universe with a file-specific update function fn,
// then executes a set of plugins on that universe, then writes the file to
// disk. Update emulates Scaffold.Execute() behavior.
func Update(universe *model.Universe, path string, fn updateFunc, plugins ...plugin.GenericSubcommand) (err error) {
	var universeFile *model.File
	for _, file := range universe.Files {
		if file.Path == path {
			universeFile = file
			break
		}
	}
	if universeFile == nil {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		universeFile = &model.File{
			Path:           path,
			Contents:       string(contents),
			IfExistsAction: input.Overwrite,
		}
		universe.Files = append(universe.Files, universeFile)
	}

	contents, err := fn(path, []byte(universeFile.Contents))
	if err != nil {
		return err
	}
	universeFile.Contents = string(contents)

	for _, p := range plugins {
		if err := p.Run(universe); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(path, []byte(universeFile.Contents), os.ModePerm)
}

// insertStrings reads content from given reader and insert string below the
// line containing marker string. So for ex. in insertStrings(r, {'m1':
// [v1], 'm2': [v2]})
// v1 will be inserted below the lines containing m1 string and v2 will be inserted
// below line containing m2 string.
func insertStrings(r io.Reader, markerAndValues map[string][]string) (io.Reader, error) {
	// reader clone is needed since we will be reading twice from the given reader
	buf := new(bytes.Buffer)
	rClone := io.TeeReader(r, buf)

	err := filterExistingValues(rClone, markerAndValues)
	if err != nil {
		return nil, err
	}

	out := new(bytes.Buffer)

	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		line := scanner.Text()

		for marker, vals := range markerAndValues {
			if strings.TrimSpace(line) == strings.TrimSpace(marker) {
				for _, val := range vals {
					_, err := out.WriteString(val)
					if err != nil {
						return nil, err
					}
				}
			}
		}
		_, err := out.WriteString(line + "\n")
		if err != nil {
			return nil, err
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func InsertStrings(path string, contents []byte, markerAndValues map[string][]string) ([]byte, error) {
	f := bufio.NewReader(bytes.NewBuffer(contents))

	r, err := insertStrings(f, markerAndValues)
	if err != nil {
		return nil, err
	}

	contents, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	formattedContent := contents
	if filepath.Ext(path) == ".go" {
		formattedContent, err = imports.Process("", contents, nil)
		if err != nil {
			return nil, err
		}
	}
	return formattedContent, nil
}

// filterExistingValues removes the single-line values that already exists in
// the given reader. Multi-line values are ignore currently simply because we
// don't have a use-case for it.
func filterExistingValues(r io.Reader, markerAndValues map[string][]string) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		for marker, vals := range markerAndValues {
			for i, val := range vals {
				if strings.TrimSpace(line) == strings.TrimSpace(val) {
					markerAndValues[marker] = append(vals[:i], vals[i+1:]...)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
