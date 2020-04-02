/*
Copyright 2017 The Kubernetes Authors.

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

package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

// var needs to be used instead of const as ldflags is used to fill this
// information in the release process
var (
	kubeBuilderVersion      = "unknown"
	kubernetesVendorVersion = "unknown"
	goos                    = "unknown"
	goarch                  = "unknown"
	gitCommit               = "$Format:%H$" // sha1 from git, output of $(git rev-parse HEAD)

	buildDate = "1970-01-01T00:00:00Z" // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)

// Version contains all the information related to the CLI version
type Version struct {
	KubeBuilderVersion string `json:"kubeBuilderVersion"`
	KubernetesVendor   string `json:"kubernetesVendor"`
	GitCommit          string `json:"gitCommit"`
	BuildDate          string `json:"buildDate"`
	GoOs               string `json:"goOs"`
	GoArch             string `json:"goArch"`
}

func getVersion() Version {
	return Version{
		kubeBuilderVersion,
		kubernetesVendorVersion,
		gitCommit,
		buildDate,
		goos,
		goarch,
	}
}

// Print prints the CLI version
func (v Version) Print() {
	fmt.Printf("Version: %#v\n", v)
}

// NewCmd creates a new command that prints the CLI version
func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Print the kubebuilder version",
		Long:    `Print the kubebuilder version`,
		Example: `kubebuilder version`,
		Run:     runVersion,
	}
}

func runVersion(_ *cobra.Command, _ []string) {
	getVersion().Print()
}
