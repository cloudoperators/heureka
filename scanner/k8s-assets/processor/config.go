// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

type Config struct {
	HeurekaUrl  string `envconfig:"HEUREKA_URL" required:"true" json:"-"`
	ClusterName string `envconfig:"HEUREKA_CLUSTER_NAME" required:"true" json:"-"`
	RegionName  string `envconfig:"HEUREKA_CLUSTER_REGION" required:"true" json:"-"`
}
