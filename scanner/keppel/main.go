// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
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
	var wg sync.WaitGroup
	var scannerCfg scanner.Config
	err := envconfig.Process("heureka", &scannerCfg)
	if err != nil {
		log.WithError(err).Fatal("Error while reading env config for scanner")
	}

	var processorCfg processor.Config
	var fqdn = scannerCfg.KeppelFQDN
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
	accounts, err := keppelScanner.ListAccounts()
	if err != nil {
		log.WithError(err).Fatal("Error during ListAccounts")
	}

	wg.Add(len(accounts))

	for _, account := range accounts {
		go HandleAccount(fqdn, account, keppelScanner, keppelProcessor, &wg)
	}

	wg.Wait()
}

func HandleAccount(fqdn string, account models.Account, keppelScanner *scanner.Scanner, keppelProcessor *processor.Processor, wg *sync.WaitGroup) error {
	defer wg.Done()
	repositories, err := keppelScanner.ListRepositories(account.Name)
	if err != nil {
		log.WithFields(log.Fields{
			"account:": account.Name,
		}).WithError(err).Error("Error during ProcessRepository")
		return err
	}

	for _, repository := range repositories {
		HandleRepository(fqdn, account, repository, keppelScanner, keppelProcessor)
	}

	return nil
}

func HandleRepository(fqdn string, account models.Account, repository models.Repository, keppelScanner *scanner.Scanner, keppelProcessor *processor.Processor) {
	component, err := keppelProcessor.ProcessRepository(fqdn, account, repository)
	if err != nil {
		log.WithFields(log.Fields{
			"account:":   account.Name,
			"repository": repository.Name,
		}).WithError(err).Error("Error during ProcessRepository")
		component, err = keppelProcessor.GetComponent(fmt.Sprintf("%s/%s/%s", fqdn, account.Name, repository.Name))
		if err != nil {
			log.WithFields(log.Fields{
				"account:":   account.Name,
				"repository": repository.Name,
			}).WithError(err).Error("Error during GetComponent")
		}
	}

	manifests, err := keppelScanner.ListManifests(account.Name, repository.Name)
	if err != nil {
		log.WithFields(log.Fields{
			"account:":   account.Name,
			"repository": repository.Name,
		}).WithError(err).Error("Error during ListManifests")
		return
	}
	for _, manifest := range manifests {
		if component == nil {
			log.WithFields(log.Fields{
				"account:":   account.Name,
				"repository": repository.Name,
			}).Error("Component not found")
			return
		}
		HandleManifest(account, repository, manifest, component, keppelScanner, keppelProcessor)
	}
}

func HandleManifest(account models.Account, repository models.Repository, manifest models.Manifest, component *client.Component, keppelScanner *scanner.Scanner, keppelProcessor *processor.Processor) {
	componentVersion, err := keppelProcessor.ProcessManifest(manifest, component.Id)
	if err != nil {
		log.WithFields(log.Fields{
			"account:":   account.Name,
			"repository": repository.Name,
		}).WithError(err).Error("Error during ProcessManifest")
		componentVersion, err = keppelProcessor.GetComponentVersion(manifest.Digest)
		if err != nil {
			log.WithFields(log.Fields{
				"account:":   account.Name,
				"repository": repository.Name,
			}).WithError(err).Error("Error during GetComponentVersion")
		}
	}
	trivyReport, err := keppelScanner.GetTrivyReport(account.Name, repository.Name, manifest.Digest)
	if err != nil {
		log.WithFields(log.Fields{
			"account:":   account.Name,
			"repository": repository.Name,
		}).WithError(err).Error("Error during GetTrivyReport")
		return
	}

	if trivyReport == nil {
		return
	}

	keppelProcessor.ProcessReport(*trivyReport, componentVersion.Id)
}
