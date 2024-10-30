// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"time"

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

	// print servers in a formatted way
	for _, server := range servers {
		fmt.Printf("Server ID: %s, Server Name: %s\n", server.ID, server.Name)
		fmt.Printf("Server Status: %s\n", server.Status)
		fmt.Print("\n\n")
	}

	// Create context with timeout (30min should be ok)
	scanTimeout, err := time.ParseDuration(scannerCfg.ScannerTimeout)
	if err != nil {
		log.WithError(err).Fatal("couldn't parse scanner timeout, setting it to 30 minutes")
		scanTimeout = 30 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	defer cancel()

	// loop through servers and process them
	for _, server := range servers {

		serviceObj := processor.ServiceInfo{
			CCRN:         server.Name,
			SupportGroup: "none",
		}

		fmt.Print("Processing server: \n")
		fmt.Print(serviceObj.CCRN)

		_, err := osProcessor.ProcessService(ctx, serviceObj)
		if err != nil {
			log.WithError(err).Fatal("Error during processor process service")
		}
	}

	results, err := osProcessor.ProcessServers(servers)
	if err != nil {
		log.WithError(err).Fatal("Error during processor process servers")
	}

	fmt.Print("Results: \n")
	fmt.Print(results)
	fmt.Print("\n\n")
}
