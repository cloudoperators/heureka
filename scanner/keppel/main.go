// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/cloudoperators/heureka/scanners/keppel/models"
	"github.com/cloudoperators/heureka/scanners/keppel/processor"
	"github.com/cloudoperators/heureka/scanners/keppel/scanner"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
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

	// For each component get all component versions
	for _, comp := range components {
		imageInfo, err := keppelScanner.ExtractImageInfo(comp.Name)
		if err != nil {
			log.WithError(err).Error("Could't extract image information from component name")
		} else {
			fmt.Printf("imageInfo: %#v\n", imageInfo)
		}

		// Get component versions
		compVersions, err := keppelProcessor.GetComponentVersions(comp.Id)
		if err != nil {
			log.WithError(err).Errorf("couldn't fetch component versions for componentId: %s", comp.Id)
		}

		HandleRepository(comp.Id, compVersions[0].Id, imageInfo.Account, imageInfo.FullRepository(), keppelScanner, keppelProcessor)

		// Get trivy report
		for _, cv := range compVersions {
			trivyReport, err := keppelScanner.GetTrivyReport(imageInfo.Account, imageInfo.Repository, cv.Version)
			if err != nil {
				log.WithError(err).Errorf("couldn't fetch trivy report")
				continue
			}
			fmt.Print(trivyReport)
		}

		fmt.Print(len(compVersions))
	}

	// accounts, err := keppelScanner.ListAccounts()
	// if err != nil {
	// 	log.WithError(err).Fatal("Error during ListAccounts")
	// }

	// wg.Add(len(accounts))

	// for _, account := range accounts {
	// 	// go HandleAccount(fqdn, account, keppelScanner, keppelProcessor, &wg)
	// }

	// wg.Wait()
}

func HandleAccount(fqdn string, account models.Account, keppelScanner *scanner.Scanner, keppelProcessor *processor.Processor, wg *sync.WaitGroup) error {
	defer wg.Done()
	repositories, err := keppelScanner.ListRepositories(account.Name)
	fmt.Print(len(repositories))
	if err != nil {
		log.WithFields(log.Fields{
			"account:": account.Name,
		}).WithError(err).Error("Error during listing ProcessRepository")
		return err
	}

	// for _, repository := range repositories {
	// 	HandleRepository(fqdn, account, repository, keppelScanner, keppelProcessor)
	// }

	return nil
}

// HandleRepository does what ???
func HandleRepository(
	componentId string,
	componentVersionId string,
	account string,
	repository string,
	keppelScanner *scanner.Scanner,
	keppelProcessor *processor.Processor,
) {
	// 	componentId, err := keppelProcessor.ProcessRepository(fqdn, account, repository)
	// 	if err != nil {
	// 		log.WithFields(log.Fields{
	// 			"account:":   account.Name,
	// 			"repository": repository.Name,
	// 		}).WithError(err).Error("Error during ProcessRepository")
	// 		componentId, err = keppelProcessor.GetComponent(fmt.Sprintf("%s/%s/%s", fqdn, account.Name, repository.Name))
	// 		if err != nil {
	// 			log.WithFields(log.Fields{
	// 				"account:":   account.Name,
	// 				"repository": repository.Name,
	// 			}).WithError(err).Error("Error during GetComponent")
	// 		}
	// 	}

	manifests, err := keppelScanner.ListManifests(account, repository)
	if err != nil {
		log.WithFields(log.Fields{
			"account:":   account,
			"repository": repository,
		}).WithError(err).Error("Error during ListManifests")
		return
	}
	for _, manifest := range manifests {
		// TODO: What for is this needed?
		// if componentId == nil {
		// 	log.WithFields(log.Fields{
		// 		"account:":   account.Name,
		// 		"repository": repository.Name,
		// 	}).Error("Component not found")
		// 	return
		// }
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
		HandleManifest(account, repository, manifest, componentId, componentVersionId, keppelScanner, keppelProcessor)
	}
}

// HandleManifest does what ???
func HandleManifest(
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

	// NOTE: Not really need because we already have the component versions (???)
	// componentVersion, err := keppelProcessor.ProcessManifest(manifest, componentId)
	// if err != nil {
	// 	log.WithFields(log.Fields{
	// 		"account:":   account.Name,
	// 		"repository": repository.Name,
	// 	}).WithError(err).Error("Error during ProcessManifest")
	// 	componentVersions, err := keppelProcessor.GetComponentVersions(component.Id)

	// 	if err != nil || componentVersion == nil {
	// 		log.WithFields(log.Fields{
	// 			"account:":   account.Name,
	// 			"repository": repository.Name,
	// 		}).WithError(err).Error("Error during GetComponentVersion")
	// 		return
	// 	}
	// }

	childManifests = append(childManifests, manifest)

	for _, m := range childManifests {
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
