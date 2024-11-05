package keystone

import (
	"context"
	"fmt"
	"strings"
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
	issuePrimaryName := "2.2 Ensure secure permission & role concept"
	issueDescription := "It MUST be ensured that a secure group & permission concept, following the Need-to-know Principle, is in place for every Plus One Converged Cloud project."

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

	service, err := osScanner.CreateIdentityClient()
	if err != nil {
		log.WithError(err).Fatal("Error during scanner setup")
	}

	users := osScanner.GetUsers(service, osScanner.ProjectId)
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
	issue22Id, err := processor.CreateIssueObject(*osProcessor, ctx, issuePrimaryName, issueDescription)
	if err != nil {
		log.WithError(err).Fatal("Error during createIssueObject")
	}

	//print servers in a formatted way
	for _, user := range users {
		fmt.Printf("User Name: %s\n", user["user"].(string))
		fmt.Printf("User Roles %s\n", user["roles"].([]string))
		fmt.Print("\n\n")
	}

	// Create component object for each server
	if len(users) != 0 {
		for _, user := range users {
			userName := user["user"].(string)
			userRoles := user["roles"].([]string)

			if userName == "" {
				// Skip servers without image name
				// Need to figure out how to handle this case in the future
				log.Warning("user name is empty")
				continue
			}

			// Join the slice into a single string
			joined := strings.Join(userRoles, ",") // Use a delimiter if needed
			hashedConf := osScanner.Md5Hash(joined)

			componentId, err := processor.CreateComponentObject(*osProcessor, ctx, userName)
			if err != nil {
				log.WithError(err).Fatal("Error during createComponentObject")
			}

			componentVersionId, err := processor.CreateComponentVersionObject(*osProcessor, ctx, hashedConf, componentId)
			if err != nil {
				log.WithError(err).Fatal("Error during createComponentVersionObject")
			}

			componentInstanceCCRN := serviceCCRN + "_" + userName
			_, err = processor.CreateComponentInstanceObject(*osProcessor, ctx, componentInstanceCCRN, componentVersionId, serviceId, serviceCCRN)
			if err != nil {
				log.WithError(err).Fatal("Error during createComponentInstanceObject")
			}

			// Perform policy checks
			if osProcessor.Policy2dot2Check(userRoles) {
				// Compliant
				// Need to decide what to do here, if anything.
				// but for now we can just log it
				log.WithFields(log.Fields{
					"user_name":  userName,
					"user_roles": joined,
				}).Info("User is compliant")
			} else {
				// Non-Compliant
				// Connect component version to relevant issue
				err = osProcessor.ConnectComponentVersionToIssue(ctx, componentVersionId, issue22Id)
				if err != nil {
					log.WithError(err).Warning("Failed adding component version to issue")
				}
			}
		}
	}
}
