// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"context"
	"runtime"
	"sync"

	"github.com/cloudoperators/heureka/scanners/keppel/client"
	"github.com/cloudoperators/heureka/scanners/keppel/models"
	"github.com/cloudoperators/heureka/scanners/keppel/processor"
	"github.com/cloudoperators/heureka/scanners/keppel/scanner"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"debug" required:"true" json:"-"`
}

// ManifestInfo groups related manifest information
type ManifestInfo struct {
	ComponentID      string
	ComponentVersion *client.ComponentVersion
	Account          string
	Repository       string
}

// ChildManifestInfo groups related child manifest information
type ChildManifestInfo struct {
	Account          string
	Repository       string
	Manifest         models.Manifest
	ComponentID      string
	ComponentVersion *client.ComponentVersion
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
	var scannerCfg scanner.Config
	err := envconfig.Process("heureka", &scannerCfg)
	if err != nil {
		log.WithError(err).Fatal("Error while reading env config for scanner")
	}

	var processorCfg processor.Config
	err = envconfig.Process("heureka", &processorCfg)
	if err != nil {
		log.WithError(err).Fatal("Error while reading env config for processor")
	}

	keppelScanner := scanner.NewScanner(scannerCfg)
	keppelProcessor := processor.NewProcessor(processorCfg, "Keppel")

	err = keppelScanner.Setup()
	if err != nil {
		log.WithError(err).Fatal("Error during scanner setup")
	}

	// Get components and correponding componentVersions
	components, err := keppelProcessor.GetAllComponents(nil, 100)
	if err != nil {
		log.WithError(err).Fatal("cannot list Components")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	keppelProcessor.CreateScannerRun(ctx)

	if err := processConcurrently(ctx, components, keppelScanner, keppelProcessor); err != nil {
		log.WithError(err).Error("Error during concurrent processing")
	}

	keppelProcessor.CompleteScannerRun(ctx)
}

func processConcurrently(
	ctx context.Context,
	components []*client.ComponentAggregate,
	keppelScanner *scanner.Scanner,
	keppelProcessor *processor.Processor,
) error {
	maxWorkers := runtime.GOMAXPROCS(0)
	componentCh := make(chan *client.ComponentAggregate, len(components))
	errors := make(chan error, maxWorkers)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(errors chan error) {
			defer wg.Done()
			for comp := range componentCh {
				select {
				case <-ctx.Done():
					return
				default:
					if err := processComponent(ctx, comp, keppelScanner, keppelProcessor); err != nil {
						errors <- err
					}
				}
			}
		}(errors)
	}

	// Feed components to workers
	for _, comp := range components {
		componentCh <- comp
	}
	close(componentCh)

	// Wait for all workers to finish
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}

func processComponent(
	ctx context.Context,
	comp *client.ComponentAggregate,
	keppelScanner *scanner.Scanner,
	keppelProcessor *processor.Processor,
) error {
	log.Infof("Processing component: %s", comp.Ccrn)

	imageInfo, err := keppelScanner.ExtractImageInfo(comp.Ccrn)
	if err != nil {
		return fmt.Errorf("Cannot extract image information from component ccrn: %w", err)
	}

	for _, cv := range comp.ComponentVersions.Edges {
		manifestInfo := ManifestInfo{
			ComponentID:      comp.Id,
			ComponentVersion: cv.Node,
			Account:          imageInfo.Account,
			Repository:       imageInfo.FullRepository(),
		}

		if err := HandleImageManifests(ctx, manifestInfo, keppelScanner, keppelProcessor); err != nil {
			return err
		}
	}
	return nil
}

func HandleImageManifests(
	ctx context.Context,
	info ManifestInfo,
	keppelScanner *scanner.Scanner,
	keppelProcessor *processor.Processor,
) error {

	log.Info("Handling manifest")
	manifests, err := keppelScanner.GetManifest(info.Account, info.Repository, info.ComponentVersion.Version)
	if err != nil {
		log.WithFields(log.Fields{
			"account:":   info.Account,
			"repository": info.Repository,
		}).WithError(err).Error("Error during GetManifest")
		return fmt.Errorf("Couldn't get manifest: %w", err)
	}

	for _, manifest := range manifests {
		if manifest.VulnerabilityStatus == "Unsupported" {
			log.WithFields(log.Fields{
				"account:":   info.Account,
				"repository": info.Repository,
			}).Warn("Manifest has UNSUPPORTED type: " + manifest.MediaType)
			continue
		}
		if manifest.VulnerabilityStatus == "Clean" {
			log.WithFields(log.Fields{
				"account:":   info.Account,
				"repository": info.Repository,
			}).Info("Manifest has no Vulnerabilities")
			continue
		}

		childInfo := ChildManifestInfo{
			Account:          info.Account,
			Repository:       info.Repository,
			Manifest:         manifest,
			ComponentID:      info.ComponentID,
			ComponentVersion: info.ComponentVersion,
		}
		HandleChildManifests(ctx, childInfo, keppelScanner, keppelProcessor)
	}
	return nil
}

func HandleChildManifests(
	ctx context.Context,
	info ChildManifestInfo,
	keppelScanner *scanner.Scanner,
	keppelProcessor *processor.Processor,
) {
	childManifests, err := keppelScanner.ListChildManifests(info.Account, info.Repository, info.Manifest.Digest)
	if err != nil {
		log.WithFields(log.Fields{
			"account:":   info.Account,
			"repository": info.Repository,
		}).WithError(err).Error("Error during ListChildManifests")
	}

	childManifests = append(childManifests, info.Manifest)
	for _, m := range childManifests {

		// Get Trivy report for a specific repository and image version (componentVersion)
		trivyReport, err := keppelScanner.GetTrivyReport(info.Account, info.Repository, m.Digest)
		if err != nil {
			log.WithFields(log.Fields{
				"account:":   info.Account,
				"repository": info.Repository,
			}).WithError(err).Error("Error during GetTrivyReport")
			return
		}

		if trivyReport == nil {
			return
		}

		keppelProcessor.ProcessReport(*trivyReport, info.ComponentVersion.Id)
	}
}
