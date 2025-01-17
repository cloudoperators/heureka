// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package nova_test

import (
	"context"
	"os"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/scanner/openstack/modules/nova"
	"github.com/cloudoperators/heureka/scanner/openstack/processor"
	"github.com/cloudoperators/heureka/scanner/openstack/scanner"
	"github.com/kelseyhightower/envconfig"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("Processor", func() {
	var (
		scannerCfg        scanner.Config
		processorsCfg     processor.Config
		s                 *scanner.Scanner
		p                 *processor.Processor
		ctx               context.Context
		sapCompliantOwner string
		gardenLinuxOwner  string
	)

	BeforeEach(func() {
		err := envconfig.Process("openstack", &scannerCfg)
		if err != nil {
			log.WithError(err).Fatal("Error while reading env config for scanner")
		}

		err = envconfig.Process("openstack", &processorsCfg)
		if err != nil {
			log.WithError(err).Fatal("Error while reading env config for processor")
		}

		s = scanner.NewScanner(scannerCfg)
		p = processor.NewProcessor(processorsCfg)
		ctx = context.Background()

		sapCompliantOwner = os.Getenv("SAP_COMPLIANT_OWNER_ID")
		gardenLinuxOwner = os.Getenv("GARDENLINUX_OWNER_ID")
	})

	Describe("Policy4dot5Check", func() {
		Context("when the image name contains gardenlinux and is compliant", func() {
			It("should return true", func() {
				compliantString := "gardenlinux"

				imageName := gofakeit.AppName() + compliantString + gofakeit.UUID()
				imageOwner := gardenLinuxOwner

				result := nova.Policy4dot5Check(imageName, imageOwner)
				Expect(result).To(BeTrue())
			})
		})

		Context("when the image name contains SAP-compliant and is compliant", func() {
			It("should return true", func() {
				compliantString := "SAP-compliant"

				imageName := gofakeit.AppName() + compliantString + gofakeit.UUID()
				imageOwner := sapCompliantOwner

				result := nova.Policy4dot5Check(imageName, imageOwner)
				Expect(result).To(BeTrue())
			})
		})

		Context("when the image name is compliant but the image owner is not", func() {
			It("should return false", func() {
				compliantString := "SAP-compliant"

				imageName := gofakeit.AppName() + compliantString + gofakeit.UUID()
				imageOwner := gofakeit.UUID()

				result := nova.Policy4dot5Check(imageName, imageOwner)
				Expect(result).To(BeFalse())
			})
		})

		Context("when the image name is not compliant but the image owner is", func() {
			It("should return false", func() {
				imageName := gofakeit.AppName() + gofakeit.UUID()
				imageOwner := sapCompliantOwner

				result := nova.Policy4dot5Check(imageName, imageOwner)
				Expect(result).To(BeFalse())
			})
		})

		Context("when the image name and image owner are not compliant", func() {
			It("should return false", func() {
				imageName := gofakeit.AppName() + gofakeit.UUID()
				imageOwner := gofakeit.UUID()

				result := nova.Policy4dot5Check(imageName, imageOwner)
				Expect(result).To(BeFalse())
			})
		})
	})

	Describe("ComputeGoldenImageCompliance", func() {
		Context("when the golden image is not compliant", func() {
			It("should create an issue", func() {
				// TODO: Decide how to mock data coming from openstack

				// Create service object
				serviceCCRN := gofakeit.AppName() + gofakeit.UUID()
				serviceId, err := processor.CreateServiceObject(*p, ctx, serviceCCRN)
				Expect(err).To(BeNil())
				Expect(serviceId).ToNot(BeEmpty())

				// Create support group object
				supportGroupCCRN := gofakeit.AppName() + gofakeit.UUID()
				supportGroupId, err := processor.CreateSupportGroupObject(*p, ctx, supportGroupCCRN)
				Expect(err).To(BeNil())
				Expect(supportGroupId).ToNot(BeEmpty())

				// join service to support group
				err = p.ConnectServiceToSupportGroup(ctx, serviceId, supportGroupId)
				Expect(err).To(BeNil())

				issueRepositoryName := gofakeit.AppName() + gofakeit.UUID()
				issueRepositoryUrl := gofakeit.URL()

				// Create issue repository object
				issueRepositoryId, err := processor.CreateIssueRepositoryObject(*p, ctx, issueRepositoryName, issueRepositoryUrl)
				Expect(err).To(BeNil())
				Expect(issueRepositoryId).ToNot(BeEmpty())

				nova.ComputeGoldenImageCompliance(s, p, ctx, serviceId, serviceCCRN, issueRepositoryId)
			})

		})

	})

})
