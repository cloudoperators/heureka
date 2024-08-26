// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/cloudoperators/heureka/scanners/k8s-assets/processor"
	"github.com/cloudoperators/heureka/scanners/k8s-assets/scanner"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Config struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"debug" required:"true" json:"-"`
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	var cfg Config
	err := envconfig.Process("heureka", &cfg)
	if err != nil {
		log.WithError(err).Fatal("Error while reading env config")
	}

	level, err := log.ParseLevel(cfg.LogLevel)

	if err != nil {
		log.WithError(err).Fatal("Error while parsing log level")
	}

	// Only log the warning severity or above.
	log.SetLevel(level)
}

func oidcBasedConfig() (*rest.Config, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		//replace path with a kubeconfig that has a valid oidc token for your cluster
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: path.Join(homedir.HomeDir(), "Library", "Application Support", "SAPCC", "u8s", ".kube", "config")},
		//replace with the context you want to use
		&clientcmd.ConfigOverrides{CurrentContext: "qa-de-1"},
	).ClientConfig()

	if err != nil {
		return nil, err
	}
	return config, nil
}

func main() {
	var scannerCfg scanner.Config
	err := envconfig.Process("heureka", &scannerCfg)
	if err != nil {
		log.WithFields(log.Fields{
			"errror": err,
		}).Warn("Couldn't initialize scanner config")
	}
	// Configure new k8s scanner
	kubeConfig, err := oidcBasedConfig()
	if err != nil {
		log.WithError(err).Fatal("couldn't load kubeConfig")
	}

	// create k8s client
	k8sClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.WithError(err).Fatal("couldn't create k8sClient")
	}

	// Create new k8s scanner
	scanner := scanner.NewScanner(kubeConfig, k8sClient, scannerCfg)

	// Create new processor
	var cfg processor.Config
	err = envconfig.Process("heureka", &cfg)
	if err != nil {
		log.WithError(err).Fatal("Error while reading env config")
	}
	processor := processor.NewProcessor(cfg)

	// Get namespaces
	namespaces, err := scanner.GetNamespaces(v1.ListOptions{})
	if err != nil {
		log.WithError(err).Fatal("no namespaces available")
	}

	// For each namespace fetch list of pods
	for _, n := range namespaces {
		pods, err := scanner.GetPodsByNamespace(n.Name, v1.ListOptions{})
		if err != nil {
			log.WithFields(log.Fields{
				"error":     err,
				"namespace": n.Name,
			}).Error("cannot get pods for namespace")
		}

		// Process pods
		for _, p := range pods {
			podInfo := scanner.GetPodInfo(p)
			serviceInfo := scanner.GetServiceInfo(podInfo)

			// Process service
			serviceId, err := processor.ProcessService(context.Background(), n.Name, serviceInfo)
			if err != nil {
				log.WithFields(log.Fields{
					"err":         err,
					"serviceInfo": serviceInfo,
					"podInfo":     podInfo,
				}).Error("Couldn't process service")
			}

			// Process pod
			err = processor.ProcessPod(context.Background(), n.Name, serviceId, podInfo)
			if err != nil {
				log.WithFields(log.Fields{
					"err":     err,
					"podInfo": podInfo,
				}).Error("Couldn't process pod")
			}

			fmt.Printf("PodInfo: %#v\n", podInfo)
			// fmt.Printf("serviceInfo: %#v\n", serviceInfo)

			// fmt.Printf("\n\n")
		}

		log.Infof("Number of pods (%s): %d", n.Name, len(pods))
	}
}
