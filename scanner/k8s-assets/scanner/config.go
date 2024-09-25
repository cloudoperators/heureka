// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

type Config struct {
	KubeConfigPath    string `envconfig:"KUBE_CONFIG_PATH" default:"~/.kube/config" required:"true" json:"-"`
	KubeconfigContext string `envconfig:"KUBE_CONFIG_CONTEXT"`
	KubeconfigType    string `envconfig:"KUBE_CONFIG_TYPE" default:"oidc"`
	SupportGroupLabel string `envconfig:"SUPPORT_GROUP_LABEL" default:"ccloud/support-group" required:"true"`
	ServiceNameLabel  string `envconfig:"SERVICE_NAME_LABEL" default:"ccloud/service" required:"true"`
	ScannerTimeout    string `envconfig:"SCANNER_TIMEOUT" default:"30m"`
}
