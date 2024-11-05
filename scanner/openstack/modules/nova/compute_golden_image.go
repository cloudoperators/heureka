package nova

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/cloudoperators/heureka/scanner/openstack/processor"
	"github.com/cloudoperators/heureka/scanner/openstack/scanner"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

func GetCompliance() {
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

	service, err := osScanner.CreateComputeClient()
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

	// Create issue object
	issue45Id, err := processor.CreateIssueObject(*osProcessor, ctx, issuePrimaryName, issueDescription)
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

		componentId, err := processor.CreateComponentObject(*osProcessor, ctx, imageName)
		if err != nil {
			log.WithError(err).Fatal("Error during createComponentObject")
		}

		componentVersionId, err := processor.CreateComponentVersionObject(*osProcessor, ctx, imageVersion, componentId)
		if err != nil {
			log.WithError(err).Fatal("Error during createComponentVersionObject")
		}

		componentInstanceCCRN := serviceCCRN + "_" + fullImageName
		_, err = processor.CreateComponentInstanceObject(*osProcessor, ctx, componentInstanceCCRN, componentVersionId, serviceId, serviceCCRN)
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
