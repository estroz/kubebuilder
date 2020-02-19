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

	"sigs.k8s.io/kubebuilder/pkg/model"
	"sigs.k8s.io/kubebuilder/pkg/model/config"
	"sigs.k8s.io/kubebuilder/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/pkg/plugin"
	"sigs.k8s.io/kubebuilder/pkg/scaffold/input"
	managerv1 "sigs.k8s.io/kubebuilder/pkg/scaffold/v1/manager"
	webhookv1 "sigs.k8s.io/kubebuilder/pkg/scaffold/v1/webhook"
	scaffoldv2 "sigs.k8s.io/kubebuilder/pkg/scaffold/v2"
	webhookv2 "sigs.k8s.io/kubebuilder/pkg/scaffold/v2/webhook"
)

type webhookScaffolder struct {
	config   *config.Config
	resource *resource.Resource
	// v1
	server      string
	webhookType string
	operations  []string
	// v2
	defaulting, validation, conversion bool
	// plugins is the list of plugins we should allow to transform our generated scaffolding
	plugins []plugin.GenericSubcommand
}

func NewV1WebhookScaffolder(
	config *config.Config,
	resource *resource.Resource,
	server string,
	webhookType string,
	operations []string,
	plugins ...plugin.GenericSubcommand,
) Scaffolder {
	return &webhookScaffolder{
		config:      config,
		resource:    resource,
		server:      server,
		webhookType: webhookType,
		operations:  operations,
		plugins:     plugins,
	}
}

func NewV2WebhookScaffolder(
	config *config.Config,
	resource *resource.Resource,
	defaulting bool,
	validation bool,
	conversion bool,
	plugins ...plugin.GenericSubcommand,
) Scaffolder {
	return &webhookScaffolder{
		config:     config,
		resource:   resource,
		defaulting: defaulting,
		validation: validation,
		conversion: conversion,
		plugins:    plugins,
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

func (s *webhookScaffolder) buildUniverse() (*model.Universe, error) {
	return model.NewUniverse(
		model.WithConfig(s.config),
		// TODO(adirio): missing model.WithBoilerplate[From], needs boilerplate or path
		model.WithResource(s.resource),
	)
}

func (s *webhookScaffolder) scaffoldV1() error {
	universe, err := s.buildUniverse()
	if err != nil {
		return err
	}

	webhookConfig := webhookv1.Config{Server: s.server, Type: s.webhookType, Operations: s.operations}

	return (&Scaffold{Plugins: s.plugins}).Execute(
		universe,
		input.Options{},
		&managerv1.Webhook{},
		&webhookv1.AdmissionHandler{Resource: s.resource, Config: webhookConfig},
		&webhookv1.AdmissionWebhookBuilder{Resource: s.resource, Config: webhookConfig},
		&webhookv1.AdmissionWebhooks{Resource: s.resource, Config: webhookConfig},
		&webhookv1.AddAdmissionWebhookBuilderHandler{Resource: s.resource, Config: webhookConfig},
		&webhookv1.Server{Config: webhookConfig},
		&webhookv1.AddServer{Config: webhookConfig},
	)
}

func (s *webhookScaffolder) scaffoldV2() error {
	if s.config.MultiGroup {
		fmt.Println(filepath.Join("apis", s.resource.Group, s.resource.Version,
			fmt.Sprintf("%s_webhook.go", strings.ToLower(s.resource.Kind))))
	} else {
		fmt.Println(filepath.Join("api", s.resource.Version,
			fmt.Sprintf("%s_webhook.go", strings.ToLower(s.resource.Kind))))
	}

	if s.conversion {
		fmt.Println(`Webhook server has been set up for you.
You need to implement the conversion.Hub and conversion.Convertible interfaces for your CRD types.`)
	}

	universe, err := s.buildUniverse()
	if err != nil {
		return err
	}

	webhookScaffolder := &webhookv2.Webhook{
		Resource:   s.resource,
		Defaulting: s.defaulting,
		Validating: s.validation,
	}
	if err := (&Scaffold{Plugins: s.plugins}).Execute(
		universe,
		input.Options{},
		webhookScaffolder,
	); err != nil {
		return err
	}

	universe, err = s.buildUniverse()
	if err != nil {
		return err
	}
	if err := (&scaffoldv2.Main{}).Update(
		universe,
		&scaffoldv2.MainUpdateOptions{
			Config:         s.config,
			WireResource:   false,
			WireController: false,
			WireWebhook:    true,
			Resource:       s.resource,
		},
		s.plugins...,
	); err != nil {
		return fmt.Errorf("error updating main.go: %v", err)
	}

	return nil
}
