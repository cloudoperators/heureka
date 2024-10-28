// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

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

	service, err := osScanner.Setup()
	if err != nil {
		log.WithError(err).Fatal("Error during scanner setup")
	}

	servers, err := osScanner.GetServers(service)
	if err != nil {
		log.WithError(err).Fatal("Error during scanner get servers")
	}

	fmt.Print("Servers: \n")
	fmt.Print(servers)
	fmt.Print("\n\n")

	results, err := osProcessor.ProcessServers(servers)
	if err != nil {
		log.WithError(err).Fatal("Error during processor process servers")
	}

	fmt.Print("Results: \n")
	fmt.Print(results)
	fmt.Print("\n\n")
}
