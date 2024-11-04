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

// createServiceObject creates a service object in the processor
// and returns the service ID
//
// Parameters:
//
//	osProcessor: Processor object
//	ctx: context object
//	serviceCCRN: CCRN of the service
//
// Returns:
//
//	string: Service ID
//	error: Error object
func createServiceObject(osProcessor processor.Processor, ctx context.Context, serviceCCRN string) (string, error) {
	serviceObj := processor.ServiceInfo{
		CCRN: serviceCCRN,
	}

	serviceId, err := osProcessor.ProcessService(ctx, serviceObj)
	if err != nil {
		log.WithError(err).Fatal("Error during processor processService")
	}

	return serviceId, err
}

// createSupportGroupObject creates a support group object in the processor
// and returns the support group ID
//
// Parameters:
//
//	osProcessor: Processor object
//	ctx: context object
//	supportGroupCCRN: CCRN of the support group
//
// Returns:
//
//	string: Support Group ID
//	error: Error object
func createSupportGroupObject(osProcessor processor.Processor, ctx context.Context, supportGroupCCRN string) (string, error) {
	supportGroupObj := processor.SupportGroupInfo{
		CCRN: supportGroupCCRN,
	}

	supportGroupId, err := osProcessor.ProcessSupportGroup(ctx, supportGroupObj)
	if err != nil {
		log.WithError(err).Fatal("Error during processor processSupportGroup")
	}

	return supportGroupId, err
}

// createIssueRepositoryObject creates an issue repository object in the processor
// and returns the issue repository ID
//
// Parameters:
//
//	osProcessor: Processor object
//	ctx: context object
//	issueRepositoryName: Name of the issue repository
//	issueRepositoryUrl: URL of the issue repository
//
// Returns:
//
//	string: Issue Repository ID
//	error: Error object
func createIssueRepositoryObject(osProcessor processor.Processor, ctx context.Context, issueRepositoryName string, issueRepositoryUrl string) (string, error) {
	issueRepositoryObj := processor.IssueRepositoryInfo{
		Name: issueRepositoryName,
		Url:  issueRepositoryUrl,
	}

	issueRepositoryId, err := osProcessor.ProcessIssueRepository(ctx, issueRepositoryObj)
	if err != nil {
		log.WithError(err).Fatal("Error during processor processIssueRepository")
	}

	return issueRepositoryId, err
}

// createComponentObject creates a component object in the processor
// and returns the component ID
//
// Parameters:
//
//	osProcessor: Processor object
//	ctx: context object
//	componentCCRN: CCRN of the component
//
// Returns:
//
//	string: Component ID
//	error: Error object
func createComponentObject(osProcessor processor.Processor, ctx context.Context, componentCCRN string) (string, error) {
	ComponentObj := processor.ComponentInfo{
		CCRN: componentCCRN,
	}

	componentId, err := osProcessor.ProcessComponent(ctx, ComponentObj)
	if err != nil {
		log.WithError(err).Fatal("Error during processor processComponent")
	}

	return componentId, err
}

// createComponentVersionObject creates a component version object in the processor
// and returns the component version ID
//
// Parameters:
//
//	osProcessor: Processor object
//	ctx: context object
//	version: Version of the component
//	componentID: ID of the component
//
// Returns:
//
//	string: Component Version ID
//	error: Error object
func createComponentVersionObject(osProcessor processor.Processor, ctx context.Context, version string, componentID string) (string, error) {
	componentVersionObj := processor.ComponentVersionInfo{
		Version:     version,
		ComponentID: componentID,
	}

	componentVersionId, err := osProcessor.ProcessComponentVersion(ctx, componentVersionObj)
	if err != nil {
		log.WithError(err).Fatal("Error during processor processComponentVersion")
	}

	return componentVersionId, err
}

// createComponentInstanceObject creates a component instance object in the processor
// and returns the component instance ID
//
// Parameters:
//
//	osProcessor: Processor object
//	ctx: context object
//	componentInstanceCCRN: CCRN of the component instance
//	componentVersionID: ID of the component version
//	serviceID: ID of the service
//	serviceCCRN: CCRN of the service
//
// Returns:
//
//	string: Component Instance ID
//	error: Error object
func createComponentInstanceObject(osProcessor processor.Processor, ctx context.Context, componentInstanceCCRN string, componentVersionID string, serviceID string, serviceCCRN string) (string, error) {
	componentInstanceObj := processor.ComponentInstanceInfo{
		CCRN:               componentInstanceCCRN,
		ComponentVersionID: componentVersionID,
		ServiceID:          serviceID,
		ServiceCCRN:        serviceCCRN,
	}

	componentInstanceID, err := osProcessor.ProcessComponentInstance(ctx, componentInstanceObj)
	if err != nil {
		log.WithError(err).Fatal("Error during processor processComponentInstance")
	}

	return componentInstanceID, err
}

