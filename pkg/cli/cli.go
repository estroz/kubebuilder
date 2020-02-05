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
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	internalconfig "sigs.k8s.io/kubebuilder/internal/config"
	"sigs.k8s.io/kubebuilder/pkg/cli/internal"
	"sigs.k8s.io/kubebuilder/pkg/plugin"
)

const (
	noticeColor = "\033[1;36m%s\033[0m"
)

// CLI interacts with a command line interface.
type CLI interface {
	// Run runs the CLI, usually returning an error if command line configuration
	// is incorrect.
	Run() error
}

// Option is a function that can configure the cli
type Option func(*cli) error

// cli defines the command line structure and interfaces that are used to
// scaffold kubebuilder project files.
type cli struct {
	commandName    string
	defProjVersion string
	cmd            *cobra.Command
	extraCommands  []*cobra.Command
	plugins        map[string][]plugin.Base
}

// New creates a new cli instance.
func New(opts ...Option) (CLI, error) {
	c := &cli{
		commandName:    "kubebuilder",
		defProjVersion: internalconfig.DefaultVersion,
		plugins:        map[string][]plugin.Base{},
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if err := c.initialize(); err != nil {
		return nil, err
	}
	return c, nil
}

// Run runs the cli.
func (c cli) Run() error {
	return c.cmd.Execute()
}

// WithCommandName is an Option that sets the cli's root command name.
func WithCommandName(name string) Option {
	return func(c *cli) error {
		c.commandName = name
		return nil
	}
}

// WithDefaultProjectVersion is an Option that sets the cli's default project
// version. Setting an unknown version will result in an error.
func WithDefaultProjectVersion(version string) Option {
	return func(c *cli) error {
		c.defProjVersion = version
		return nil
	}
}

// WithPlugins is an Option that sets the cli's plugins.
func WithPlugins(plugins ...plugin.Base) Option {
	return func(c *cli) error {
		for _, p := range plugins {
			ver := p.Version()
			if _, ok := c.plugins[ver]; !ok {
				c.plugins[ver] = []plugin.Base{}
			}
			c.plugins[ver] = append(c.plugins[ver], p)
		}
		return nil
	}
}

// WithExtraCommands is an Option that adds extra subcommands to the cli.
// Adding extra commands that duplicate existing commands results in an error.
func WithExtraCommands(cmds ...*cobra.Command) Option {
	return func(c *cli) error {
		c.extraCommands = append(c.extraCommands, cmds...)
		return nil
	}
}

// initialize initializes the cli.
func (c *cli) initialize() error {
	var projectVersion string
	if internal.Configured() {
		projectVersion, _ = getProjectVersion()
	} else {
		// If the project isn't configured, see if we can find the
		// project version in a command line flag.
		projectVersion, _ = c.getBaseFlags()
	}

	if projectVersion != "" {
		ps, ok := c.plugins[projectVersion]
		if !ok {
			return fmt.Errorf("unknown project version %q", projectVersion)
		}
		for _, p := range ps {
			if dp, ok := p.(plugin.Deprecated); ok {
				fmt.Printf(noticeColor, fmt.Sprintf("[Deprecation Notice] %s\n\n",
					dp.DeprecationWarning()))
			}
		}
	}

	rootCmd := c.defaultCommand()

	rootCmd.AddCommand(
		c.newInitProjectCmd(),
		c.newCreateCmd(),
	)

	if internal.ConfiguredAndV1() {
		rootCmd.AddCommand(
			c.newAlphaCmd(),
			c.newVendorUpdateCmd(),
		)
	}

	for _, cmd := range c.extraCommands {
		for _, subc := range rootCmd.Commands() {
			if cmd.Name() == subc.Name() {
				return fmt.Errorf("command %q already exists", cmd.Name())
			}
		}
		rootCmd.AddCommand(cmd)
	}

	c.cmd = rootCmd
	return nil
}

// defaultCommand results the root command without its subcommands.
func (c cli) defaultCommand() *cobra.Command {
	return &cobra.Command{
		Use:   c.commandName,
		Short: "Development kit for building Kubernetes extensions and tools.",
		Long: fmt.Sprintf(`Development kit for building Kubernetes extensions and tools.

Provides libraries and tools to create new projects, APIs and controllers.
Includes tools for packaging artifacts into an installer container.

Typical project lifecycle:

- initialize a project:

  %s init --domain example.com --license apache2 --owner "The Kubernetes authors"

- create one or more a new resource APIs and add your code to them:

  %s create api --group <group> --version <version> --kind <Kind>

Create resource will prompt the user for if it should scaffold the Resource and / or Controller. To only
scaffold a Controller for an existing Resource, select "n" for Resource. To only define
the schema for a Resource without writing a Controller, select "n" for Controller.

After the scaffold is written, api will run make on the project.
`, c.commandName, c.commandName),
		Example: fmt.Sprintf(`
  # Initialize your project
  %s init --domain example.com --license apache2 --owner "The Kubernetes authors"

  # Create a frigates API with Group: ship, Version: v1beta1 and Kind: Frigate
  %s create api --group ship --version v1beta1 --kind Frigate

  # Edit the API Scheme
  nano api/v1beta1/frigate_types.go

  # Edit the Controller
  nano controllers/frigate_controller.go

  # Install CRDs into the Kubernetes cluster using kubectl apply
  make install

  # Regenerate code and run against the Kubernetes cluster configured by ~/.kube/config
  make run
`, c.commandName, c.commandName),

		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				log.Fatal(err)
			}
		},
	}
}

// cmdErr updates a cobra command to output error information when executed
// or used with the help flag.
func cmdErr(cmd *cobra.Command, err error) {
	cmd.Long = fmt.Sprintf("%s\nNote: %v", cmd.Long, err)
	cmd.RunE = errCmdFunc(err)
}

// errCmdFunc returns a cobra RunE function that returns the provided error
func errCmdFunc(err error) func(*cobra.Command, []string) error {
	return func(*cobra.Command, []string) error {
		return err
	}
}

// getProjectVersion tries to load PROJECT file and returns if the file exist
// and the version string
func getProjectVersion() (string, bool) {
	config, err := internalconfig.Read()
	if err != nil {
		if os.IsNotExist(err) {
			return "", false
		}
		log.Fatalf("failed to read config: %v", err)
	}
	return config.Version, true
}

// getBaseFlags parses the command line arguments, looking for --project-version
// and help. If an error occurs or only --help is set, getBaseFlags returns an
// empty string and true. Otherwise, getBaseFlags returns the project version
// and false.
func (c cli) getBaseFlags() (string, bool) {
	fs := pflag.NewFlagSet("base", pflag.ExitOnError)
	fs.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{UnknownFlags: true}

	var (
		projectVersion string
		help           bool
	)
	fs.StringVar(&projectVersion, "project-version", c.defProjVersion, "project version")
	fs.BoolVarP(&help, "help", "h", false, "print help")

	err := fs.Parse(os.Args[1:])
	doHelp := err != nil || help && !fs.Lookup("project-version").Changed
	if doHelp {
		return "", true
	}
	return projectVersion, false
}
