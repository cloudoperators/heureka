// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

type Config struct {
	// ~/.kube/config
	KubeConfigPath    string `envconfig:"KUBE_CONFIG_PATH" required:"true" json:"-"`
	SupportGroupLabel string `envconfig:"SUPPORT_GROUP_LABEL" required:"true"`
	ServiceNameLabel  string `envconfig:"SERVICE_NAME_LABEL" required:"true"`
}
