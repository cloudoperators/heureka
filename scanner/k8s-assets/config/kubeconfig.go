// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubeConfigFactory interface
type KubeConfigFactory interface {
	CreateConfig() (*rest.Config, error)
}

// OIDCConfigFactory implements KubeConfigFactory for OIDC-based configs
type OIDCConfigFactory struct {
	path    string
	context string
}

func NewOIDCConfigFactory(path, context string) *OIDCConfigFactory {
	return &OIDCConfigFactory{path: path, context: context}
}

func (f *OIDCConfigFactory) CreateConfig() (*rest.Config, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: f.path},
		&clientcmd.ConfigOverrides{CurrentContext: f.context},
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load OIDC kubeconfig: %w", err)
	}
	return config, nil
}

// InClusterConfigFactory implements KubeConfigFactory for in-cluster configs
type InClusterConfigFactory struct{}

func NewInClusterConfigFactory() *InClusterConfigFactory {
	return &InClusterConfigFactory{}
}

func (f *InClusterConfigFactory) CreateConfig() (*rest.Config, error) {
	return rest.InClusterConfig()
}

// DefaultConfigFactory implements KubeConfigFactory for default configs
type DefaultConfigFactory struct {
	path string
}

func NewDefaultConfigFactory(path string) *DefaultConfigFactory {
	return &DefaultConfigFactory{path: path}
}

func (f *DefaultConfigFactory) CreateConfig() (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", f.path)
	if err != nil {
		return nil, fmt.Errorf("failed to load default kubeconfig: %w", err)
	}
	return config, nil
}

// createConfigFactory creates the appropriate KubeConfigFactory based on the type and parameters
func createConfigFactory(configType, path, context string) (KubeConfigFactory, error) {
	switch configType {
	case "oidc":
		return NewOIDCConfigFactory(path, context), nil
	case "in-cluster":
		return NewInClusterConfigFactory(), nil
	case "default":
		return NewDefaultConfigFactory(path), nil
	default:
		return nil, fmt.Errorf("unknown KUBECONFIG_TYPE: %s", configType)
	}
}

// getKubeConfig is the main function to get the Kubernetes configuration
func GetKubeConfig(configType, path, context string) (*rest.Config, error) {
	factory, err := createConfigFactory(configType, path, context)
	if err != nil {
		return nil, err
	}
	return factory.CreateConfig()
}
