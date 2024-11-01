// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
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

func createServiceObject(osProcessor processor.Processor, ctx context.Context, projectName string) (string, error) {
	serviceObj := processor.ServiceInfo{
		CCRN: projectName,
	}

	serviceId, err := osProcessor.ProcessService(ctx, serviceObj)
	if err != nil {
		log.WithError(err).Fatal("Error during processor process service")
	}

	return serviceId, err
}

func createSupportGroupObject(osProcessor processor.Processor, ctx context.Context, supportGroupName string) (string, error) {
	supportGroupObj := processor.SupportGroupInfo{
		CCRN: supportGroupName,
	}

	supportGroupId, err := osProcessor.ProcessSupportGroup(ctx, supportGroupObj)
	if err != nil {
		log.WithError(err).Fatal("Error during processor process support group")
	}

	return supportGroupId, err
}

func createIssueRepositoryObject(osProcessor processor.Processor, ctx context.Context, issueRepositoryName string, issueRepositoryUrl string) (string, error) {
	issueRepositoryObj := processor.IssueRepositoryInfo{
		Name: issueRepositoryName,
		Url:  issueRepositoryUrl,
	}

	issueRepositoryId, err := osProcessor.ProcessIssueRepository(ctx, issueRepositoryObj)
	if err != nil {
		log.WithError(err).Fatal("Error during processor process issue repository")
	}

	return issueRepositoryId, err
}

func createComponentObject(osProcessor processor.Processor, ctx context.Context, componentName string) (string, error) {
	ComponentObj := processor.ComponentInfo{
		CCRN: componentName,
	}

	componentId, err := osProcessor.ProcessComponent(ctx, ComponentObj)
	if err != nil {
		log.WithError(err).Fatal("Error during processor process component")
	}

	return componentId, err
}

func createComponentVersionObject(osProcessor processor.Processor, ctx context.Context, componentVersionID string) (string, error) {
	componentVersionObj := processor.ComponentVersionInfo{
		ComponentID: componentVersionID,
	}

	componentVersionId, err := osProcessor.ProcessComponentVersion(ctx, componentVersionObj)
	if err != nil {
		log.WithError(err).Fatal("Error during processor process service")
	}

	return componentVersionId, err
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

	// Create service object
	serviceName := scannerCfg.Project
	serviceId, err := createServiceObject(*osProcessor, ctx, serviceName)
	if err != nil {
		log.WithError(err).Fatal("Error during create service object")
	}

	// Create support group object
	supportGroupName := serviceName + "_SupportGroup"
	supportGroupId, err := createSupportGroupObject(*osProcessor, ctx, supportGroupName)
	if err != nil {
		log.WithError(err).Fatal("Error during create support group object")
	}

	// join service to support group
	err = osProcessor.ConnectServiceToSupportGroup(ctx, serviceId, supportGroupId)
	if err != nil {
		log.WithError(err).Warning("Failed adding service to support group")
	}

	// Create issue repository object
	// Hardcoded name & url for hardening guide for PoC
	issueRepositoryName := "SAP Converged Cloud - Security Hardening"
	issueRepositoryUrl := "https://wiki.one.int.sap/wiki/display/itsec/SAP+Converged+Cloud+-+Security+Hardening"
	_, err = createIssueRepositoryObject(*osProcessor, ctx, issueRepositoryName, issueRepositoryUrl)
	if err != nil {
		log.WithError(err).Fatal("Error during create issue repository object")
	}

	// Create component object for each server
	for _, server := range servers {
		// Seperate Component name and version from server data
		re := regexp.MustCompile(`^([a-zA-Z\-]+)-([0-9].*)$`)
		matches := re.FindStringSubmatch(server.Metadata["image_name"])
		//imageVersion := matches[2]
		imageName := matches[1]

		_, err = createComponentObject(*osProcessor, ctx, imageName)
		if err != nil {
			log.WithError(err).Fatal("Error during create component object")
		}
	}

	// Create component version object for each server
	for _, server := range servers {
		_, err = createComponentVersionObject(*osProcessor, ctx, server.Metadata["image_id"])
		if err != nil {
			log.WithError(err).Fatal("Error during create component version object")
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