// createIssueObject creates an issue object in the processor
// and returns the issue ID
//
// Parameters:
//
//	osProcessor: Processor object
//	ctx: context object
//	primaryName: Primary name of the issue
//	description: Description of the issue
//
// Returns:
//
//	string: Issue ID
//	error: Error object
func createIssueObject(osProcessor processor.Processor, ctx context.Context, primaryName string, description string) (string, error) {
	issueObj := processor.IssueInfo{
		PrimaryName: primaryName,
		Description: description,
	}

	issueId, err := osProcessor.ProcessIssue(ctx, issueObj)
	if err != nil {
		log.WithError(err).Fatal("Error during processor processIssue")
	}

	return issueId, err
}

func main() {

	// Some hardcoded values for now
	issueRepositoryName := "SAP Converged Cloud - Security Hardening"
	issueRepositoryUrl := "https://wiki.one.int.sap/wiki/display/itsec/SAP+Converged+Cloud+-+Security+Hardening"
	issuePrimaryName := "4.5 Ensure only approved Golden images are used in VM creation"
	issueDescription := "Only SAP approved Golden Images SHOULD be used. These Golden images are compliant to SAP security, legal, license and compliance requirements per default. The owner of the VM image is responsible to ensure it is compliant and secure."

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
	serviceId, err := createServiceObject(*osProcessor, ctx, serviceCCRN)
	if err != nil {
		log.WithError(err).Fatal("Error during createServiceObject")
	}

	// Create support group object
	supportGroupCCRN := serviceCCRN + "_SupportGroup"
	supportGroupId, err := createSupportGroupObject(*osProcessor, ctx, supportGroupCCRN)
	if err != nil {
		log.WithError(err).Fatal("Error during createSupportGroupObject")
	}

	// join service to support group
	err = osProcessor.ConnectServiceToSupportGroup(ctx, serviceId, supportGroupId)
	if err != nil {
		log.WithError(err).Warning("Failed adding service to support group")
	}

	// Create issue repository object
	issueRepositoryId, err := createIssueRepositoryObject(*osProcessor, ctx, issueRepositoryName, issueRepositoryUrl)
	if err != nil {
		log.WithError(err).Fatal("Error during createIssueRepositoryObject")
	}

	// join issue repository to service
	err = osProcessor.ConnectIssueRepositoryToService(ctx, issueRepositoryId, serviceId)
	if err != nil {
		log.WithError(err).Warning("Failed adding issue repository to service")
	}

	// Create issue object
	issue45Id, err := createIssueObject(*osProcessor, ctx, issuePrimaryName, issueDescription)
	if err != nil {
		log.WithError(err).Fatal("Error during createIssueObject")
	}

	// print servers in a formatted way
	for _, server := range servers {
		fmt.Printf("Server ID: %s, Server Name: %s\n", server.ID, server.Name)
		fmt.Printf("Server Status: %s\n", server.Status)
		fmt.Printf("Server Image Data: %v\n", server.Image)
		fmt.Print("\n\n")
	}

	// Create component object for each server
	for _, server := range servers {

		if server.Metadata == nil || server.Metadata["image_name"] == "" {
			// Skip servers without image name
			// Need to figure out how to handle this case in the future
			log.WithFields(log.Fields{
				"server_id": server.ID,
			}).Warning("Server image name is empty")
			continue
		}

		fullImageName := server.Metadata["image_name"]

		// Seperate Component name and version from server data
		re := regexp.MustCompile(`^([a-zA-Z\-]+)-([0-9].*)$`)
		matches := re.FindStringSubmatch(fullImageName)

		imageName := matches[1]
		imageVersion := matches[2]

		componentId, err := createComponentObject(*osProcessor, ctx, imageName)
		if err != nil {
			log.WithError(err).Fatal("Error during createComponentObject")
		}

		componentVersionId, err := createComponentVersionObject(*osProcessor, ctx, imageVersion, componentId)
		if err != nil {
			log.WithError(err).Fatal("Error during createComponentVersionObject")
		}

		componentInstanceCCRN := serviceCCRN + "_" + server.Metadata["image_name"]
		_, err = createComponentInstanceObject(*osProcessor, ctx, componentInstanceCCRN, componentVersionId, serviceId, serviceCCRN)
		if err != nil {
			log.WithError(err).Fatal("Error during createComponentInstanceObject")
		}

		// Perform policy checks
		if osProcessor.Policy4dot5Check(fullImageName) {
			// Compliant
			// Need to decide what to do here, if anything.
			// but for now we can just log it
			log.WithFields(log.Fields{
				"server_id":  server.ID,
				"image_name": fullImageName,
			}).Info("Server image is compliant")
		} else {
			// Non-Compliant
			// Connect component version to relevant issue
			err = osProcessor.ConnectComponentVersionToIssue(ctx, componentVersionId, issue45Id)
			if err != nil {
				log.WithError(err).Warning("Failed adding component version to issue")
			}
		}
	}
}
