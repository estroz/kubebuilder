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

package v1

import (
	"sigs.k8s.io/kubebuilder/pkg/plugin"
)

type Plugin struct {
	initPlugin
	createAPIPlugin
	createWebhookPlugin
}

var _ plugin.Base = Plugin{}

func (Plugin) Name() string {
	return "go"
}

func (Plugin) Version() string {
	return "1"
}

var _ plugin.InitPluginGetter = Plugin{}

func (p Plugin) GetInitPlugin() plugin.Init {
	return &p.initPlugin
}

var _ plugin.CreateAPIPluginGetter = Plugin{}

func (p Plugin) GetCreateAPIPlugin() plugin.CreateAPI {
	return &p.createAPIPlugin
}

var _ plugin.CreateWebhookPluginGetter = Plugin{}

func (p Plugin) GetCreateWebhookPlugin() plugin.CreateWebhook {
	return &p.createWebhookPlugin
}

var _ plugin.Deprecated = Plugin{}

func (Plugin) DeprecationWarning() string {
	return `The v1 projects are deprecated and will not be supported beyond Feb 1, 2020.
See how to upgrade your project to v2: https://book.kubebuilder.io/migration/guide.html`
}
