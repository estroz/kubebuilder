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

package scaffold

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"sigs.k8s.io/kubebuilder/pkg/model"
	"sigs.k8s.io/kubebuilder/pkg/model/config"
	"sigs.k8s.io/kubebuilder/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/pkg/scaffold/input"
	managerv1 "sigs.k8s.io/kubebuilder/pkg/scaffold/v1/manager"
	webhookv1 "sigs.k8s.io/kubebuilder/pkg/scaffold/v1/webhook"
	scaffoldv2 "sigs.k8s.io/kubebuilder/pkg/scaffold/v2"
	webhookv2 "sigs.k8s.io/kubebuilder/pkg/scaffold/v2/webhook"
)

type webhookScaffolder struct {
	WebhookV1Options
	WebhookV2Options
	config *config.Config
}

type WebhookV1Options struct {
	Fs          afero.Fs
	Resource    *resource.Resource
	Server      string
	WebhookType string
	Operations  []string
}

func NewV1WebhookScaffolder(config *config.Config, opts WebhookV1Options) Scaffolder {
	if opts.Fs == nil {
		opts.Fs = afero.NewOsFs()
	}
	return &webhookScaffolder{
		WebhookV1Options: opts,
		config:           config,
	}
}

type WebhookV2Options struct {
	Fs                                 afero.Fs
	Resource                           *resource.Resource
	Defaulting, Validation, Conversion bool
}

func NewV2WebhookScaffolder(config *config.Config, opts WebhookV2Options) Scaffolder {
	if opts.Fs == nil {
		opts.Fs = afero.NewOsFs()
	}
	return &webhookScaffolder{
		WebhookV2Options: opts,
		config:           config,
	}
}

func (s *webhookScaffolder) Scaffold() error {
	fmt.Println("Writing scaffold for you to edit...")

	switch {
	case s.config.IsV1():
		return s.scaffoldV1()
	case s.config.IsV2():
		return s.scaffoldV2()
	default:
		return fmt.Errorf("unknown project version %v", s.config.Version)
	}
}

func (s *webhookScaffolder) scaffoldV1() error {
	res := s.WebhookV1Options.Resource
	universe, err := model.NewUniverse(
		model.WithConfig(s.config),
		// TODO(adirio): missing model.WithBoilerplate[From], needs boilerplate or path
		model.WithResource(res),
	)
	if err != nil {
		return err
	}

	webhookConfig := webhookv1.Config{Server: s.Server, Type: s.WebhookType, Operations: s.Operations}

	return (&Scaffold{}).Execute(
		universe,
		input.Options{Fs: s.WebhookV1Options.Fs},
		&managerv1.Webhook{},
		&webhookv1.AdmissionHandler{Resource: res, Config: webhookConfig},
		&webhookv1.AdmissionWebhookBuilder{Resource: res, Config: webhookConfig},
		&webhookv1.AdmissionWebhooks{Resource: res, Config: webhookConfig},
		&webhookv1.AddAdmissionWebhookBuilderHandler{Resource: res, Config: webhookConfig},
		&webhookv1.Server{Config: webhookConfig},
		&webhookv1.AddServer{Config: webhookConfig},
	)
}

func (s *webhookScaffolder) scaffoldV2() error {
	res := s.WebhookV2Options.Resource
	if s.config.MultiGroup {
		fmt.Println(filepath.Join("apis", res.Group, res.Version,
			fmt.Sprintf("%s_webhook.go", strings.ToLower(res.Kind))))
	} else {
		fmt.Println(filepath.Join("api", res.Version,
			fmt.Sprintf("%s_webhook.go", strings.ToLower(res.Kind))))
	}

	if s.Conversion {
		fmt.Println(`Webhook server has been set up for you.
You need to implement the conversion.Hub and conversion.Convertible interfaces for your CRD types.`)
	}

	universe, err := model.NewUniverse(
		model.WithConfig(s.config),
		// TODO(adirio): missing model.WithBoilerplate[From], needs boilerplate or path
		model.WithResource(res),
	)
	if err != nil {
		return err
	}

	webhookScaffolder := &webhookv2.Webhook{
		Resource:   res,
		Defaulting: s.Defaulting,
		Validating: s.Validation,
	}
	if err := (&Scaffold{}).Execute(
		universe,
		input.Options{Fs: s.WebhookV2Options.Fs},
		webhookScaffolder,
	); err != nil {
		return err
	}

	// TODO(estroz): pipe fs here.
	if err := (&scaffoldv2.Main{}).Update(
		&scaffoldv2.MainUpdateOptions{
			Fs:             s.WebhookV2Options.Fs,
			Config:         s.config,
			WireResource:   false,
			WireController: false,
			WireWebhook:    true,
			Resource:       res,
		},
	); err != nil {
		return fmt.Errorf("error updating main.go: %v", err)
	}

	return nil
}
