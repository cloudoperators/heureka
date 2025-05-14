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
	"github.com/cloudoperators/heureka/scanners/keppel/processor"
	"github.com/cloudoperators/heureka/scanners/keppel/scanner"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

var invalidStatuses = map[string]struct{}{
	"Unsupported": {},
	"Pending":     {},
	"Error":       {},
	"Clean":       {},
}

type Config struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"debug" required:"true" json:"-"`
}

// ManifestInfo groups related manifest information
type ManifestInfo struct {
	ComponentID      string
	ComponentVersion *client.ComponentVersion
	Account          string
	Repository       string
	Digest           string
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

	processConcurrently(ctx, components, keppelScanner, keppelProcessor)

	keppelProcessor.CompleteScannerRun(ctx)
}

func processConcurrently(
	ctx context.Context,
	components []*client.ComponentAggregate,
	keppelScanner *scanner.Scanner,
	keppelProcessor *processor.Processor,
) {
	maxWorkers := runtime.GOMAXPROCS(0)
	componentCh := make(chan *client.ComponentAggregate, len(components))
	errorCh := make(chan error, len(components))
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
		}(errorCh)
	}

	// Feed components to workers
	for _, comp := range components {
		componentCh <- comp
	}
	close(componentCh)

	// Wait for all workers to finish
	wg.Wait()
	close(errorCh)

	// Log all errors
	for err := range errorCh {
		log.WithError(err).Error("Error occurred during processing")
	}
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
		return fmt.Errorf("cannot extract image information from component ccrn: %w", err)
	}

	for _, cv := range comp.ComponentVersions.Edges {
		manifestInfo := ManifestInfo{
			ComponentID:      comp.Id,
			ComponentVersion: cv.Node,
			Account:          imageInfo.Account,
			Repository:       imageInfo.FullRepository(),
			Digest:           cv.Node.Version,
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
	manifest, err := keppelScanner.GetManifest(info.Account, info.Repository, info.Digest)
	if err != nil {
		return fmt.Errorf("couldn't get manifest for account: %s, repository: %s: %w", info.Account, info.Repository, err)
	}

	// If manifest contains children, it's a multi-arch image
	// in that case the parent manifest doesn't have a manifest
	if len(manifest.Children) == 0 && !vulnerabilityStatusValid(manifest.VulnerabilityStatus) {
		trivyReport, err := keppelScanner.GetTrivyReport(info.Account, info.Repository, info.Digest)
		if err != nil {
			return fmt.Errorf("couldn't get trivy report for account: %s, repository: %s: %w", info.Account, info.Repository, err)
		}

		if trivyReport == nil {
			return fmt.Errorf("trivy report is nil")
		}

		keppelProcessor.ProcessReport(*trivyReport, info.ComponentVersion.Id)
	}

	for _, childManifest := range manifest.Children {
		// Skip non-amd64 architectures
		if childManifest.Platform.Architecture != "amd64" {
			continue
		}
		childInfo := ManifestInfo{
			Account:          info.Account,
			Repository:       info.Repository,
			ComponentID:      info.ComponentID,
			ComponentVersion: info.ComponentVersion,
			Digest:           childManifest.Digest,
		}
		HandleImageManifests(ctx, childInfo, keppelScanner, keppelProcessor)
	}
	return nil
}

func vulnerabilityStatusValid(status string) bool {
	_, invalid := invalidStatuses[status]
	return !invalid
}
