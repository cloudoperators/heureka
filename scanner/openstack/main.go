// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os"
	"time"

	"github.com/cloudoperators/heureka/scanner/openstack/modules/nova"
	"github.com/cloudoperators/heureka/scanner/openstack/processor"
	"github.com/cloudoperators/heureka/scanner/openstack/scanner"
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
	err := envconfig.Process("OS", &cfg)
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
	// Some hardcoded values for now
	issueRepositoryName := "SAP Converged Cloud - Security Hardening"
	issueRepositoryUrl := "https://wiki.one.int.sap/wiki/display/itsec/SAP+Converged+Cloud+-+Security+Hardening"

	var scannerCfg scanner.Config
	err := envconfig.Process("openstack", &scannerCfg)
	if err != nil {
		log.WithError(err).Fatal("Error while reading env config for scanner")
	}

	var processorsCfg processor.Config
	err = envconfig.Process("openstack", &processorsCfg)
	if err != nil {
		log.WithError(err).Fatal("Error while reading env config for processor")
	}

	osScanner := scanner.NewScanner(scannerCfg)
	osProcessor := processor.NewProcessor(processorsCfg)

	// Create context with timeout (30min should be ok)
	scanTimeout, err := time.ParseDuration(scannerCfg.ScannerTimeout)
	if err != nil {
		log.WithError(err).Fatal("couldn't parse scanner timeout, setting it to 30 minutes")
		scanTimeout = 30 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	defer cancel()

	// Create service object
	serviceCCRN := scannerCfg.Project
	serviceId, err := processor.CreateServiceObject(*osProcessor, ctx, serviceCCRN)
	if err != nil {
		log.WithError(err).Fatal("Error during createServiceObject")
	}

	// Create support group object
	supportGroupCCRN := serviceCCRN + "_SupportGroup"
	supportGroupId, err := processor.CreateSupportGroupObject(*osProcessor, ctx, supportGroupCCRN)
	if err != nil {
		log.WithError(err).Fatal("Error during createSupportGroupObject")
	}

	// join service to support group
	err = osProcessor.ConnectServiceToSupportGroup(ctx, serviceId, supportGroupId)
	if err != nil {
		log.WithError(err).Warning("Failed adding service to support group")
	}

	// Create issue repository object
	issueRepositoryId, err := processor.CreateIssueRepositoryObject(*osProcessor, ctx, issueRepositoryName, issueRepositoryUrl)
	if err != nil {
		log.WithError(err).Fatal("Error during createIssueRepositoryObject")
	}

	// join issue repository to service
	err = osProcessor.ConnectIssueRepositoryToService(ctx, issueRepositoryId, serviceId)
	if err != nil {
		log.WithError(err).Warning("Failed adding issue repository to service")
	}

	nova.ComputeGoldenImageCompliance(osScanner, osProcessor, ctx, serviceId, serviceCCRN, issueRepositoryId)
	// keystone.ComputeUserRoleCompliance(osScanner, osProcessor, ctx, serviceId, serviceCCRN, issueRepositoryId)
}
