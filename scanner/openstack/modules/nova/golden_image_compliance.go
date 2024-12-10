package nova

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloudoperators/heureka/scanner/openstack/processor"
	"github.com/cloudoperators/heureka/scanner/openstack/scanner"
	log "github.com/sirupsen/logrus"
)

// policy4dot5Check checks if the given image name complies with policy 4.5.
// Policy 4.5 requires that the image name contains either "gardenlinux" or "SAP-compliant".
//
// Parameters:
//
//	img_name (string): The name of the image to be checked.
//
// Returns:
//
//	bool: Returns true if the image name complies with policy 4.5, otherwise false.
func policy4dot5Check(imgName string, imgOwner string) bool {
	// This is a temporary hardcoded implementation of policy 4.5 for the OpenStack scanner PoC
	// This function will be replaced by the actual implementation of policy checks in the future
	// Policy 4.5 checks that the image name contains either "gardenlinux" or "SAP-compliant"

	// Temporary implementation for owner check, storing the owner ids as env variables
	sapCompliantOwner := os.Getenv("SAP_COMPLIANT_OWNER_ID")
	gardenLinuxOwner := os.Getenv("GARDENLINUX_OWNER_ID")

	if (strings.Contains(imgName, "gardenlinux") && imgOwner == gardenLinuxOwner) ||
		(strings.Contains(imgName, "SAP-compliant") && imgOwner == sapCompliantOwner) {
		return true
	}
	return false
}

// ComputeGoldenImageCompliance checks the compliance of the golden images used in VM creation.
// Policy 4.5 requires that only approved Golden images are used in VM creation.
//
// Parameters:
//
//	osScanner (*scanner.Scanner): The scanner object for the OpenStack cloud.
//	osProcessor (*processor.Processor): The processor object for the OpenStack cloud.
//	ctx (context.Context): The context object for the OpenStack cloud.
//	serviceId (string): The ID of the service.
//	serviceCCRN (string): The CCRN of the service.
//	issueRepoId (string): The ID of the issue repository.
//
// Returns:
//
//	None
func ComputeGoldenImageCompliance(osScanner *scanner.Scanner, osProcessor *processor.Processor, ctx context.Context, serviceId string, serviceCCRN string, issueRepoId string) {
	// Some hardcoded values for now
	issuePrimaryName := "4.5 Ensure only approved Golden images are used in VM creation"
	issueDescription := "Only SAP approved Golden Images SHOULD be used. These Golden images are compliant to SAP security, legal, license and compliance requirements per default. The owner of the VM image is responsible to ensure it is compliant and secure."

	// Hardocding the issue variant to be a copy of the issue for now
	issueVariantSecondaryName := issuePrimaryName
	issueVariantDescription := issueDescription

	computeService, err := osScanner.CreateComputeClient()
	if err != nil {
		log.WithError(err).Fatal("Error during scanner setup")
	}

	imageService, err := osScanner.CreateImageClient()
	if err != nil {
		log.WithError(err).Fatal("Error during scanner setup")
	}

	servers, err := osScanner.GetServers(computeService)
	if err != nil {
		log.WithError(err).Fatal("Error during scanner get servers")
	}

	// Create issue object
	issue45Id, err := processor.CreateIssueObject(*osProcessor, ctx, issuePrimaryName, issueDescription)
	if err != nil {
		log.WithError(err).Fatal("Error during createIssueObject")
	}

	// Create Issue Variant object
	_, err = processor.CreateIssueVariantObject(*osProcessor, ctx, issueVariantSecondaryName, issueVariantDescription, issueRepoId, issue45Id)
	if err != nil {
		log.WithError(err).Fatal("Error during createIssueVariantObject")
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

		if _, exists := server.Image["id"]; !exists {
			// Skip servers without image name
			// Need to figure out how to handle this case in the future
			log.WithFields(log.Fields{
				"server_id": server.ID,
			}).Warning("Server image name is empty")
			continue
		}

		imageId := server.Image["id"].(string)
		image, err := osScanner.GetImage(imageService, imageId)
		if err != nil {
			log.WithError(err).Warn("Error during scanner get image")
		}

		imageName := image.Name
		imageChecksum := image.Checksum
		imageOwner := image.Owner
		serverName := server.Name

		componentId, err := processor.CreateComponentObject(*osProcessor, ctx, imageName)
		if err != nil {
			log.WithError(err).Fatal("Error during createComponentObject")
		}

		componentVersionId, err := processor.CreateComponentVersionObject(*osProcessor, ctx, imageChecksum, componentId)
		if err != nil {
			log.WithError(err).Fatal("Error during createComponentVersionObject")
		}

		componentInstanceCCRN := serviceCCRN + "_" + imageName + "_" + serverName + "_" + osScanner.Domain + "_" + osScanner.Region
		_, err = processor.CreateComponentInstanceObject(*osProcessor, ctx, componentInstanceCCRN, componentVersionId, serviceId, serviceCCRN)
		if err != nil {
			log.WithError(err).Fatal("Error during createComponentInstanceObject")
		}

		// Perform policy checks
		if policy4dot5Check(imageName, imageOwner) {
			// Compliant
			// Need to decide what to do here, if anything.
			// but for now we can just log it
			log.WithFields(log.Fields{
				"server_id":  server.ID,
				"image_name": imageName,
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
