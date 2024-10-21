// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"context"
	"github.com/cloudoperators/heureka/scanners/keppel/client"
	"github.com/cloudoperators/heureka/scanners/keppel/models"
	"github.com/cloudoperators/heureka/scanners/keppel/processor"
	"github.com/cloudoperators/heureka/scanners/keppel/scanner"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"runtime"
	"sync"
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

func main() {
	// var wg sync.WaitGroup
	var scannerCfg scanner.Config
	err := envconfig.Process("heureka", &scannerCfg)
	if err != nil {
		log.WithError(err).Fatal("Error while reading env config for scanner")
	}

	var processorCfg processor.Config
	// var fqdn = scannerCfg.KeppelFQDN
	err = envconfig.Process("heureka", &processorCfg)
	if err != nil {
		log.WithError(err).Fatal("Error while reading env config for processor")
	}

	keppelScanner := scanner.NewScanner(scannerCfg)
	keppelProcessor := processor.NewProcessor(processorCfg)

	err = keppelScanner.Setup()
	if err != nil {
		log.WithError(err).Fatal("Error during scanner setup")
	}

	// Get components
	components, err := keppelProcessor.GetAllComponents(nil, 100)
	if err != nil {
		log.WithError(err).Fatal("cannot list Components")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	processConcurrently(ctx, components, keppelScanner, keppelProcessor)
}

func processConcurrently(ctx context.Context, components []*client.Component, keppelScanner *scanner.Scanner, keppelProcessor *processor.Processor) {
	maxWorkers := runtime.GOMAXPROCS(0)
	componentCh := make(chan *client.Component, len(components))
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for comp := range componentCh {
				select {
				case <-ctx.Done():
					return
				default:
					processComponent(ctx, comp, keppelScanner, keppelProcessor)
				}
			}
		}()
	}

	// Feed components to workers
	for _, comp := range components {
		componentCh <- comp
	}
	close(componentCh)

	// Wait for all workers to finish
	wg.Wait()
}

func processComponent(ctx context.Context, comp *client.Component, keppelScanner *scanner.Scanner, keppelProcessor *processor.Processor) {
	imageInfo, err := keppelScanner.ExtractImageInfo(comp.Name)
	if err != nil {
		log.WithError(err).Error("Couldn't extract image information from component name")
		return
	}

	log.Infof("Processing component: %s", comp.Name)
	compVersions, err := keppelProcessor.GetComponentVersions(comp.Id)
	if err != nil {
		log.WithError(err).Errorf("couldn't fetch component versions for componentId: %s", comp.Id)
		return
	}

	for _, cv := range compVersions {
		HandleImageManifests(ctx, comp.Id, cv.Id, imageInfo.Account, imageInfo.FullRepository(), keppelScanner, keppelProcessor)
	}
}

func HandleImageManifests(
	ctx context.Context,
	componentId string,
	componentVersionId string,
	account string,
	repository string,
	keppelScanner *scanner.Scanner,
	keppelProcessor *processor.Processor,
) {

	log.Info("Handling manifest")
	manifests, err := keppelScanner.ListManifests(account, repository)
	if err != nil {
		log.WithFields(log.Fields{
			"account:":   account,
			"repository": repository,
		}).WithError(err).Error("Error during ListManifests")
		return
	}

	for _, manifest := range manifests {
		if manifest.VulnerabilityStatus == "Unsupported" {
			log.WithFields(log.Fields{
				"account:":   account,
				"repository": repository,
			}).Warn("Manifest has UNSUPPORTED type: " + manifest.MediaType)
			continue
		}
		if manifest.VulnerabilityStatus == "Clean" {
			log.WithFields(log.Fields{
				"account:":   account,
				"repository": repository,
			}).Info("Manifest has no Vulnerabilities")
			continue
		}
		HandleChildManifests(account, repository, manifest, componentId, componentVersionId, keppelScanner, keppelProcessor)
	}
}

func HandleChildManifests(
	account string,
	repository string,
	manifest models.Manifest,
	componentId string,
	componentVersionId string,
	keppelScanner *scanner.Scanner,
	keppelProcessor *processor.Processor,
) {
	childManifests, err := keppelScanner.ListChildManifests(account, repository, manifest.Digest)

	if err != nil {
		log.WithFields(log.Fields{
			"account:":   account,
			"repository": repository,
		}).WithError(err).Error("Error during ListChildManifests")
	}

	childManifests = append(childManifests, manifest)

	for _, m := range childManifests {
		// Get Trivy report for a specific repository and image version (componentVersion)
		trivyReport, err := keppelScanner.GetTrivyReport(account, repository, m.Digest)
		if err != nil {
			log.WithFields(log.Fields{
				"account:":   account,
				"repository": repository,
			}).WithError(err).Error("Error during GetTrivyReport")
			return
		}

		if trivyReport == nil {
			return
		}

		keppelProcessor.ProcessReport(*trivyReport, componentVersionId)
	}
}
