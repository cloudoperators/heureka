// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package keystone

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/scanner/openstack/processor"
	"github.com/cloudoperators/heureka/scanner/openstack/scanner"
	log "github.com/sirupsen/logrus"
)

var adminCount int

// policy2dot2Check checks if the given user complies with policy 2.2.
// Policy 2.2 requires that project users do not contain the role "admin".
//
// Parameters:
// - userName (string): the name of the user
// - roles ([]string): slice of roles for a given user
//
// Returns:
// - bool: Returns true if the user complies with policy 2.2, otherwise false.
func policy2dot2Check(userName string, roles []string) bool {
	isAdmin := false
	isTechnicalUser := strings.HasPrefix(userName, "TM3")
	compliant := true

	// Iterate over the roles to check for "registry_admin" and invalid roles for technical users
	for _, role := range roles {
		if role == "registry_admin" {
			isAdmin = true
		}
		if isTechnicalUser && strings.Contains(role, "admin") {
			compliant = false
			log.WithFields(log.Fields{
				"user_name": userName,
				"reason":    "Technical user cannot have roles containing 'admin'",
			}).Info("User is non-compliant")
			return false
		}
	}

	// Check if the user is both an admin and a technical user
	if isAdmin && isTechnicalUser {
		compliant = false
		log.WithFields(log.Fields{
			"user_name": userName,
			"reason":    "Technical user cannot have 'registry_admin' role",
		}).Info("User is non-compliant")
		return false
	}

	// Check if the user is a technical user with admin roles
	if isTechnicalUser && !isAdmin && strings.Contains(strings.Join(roles, ","), "admin") {
		compliant = false
		log.WithFields(log.Fields{
			"user_name": userName,
			"reason":    "Technical user with admin roles",
		}).Info("User is non-compliant")
		return false
	}

	// Check if the user is an admin with technical roles
	if !isTechnicalUser && isAdmin && !strings.Contains(strings.Join(roles, ","), "admin") {
		compliant = false
		log.WithFields(log.Fields{
			"user_name": userName,
			"reason":    "Admin user with technical roles",
		}).Info("User is non-compliant")
		return false
	}

	return compliant
}

// ComputeUserRoleCompliance checks the compliance of the user roles in the OpenStack project.
// Policy 2.2 requires that project users do not contain the role "admin".
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
func ComputeUserRoleCompliance(osScanner *scanner.Scanner, osProcessor *processor.Processor, ctx context.Context, serviceId string, serviceCCRN string, issueRepoId string) {
	// Some hardcoded values for now
	issuePrimaryName := "2.2 Ensure secure permission & role concept"
	issueDescription := "It MUST be ensured that a secure group & permission concept, following the Need-to-know Principle, is in place for every Plus One Converged Cloud project."

	identityService, err := osScanner.CreateIdentityClient()
	if err != nil {
		log.WithError(err).Fatal("Error during scanner setup")
	}

	users := osScanner.GetUsers(identityService, osScanner.ProjectId)
	if err != nil {
		log.WithError(err).Fatal("Error during scanner get servers")
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

	// Create component object for each user
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
			if policy2dot2Check(userName, userRoles) {
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
		// Check if the number of admin users is less than 2
		if adminCount < 2 {
			for _, user := range users {
				userName := user["user"].(string)
				userRoles := user["roles"].([]string)
				joined := strings.Join(userRoles, ",")
				hashedConf := osScanner.Md5Hash(joined)
				componentId, err := processor.CreateComponentObject(*osProcessor, ctx, userName)
				if err != nil {
					log.WithError(err).Fatal("Error during createComponentObject")
				}
				componentVersionId, err := processor.CreateComponentVersionObject(*osProcessor, ctx, hashedConf, componentId)
				if err != nil {
					log.WithError(err).Fatal("Error during createComponentVersionObject")
				}
				err = osProcessor.ConnectComponentVersionToIssue(ctx, componentVersionId, issue22Id)
				if err != nil {
					log.WithError(err).Warning("Failed adding component version to issue")
				}
				log.WithFields(log.Fields{
					"user_name": userName,
					"reason":    "Less than 2 admin users",
				}).Info("User is non-compliant")
			}
		}
	}
}
