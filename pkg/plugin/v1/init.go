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

package v1

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"

	"sigs.k8s.io/kubebuilder/internal/cmdutil"
	"sigs.k8s.io/kubebuilder/internal/config"
	"sigs.k8s.io/kubebuilder/pkg/plugin"
	"sigs.k8s.io/kubebuilder/pkg/plugin/internal"
	"sigs.k8s.io/kubebuilder/pkg/scaffold"
)

type initPlugin struct { // nolint:maligned
	config *config.Config

	// boilerplate options
	license string
	owner   string

	// deprecated flags
	depFlag *pflag.Flag
	depArgs []string
	dep     bool

	// flags
	fetchDeps          bool
	skipGoVersionCheck bool
}

var (
	_ plugin.Init        = &initPlugin{}
	_ cmdutil.RunOptions = &initPlugin{}
)

func (p initPlugin) UpdateContext(ctx *plugin.Context) {
	ctx.Description = `Initialize a new project including vendor/ directory and Go package directories.

Writes the following files:
- a boilerplate license file
- a PROJECT file with the domain and repo
- a Makefile to build the project
- a go.mod with project dependencies
- a Kustomization.yaml for customizating manifests
- a Patch file for customizing image for manager manifests
- a Patch file for enabling prometheus metrics
- a cmd/manager/main.go to run

project will prompt the user to run 'dep ensure' after writing the project files.
`
	ctx.Examples = fmt.Sprintf(`  # Scaffold a project using the apache2 license with "The Kubernetes authors" as owners
  %s init --project-version=1 --domain example.org --license apache2 --owner "The Kubernetes authors"
`,
		ctx.CommandName)
}

func (p *initPlugin) BindFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&p.skipGoVersionCheck, "skip-go-version-check",
		false, "if specified, skip checking the Go version")

	// dependency args
	fs.BoolVar(&p.fetchDeps, "fetch-deps", true, "ensure dependencies are downloaded")

	// deprecated dependency args
	fs.BoolVar(&p.dep, "dep", true, "if specified, determines whether dep will be used.")
	p.depFlag = fs.Lookup("dep")
	fs.StringArrayVar(&p.depArgs, "depArgs", nil, "additional arguments for dep")

	if err := fs.MarkDeprecated("dep", "use the fetch-deps flag instead"); err != nil {
		log.Printf("error to mark dep flag as deprecated: %v", err)
	}
	if err := fs.MarkDeprecated("depArgs", "will be removed with version 1 scaffolding"); err != nil {
		log.Printf("error to mark dep flag as deprecated: %v", err)
	}

	// boilerplate args
	fs.StringVar(&p.license, "license", "apache2",
		"license to use to boilerplate, may be one of 'apache2', 'none'")
	fs.StringVar(&p.owner, "owner", "", "owner to add to the copyright")

	// project args
	if p.config == nil {
		p.config = config.New(config.DefaultPath)
	}
	fs.StringVar(&p.config.Repo, "repo", "", "name to use for go module (e.g., github.com/user/repo), "+
		"defaults to the go package of the current working directory.")
	fs.StringVar(&p.config.Domain, "domain", "my.domain", "domain for groups")
}

func (p *initPlugin) Run() error {
	return cmdutil.Run(p)
}

func (p *initPlugin) SetVersion(v string) {
	p.config.Version = v
}

func (p *initPlugin) LoadConfig() (*config.Config, error) {
	_, err := config.Read()
	if err == nil || os.IsExist(err) {
		return nil, errors.New("already initialized")
	}
	return p.config, nil
}

func (p *initPlugin) Validate(c *config.Config) error {
	// Requires go1.11+
	if !p.skipGoVersionCheck {
		if err := internal.ValidateGoVersion(); err != nil {
			return err
		}
	}

	// Check if the project name is a valid namespace according to k8s
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error to get the current path: %v", err)
	}
	projectName := filepath.Base(dir)
	if err := internal.IsDNS1123Label(strings.ToLower(projectName)); err != nil {
		return fmt.Errorf("project name (%s) is invalid: %v", projectName, err)
	}

	// Try to guess repository if flag is not set
	if c.Repo == "" {
		repoPath, err := internal.FindCurrentRepo()
		if err != nil {
			return fmt.Errorf("error finding current repository: %v", err)
		}
		c.Repo = repoPath
	}

	// Verify dep is installed
	if _, err := exec.LookPath("dep"); err != nil {
		return fmt.Errorf("dep is not installed: %v\n"+
			"Follow steps at: https://golang.github.io/dep/docs/installation.html", err)
	}

	return nil
}

func (p *initPlugin) GetScaffolder(c *config.Config) (scaffold.Scaffolder, error) { // nolint:unparam
	return scaffold.NewInitScaffolder(c, p.license, p.owner), nil
}

func (p *initPlugin) PostScaffold(_ *config.Config) error {
	if !p.depFlag.Changed {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Run `dep ensure` to fetch dependencies (Recommended) [y/n]?")
		p.dep = internal.YesNo(reader)
	}
	if !p.dep {
		fmt.Println("Skipping fetching dependencies.")
		return nil
	}

	err := internal.RunCmd("Fetching dependencies", "dep", append([]string{"ensure"}, p.depArgs...)...)
	if err != nil {
		return err
	}

	err = internal.RunCmd("Running make", "make")
	if err != nil {
		return err
	}

	fmt.Println("Next: define a resource with:\n$ kubebuilder create api")
	return nil
}
