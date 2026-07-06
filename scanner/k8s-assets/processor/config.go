// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	HeurekaUrl            string  `envconfig:"HEUREKA_URL"             required:"true" json:"-"`
	ClusterName           string  `envconfig:"HEUREKA_CLUSTER_NAME"    required:"true" json:"-"`
	RegionName            string  `envconfig:"HEUREKA_CLUSTER_REGION"  required:"true" json:"-"`
	ConfigPath            string  `envconfig:"CONFIG_PATH"             required:"true"          default:"/etc/config"`
	DefaultKeppelRegistry string  `envconfig:"DEFAULT_KEPPEL_REGISTRY" required:"true"`
	HeurekaRateLimit      float64 `envconfig:"HEUREKA_RATE_LIMIT"                      json:"-" default:"100.0"`
	HeurekaRateBurst      int     `envconfig:"HEUREKA_RATE_BURST"                      json:"-" default:"100"`
}

func (c *Config) LoadAdvancedConfig() (*AdvancedConfig, error) {
	return NewAdvancedConfig(c.ConfigPath)
}

// AdvancedConfig Represents the ConfigMap object in the Kubernetes API that is mounted to the
// scanner pod its loaded from
// the Config.ConfigPath. This is a non-requred configuration for additional configurations
// if the file is not found the config is initialized as empty
type AdvancedConfig struct {
	SideCars     []SideCar `yaml:"side_cars"`
	ExcludedPods []string  `yaml:"excluded_pods"`

	sideCarMap      map[string]SideCar
	excludedPodsSet map[string]struct{}
}

// SideCar represents the service re-mapping configuration for a sidecar container within a Pod
type SideCar struct {
	ContainerName string `yaml:"name"`
	ServiceName   string `yaml:"service"`
	SupportGroup  string `yaml:"support_group"`
}

// SideCarMap returns a map of sidecar containers to their respective service re-mapping
// configurations. Its lazy initialized
func (c *AdvancedConfig) SideCarMap() map[string]SideCar {
	if len(c.sideCarMap) == 0 && len(c.SideCars) > 0 {
		c.sideCarMap = make(map[string]SideCar)
		for _, sc := range c.SideCars {
			c.sideCarMap[sc.ContainerName] = sc
		}
	}

	return c.sideCarMap
}

// GetSideCar returns the service re-mapping configuration for a sidecar container within a Pod
func (c *AdvancedConfig) GetSideCar(containerName string) (SideCar, bool) {
	sc, ok := c.SideCarMap()[containerName]
	return sc, ok
}

// ExcludedPodsSet returns a set of excluded pod name prefixes, lazy initialized from ExcludedPods.
func (c *AdvancedConfig) ExcludedPodsSet() map[string]struct{} {
	if len(c.excludedPodsSet) == 0 && len(c.ExcludedPods) > 0 {
		c.excludedPodsSet = make(map[string]struct{}, len(c.ExcludedPods))
		for _, name := range c.ExcludedPods {
			c.excludedPodsSet[name] = struct{}{}
		}
	}

	return c.excludedPodsSet
}

// IsExcludedPod returns true if the given pod generate-name matches an entry in ExcludedPods.
// A trailing dash is stripped before lookup so that Deployment/DaemonSet pods (whose
// generateName ends with "-") match the same entry as standalone pods.
func (c *AdvancedConfig) IsExcludedPod(generateName string) bool {
	_, ok := c.ExcludedPodsSet()[strings.TrimRight(generateName, "-")]
	return ok
}

// NewAdvancedConfig creates a new AdvancedConfig object from a file path
func NewAdvancedConfig(path string) (*AdvancedConfig, error) {
	cfg := &AdvancedConfig{
		SideCars: make(
			[]SideCar,
			0,
		), // Initialize the SideCars slice to avoid nil pointer dereference
		ExcludedPods: make([]string, 0),
	}

	b, err := os.ReadFile(path)
	if err != nil {
		// gracefully handle the case where the file does not exist and just return an empty Config
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	err = yaml.Unmarshal(b, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal advanced config: %w", err)
	}

	return cfg, nil
}
