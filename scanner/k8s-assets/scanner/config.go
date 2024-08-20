// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

type Config struct {
	// ~/.kube/config
	KubeConfigPath string `envconfig:"KUBE_CONFIG_PATH" required:"true" json:"-"`
}
