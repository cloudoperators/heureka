// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"runtime"

	kubeconfig "github.com/cloudoperators/heureka/scanners/k8s-assets/config"
	"github.com/cloudoperators/heureka/scanners/k8s-assets/processor"
	"github.com/cloudoperators/heureka/scanners/k8s-assets/scanner"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

type Config struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"debug" required:"true" json:"-"`
}

type WorkerResult struct {
	Namespace string
	PodCount  int
	Error     error
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

func processNamespace(ctx context.Context, s *scanner.Scanner, p *processor.Processor, namespace string) WorkerResult {
	result := WorkerResult{Namespace: namespace}

	pods, err := s.GetPodsByNamespace(namespace, metav1.ListOptions{})
	if err != nil {
		result.Error = fmt.Errorf("failed to get pods for namespace %s: %w", namespace, err)
		return result
	}

	result.PodCount = len(pods)
	podReplicas := s.GroupPodsByGenerateName(pods)

	for _, podReplica := range podReplicas {
		if len(podReplica.Pods) == 0 {
			continue
		}
		// TODO
		serviceInfo := s.GetServiceInfo(podReplica.Pods[0])

		serviceId, err := p.ProcessService(ctx, serviceInfo)
		if err != nil {
			log.WithFields(log.Fields{
				"error":       err,
				"namespace":   namespace,
				"serviceCcrn": serviceInfo.CCRN,
			}).Error("Failed to process service")
		}

		err = p.ProcessPodReplicaSet(ctx, namespace, serviceId, podReplica)
		if err != nil {
			log.WithFields(log.Fields{
				"error":       err,
				"namespace":   namespace,
				"serviceCcrn": serviceInfo.CCRN,
				"podName":     podReplica.GenerateName,
			}).Error("Failed to process pod")
		}
	}
	return result
}

func processConcurrently(ctx context.Context, s *scanner.Scanner, p *processor.Processor, namespaces []v1.Namespace) (bool, error) {
	var err error
	success := true

	maxConcurrency := runtime.GOMAXPROCS(0)

	// sem is an unbuffered channel meaning that sending onto it will block
	// until there's a corresponding receive operation
	sem := make(chan struct{})
	results := make(chan WorkerResult, len(namespaces))
	var wg sync.WaitGroup

	// Start maxConcurrency number of worker goroutines
	for i := 0; i < maxConcurrency; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					// Context cancelled!
					return
				case sem <- struct{}{}:
					// Go routines will constantly try to send this empty struct to this channel. This will block until
					// there is a corresponding receive operation.
				}
			}
		}()
	}

	// Process namespaces concurrently (in own Go routine). There can be only "maxConcurrency" worker go routines
	// processing data at a given time. Any additional Go routines will be blocked (waiting for a slot to become
	// available)
	for _, ns := range namespaces {
		wg.Add(1)
		go func(namespace string) {
			defer wg.Done()
			<-sem // Wait for an available slot
			result := processNamespace(ctx, s, p, namespace)
			results <- result
		}(ns.Name)
	}

	// Close results channel when all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and process results
	for result := range results {
		if result.Error != nil {
			log.WithFields(log.Fields{
				"namespace": result.Namespace,
				"error":     result.Error,
			}).Error("Failed to process namespace")
			err = result.Error
			success = false
		} else {
			log.WithFields(log.Fields{
				"namespace": result.Namespace,
				"podCount":  result.PodCount,
			}).Info("Successfully processed namespace")
		}
	}
	wg.Wait()
	return success, err
}

func main() {
	var scannerCfg scanner.Config
	err := envconfig.Process("heureka", &scannerCfg)
	if err != nil {
		log.WithFields(log.Fields{
			"errror": err,
		}).Warn("Couldn't initialize scanner config")
	}

	kubeConfig, err := kubeconfig.GetKubeConfig(scannerCfg.KubeconfigType, scannerCfg.KubeConfigPath, scannerCfg.KubeconfigContext)
	if err != nil {
		log.WithError(err).Fatal("couldn't load kubeConfig")
	}

	// Create k8s client
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
	tag := fmt.Sprintf("k8s-assets-%s", cfg.ClusterName)
	processor := processor.NewProcessor(cfg, tag)
	processor.CreateScannerRun(context.Background())

	// Create context with timeout (30min should be ok)
	scanTimeout, err := time.ParseDuration(scannerCfg.ScannerTimeout)
	if err != nil {
		log.WithError(err).Fatal("couldn't parse scanner timeout, setting it to 30 minutes")
		scanTimeout = 30 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	defer cancel()

	// Get namespaces
	namespaces, err := scanner.GetNamespaces(metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Fatal("no namespaces available")
	}

	// Process namespaces concurrently
	ok, err := processConcurrently(ctx, &scanner, processor, namespaces)
	if err != nil {
		log.WithError(err).Fatal("ProcessConcurrently failed")
	}

	log.Info("Finished processing all namespaces")
	if ok {
		processor.CompleteScannerRun(context.Background())
	}

}
