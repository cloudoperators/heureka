// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"sync"

	"github.com/cloudoperators/heureka/scanners/keppel/models"
	"github.com/cloudoperators/heureka/scanners/keppel/processor"
	"github.com/cloudoperators/heureka/scanners/keppel/scanner"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func main() {
	var wg sync.WaitGroup
	var scannerCfg scanner.Config
	err := envconfig.Process("heureka", &scannerCfg)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Error while reading env config for scanner")
	}

	var processorCfg processor.Config
	err = envconfig.Process("heureka", &processorCfg)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Error while reading env config for processor")
	}

	keppelScanner := scanner.NewScanner(scannerCfg)
	keppelProcessor := processor.NewProcessor(processorCfg)

	err = keppelScanner.Setup()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Error during scanner setup")
	}
	accounts, err := keppelScanner.ListAccounts()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Error during ListAccounts")
	}

	wg.Add(len(accounts))

	for _, account := range accounts {
		go ProcessAccount(account, keppelScanner, keppelProcessor, &wg)
	}

	wg.Wait()
}

func ProcessAccount(account models.Account, keppelScanner *scanner.Scanner, keppelProcessor *processor.Processor, wg *sync.WaitGroup) error {
	defer wg.Done()
	repositories, err := keppelScanner.ListRepositories(account.Name)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"account:": account.Name,
		}).Error("Error during ProcessRepository")
		return err
	}

	for _, repository := range repositories {
		component, err := keppelProcessor.ProcessRepository(repository)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"account:":   account.Name,
				"repository": repository.Name,
			}).Error("Error during ProcessRepository")
			componentPtr, err := keppelProcessor.GetComponent(repository.Name)
			if err != nil {
				log.WithFields(log.Fields{
					"error":      err,
					"account:":   account.Name,
					"repository": repository.Name,
				}).Error("Error during GetComponent")
			}
			component = *componentPtr
		}

		manifests, err := keppelScanner.ListManifests(account.Name, repository.Name)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"account:":   account.Name,
				"repository": repository.Name,
			}).Error("Error during ListManifests")
			continue
		}
		for _, manifest := range manifests {
			componentVersion, err := keppelProcessor.ProcessManifest(manifest, component.ID)
			if err != nil {
				log.WithFields(log.Fields{
					"error":      err,
					"account:":   account.Name,
					"repository": repository.Name,
				}).Error("Error during ProcessManifest")
				componentVersionPtr, err := keppelProcessor.GetComponentVersion(manifest.Digest)
				if err != nil {
					log.WithFields(log.Fields{
						"error":      err,
						"account:":   account.Name,
						"repository": repository.Name,
					}).Error("Error during GetComponentVersion")
				}
				componentVersion = *componentVersionPtr
			}
			trivyReport, err := keppelScanner.GetTrivyReport(account.Name, repository.Name, manifest.Digest)
			if err != nil {
				log.WithFields(log.Fields{
					"error":      err,
					"account:":   account.Name,
					"repository": repository.Name,
				}).Error("Error during GetTrivyReport")
				continue
			}
			keppelProcessor.ProcessReport(*trivyReport, componentVersion.ID)
		}
	}

	return nil
}
