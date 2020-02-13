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

package scaffold

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// FileReadWriter is an io wrapper to read and write files
type FileReadWriter struct {
	Fs afero.Fs
}

// WriteCloser returns a WriteCloser to write to given path
func (rw *FileReadWriter) WriteCloser(path string) (io.Writer, error) {
	if rw.Fs == nil {
		rw.Fs = afero.NewOsFs()
	}
	dir := filepath.Dir(path)
	err := rw.Fs.MkdirAll(dir, 0700)
	if err != nil {
		return nil, err
	}

	fi, err := rw.Fs.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}

	return fi, nil
}

// WriteCloser returns a ReadCloser to read from given path
func (rw *FileReadWriter) ReadCloser(path string) (io.Reader, error) {
	if rw.Fs == nil {
		rw.Fs = afero.NewOsFs()
	}

	fi, err := rw.Fs.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}

	return fi, nil
}

// WriteFile write given content to the file path
func (rw *FileReadWriter) WriteFile(filePath string, content []byte) error {
	f, err := rw.WriteCloser(filePath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %v", filePath, err)
	}

	if c, ok := f.(io.Closer); ok {
		defer func() {
			if err := c.Close(); err != nil {
				log.Fatal(err)
			}
		}()
	}

	_, err = f.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write %s: %v", filePath, err)
	}

	return nil
}

// ReadFile reads content from the file path
func (rw *FileReadWriter) ReadFile(filePath string) ([]byte, error) {
	f, err := rw.ReadCloser(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %v", filePath, err)
	}

	if c, ok := f.(io.Closer); ok {
		defer func() {
			if err := c.Close(); err != nil {
				log.Fatal(err)
			}
		}()
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %v", filePath, err)
	}
	return b, nil
}
