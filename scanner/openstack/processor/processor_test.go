// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor_test

import (
	"context"
	"strconv"

	"github.com/cloudoperators/heureka/scanner/openstack/processor"
	"github.com/kelseyhightower/envconfig"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/brianvoe/gofakeit/v7"
)

var _ = Describe("OpenStack Processor", func() {
	var (
		cfg processor.Config
		p   *processor.Processor
		ctx context.Context
	)

	BeforeEach(func() {
		err := envconfig.Process("openstack", &cfg)
		if err != nil {
			log.WithError(err).Fatal("Error while reading env config for processor")
		}
		p = processor.NewProcessor(cfg)
		ctx = context.Background()
	})

	Describe("CreateServiceObject", func() {
		It("should create a service object", func() {
			serviceCCRN := gofakeit.AppName() + gofakeit.UUID()

			newServiceId, err := processor.CreateServiceObject(*p, ctx, serviceCCRN)
			Expect(err).ToNot(HaveOccurred())
			Expect(newServiceId).ToNot(BeEmpty())
		})
	})

	Describe("ProcessService", func() {
		It("should process a service object", func() {
			serviceCCRN := gofakeit.AppName() + gofakeit.UUID()

			serviceObj := processor.ServiceInfo{
				CCRN: serviceCCRN,
			}

			serviceId, err := p.ProcessService(ctx, serviceObj)
			Expect(err).ToNot(HaveOccurred())
			Expect(serviceId).ToNot(BeEmpty())
		})
	})

	Describe("CreateSupportGroupObject", func() {
		It("should create a support group object", func() {
			supportGroupCCRN := gofakeit.AppName() + gofakeit.UUID()

			newSupportGroupId, err := processor.CreateSupportGroupObject(*p, ctx, supportGroupCCRN)
			Expect(err).ToNot(HaveOccurred())
			Expect(newSupportGroupId).ToNot(BeEmpty())
		})
	})

	Describe("ProcessSupportGroup", func() {
		It("should process a support group object", func() {
			supportGroupCCRN := gofakeit.AppName() + gofakeit.UUID()

			supportGroupObj := processor.SupportGroupInfo{
				CCRN: supportGroupCCRN,
			}

			supportGroupId, err := p.ProcessSupportGroup(ctx, supportGroupObj)
			Expect(err).ToNot(HaveOccurred())
			Expect(supportGroupId).ToNot(BeEmpty())
		})
	})

	Describe("ConnectServiceToSupportGroup", func() {
		It("should connect a service to a support group", func() {
			serviceCCRN := gofakeit.AppName() + gofakeit.UUID()
			supportGroupCCRN := gofakeit.AppName() + gofakeit.UUID()

			serviceId, err := processor.CreateServiceObject(*p, ctx, serviceCCRN)
			Expect(err).ToNot(HaveOccurred())
			Expect(serviceId).ToNot(BeEmpty())

			supportGroupId, err := processor.CreateSupportGroupObject(*p, ctx, supportGroupCCRN)
			Expect(err).ToNot(HaveOccurred())
			Expect(supportGroupId).ToNot(BeEmpty())

			err = p.ConnectServiceToSupportGroup(ctx, serviceId, supportGroupId)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("CreateIssueRepositoryObject", func() {
		It("should create an issue repository object", func() {
			issueRepositoryName := gofakeit.AppName() + gofakeit.UUID()
			issueRepositoryUrl := gofakeit.URL()

			newIssueRepositoryId, err := processor.CreateIssueRepositoryObject(*p, ctx, issueRepositoryName, issueRepositoryUrl)
			Expect(err).ToNot(HaveOccurred())
			Expect(newIssueRepositoryId).ToNot(BeEmpty())
		})
	})

	Describe("ProcessIssueRepository", func() {
		It("should process an issue repository object", func() {
			issueRepositoryName := gofakeit.AppName() + gofakeit.UUID()
			issueRepositoryUrl := gofakeit.URL()

			issueRepositoryObj := processor.IssueRepositoryInfo{
				Name: issueRepositoryName,
				Url:  issueRepositoryUrl,
			}

			issueRepositoryId, err := p.ProcessIssueRepository(ctx, issueRepositoryObj)
			Expect(err).ToNot(HaveOccurred())
			Expect(issueRepositoryId).ToNot(BeEmpty())
		})
	})

	// In heureka the issue repository is connected to the service automatically
	// 		upon creation so this function will always return an error
	// Describe("ConnectIssueRepositoryToService", func() {
	// 	It("should connect an issue repository to a service", func() {
	// 		serviceCCRN := gofakeit.AppName() + gofakeit.UUID()
	// 		issueRepositoryName := gofakeit.AppName() + gofakeit.UUID()
	// 		issueRepositoryUrl := gofakeit.URL()

	// 		serviceId, err := processor.CreateServiceObject(*p, ctx, serviceCCRN)
	// 		Expect(err).ToNot(HaveOccurred())
	// 		Expect(serviceId).ToNot(BeEmpty())

	// 		issueRepositoryId, err := processor.CreateIssueRepositoryObject(*p, ctx, issueRepositoryName, issueRepositoryUrl)
	// 		Expect(err).ToNot(HaveOccurred())
	// 		Expect(issueRepositoryId).ToNot(BeEmpty())

	// 		err = p.ConnectIssueRepositoryToService(ctx, issueRepositoryId, serviceId)
	// 		Expect(err).ToNot(HaveOccurred())
	// 	})
	// })

	Describe("CreateComponentObject", func() {
		It("should create a component object", func() {
			componentCCRN := gofakeit.AppName() + gofakeit.UUID()

			newComponentId, err := processor.CreateComponentObject(*p, ctx, componentCCRN)
			Expect(err).ToNot(HaveOccurred())
			Expect(newComponentId).ToNot(BeEmpty())
		})
	})

	Describe("ProcessComponent", func() {
		It("should process a component object", func() {
			componentCCRN := gofakeit.AppName() + gofakeit.UUID()

			componentObj := processor.ComponentInfo{
				CCRN: componentCCRN,
			}

			componentId, err := p.ProcessComponent(ctx, componentObj)
			Expect(err).ToNot(HaveOccurred())
			Expect(componentId).ToNot(BeEmpty())
		})
	})

	Describe("CreateComponentVersionObject", func() {
		It("should create a component version object", func() {
			componentVersion := gofakeit.AppName() + strconv.Itoa(gofakeit.Number(1, 100))

			componentCCRN := gofakeit.AppName() + gofakeit.UUID()

			newComponentId, err := processor.CreateComponentObject(*p, ctx, componentCCRN)
			Expect(err).ToNot(HaveOccurred())
			Expect(newComponentId).ToNot(BeEmpty())

			newComponentVersionId, err := processor.CreateComponentVersionObject(*p, ctx, componentVersion, newComponentId)
			Expect(err).ToNot(HaveOccurred())
			Expect(newComponentVersionId).ToNot(BeEmpty())
		})
	})

	Describe("ProcessComponentVersion", func() {
		It("should process a component version object", func() {
			componentCCRN := gofakeit.AppName() + gofakeit.UUID()

			componentVersion := gofakeit.AppName() + strconv.Itoa(gofakeit.Number(1, 100))

			newComponentId, err := processor.CreateComponentObject(*p, ctx, componentCCRN)
			Expect(err).ToNot(HaveOccurred())
			Expect(newComponentId).ToNot(BeEmpty())

			componentVersionObj := processor.ComponentVersionInfo{
				Version:     componentVersion,
				ComponentID: newComponentId,
			}

			componentVersionId, err := p.ProcessComponentVersion(ctx, componentVersionObj)
			Expect(err).ToNot(HaveOccurred())
			Expect(componentVersionId).ToNot(BeEmpty())
		})
	})

	Describe("CreateComponentInstanceObject", func() {
		It("should create a component instance object", func() {
			componentInstanceCCRN := gofakeit.AppName() + gofakeit.UUID()

			componentVersion := gofakeit.AppName() + strconv.Itoa(gofakeit.Number(1, 100))

			componentCCRN := gofakeit.AppName() + gofakeit.UUID()

			newComponentId, err := processor.CreateComponentObject(*p, ctx, componentCCRN)
			Expect(err).ToNot(HaveOccurred())
			Expect(newComponentId).ToNot(BeEmpty())

			newComponentVersionId, err := processor.CreateComponentVersionObject(*p, ctx, componentVersion, newComponentId)
			Expect(err).ToNot(HaveOccurred())
			Expect(newComponentVersionId).ToNot(BeEmpty())

			serviceCCRN := gofakeit.AppName() + gofakeit.UUID()

			newServiceId, err := processor.CreateServiceObject(*p, ctx, serviceCCRN)
			Expect(err).ToNot(HaveOccurred())
			Expect(newServiceId).ToNot(BeEmpty())

			newComponentInstanceId, err := processor.CreateComponentInstanceObject(*p, ctx, componentInstanceCCRN, newComponentVersionId, newServiceId, serviceCCRN)
			Expect(err).ToNot(HaveOccurred())
			Expect(newComponentInstanceId).ToNot(BeEmpty())
		})
	})

	Describe("ProcessComponentInstance", func() {
		It("should process a component instance object", func() {
			componentInstanceCCRN := "test-component-instance"

			componentVersion := "test-component-version"

			componentCCRN := "test-component"

			newComponentId, err := processor.CreateComponentObject(*p, ctx, componentCCRN)
			Expect(err).ToNot(HaveOccurred())
			Expect(newComponentId).ToNot(BeEmpty())

			newComponentVersionId, err := processor.CreateComponentVersionObject(*p, ctx, componentVersion, newComponentId)
			Expect(err).ToNot(HaveOccurred())
			Expect(newComponentVersionId).ToNot(BeEmpty())

			serviceCCRN := gofakeit.AppName() + gofakeit.UUID()

			newServiceId, err := processor.CreateServiceObject(*p, ctx, serviceCCRN)
			Expect(err).ToNot(HaveOccurred())
			Expect(newServiceId).ToNot(BeEmpty())

			componentInstanceObj := processor.ComponentInstanceInfo{
				CCRN:               componentInstanceCCRN,
				ComponentVersionID: newComponentVersionId,
				ServiceID:          newServiceId,
				ServiceCCRN:        serviceCCRN,
			}

			componentInstanceId, err := p.ProcessComponentInstance(ctx, componentInstanceObj)
			Expect(err).ToNot(HaveOccurred())
			Expect(componentInstanceId).ToNot(BeEmpty())
		})
	})

	Describe("CreateIssueObject", func() {
		It("should create an issue object", func() {
			issuePrimaryName := gofakeit.AppName() + gofakeit.UUID()
			issueDesc := gofakeit.Product().Description

			newIssueId, err := processor.CreateIssueObject(*p, ctx, issuePrimaryName, issueDesc)
			Expect(err).ToNot(HaveOccurred())
			Expect(newIssueId).ToNot(BeEmpty())
		})
	})

	Describe("ProcessIssue", func() {
		It("should process an issue object", func() {
			issuePrimaryName := gofakeit.AppName() + gofakeit.UUID()
			issueDesc := gofakeit.Product().Description

			issueObj := processor.IssueInfo{
				PrimaryName: issuePrimaryName,
				Description: issueDesc,
			}

			issueId, err := p.ProcessIssue(ctx, issueObj)
			Expect(err).ToNot(HaveOccurred())
			Expect(issueId).ToNot(BeEmpty())
		})
	})

	Describe("ConnectComponentVersionToIssue", func() {
		It("should connect a component version to an issue", func() {
			componentVersion := gofakeit.AppName() + strconv.Itoa(gofakeit.Number(1, 100))

			componentCCRN := gofakeit.AppName() + gofakeit.UUID()

			newComponentId, err := processor.CreateComponentObject(*p, ctx, componentCCRN)
			Expect(err).ToNot(HaveOccurred())
			Expect(newComponentId).ToNot(BeEmpty())

			newComponentVersionId, err := processor.CreateComponentVersionObject(*p, ctx, componentVersion, newComponentId)
			Expect(err).ToNot(HaveOccurred())
			Expect(newComponentVersionId).ToNot(BeEmpty())

			issuePrimaryName := gofakeit.AppName() + gofakeit.UUID()
			issueDesc := gofakeit.Product().Description

			newIssueId, err := processor.CreateIssueObject(*p, ctx, issuePrimaryName, issueDesc)
			Expect(err).ToNot(HaveOccurred())
			Expect(newIssueId).ToNot(BeEmpty())

			err = p.ConnectComponentVersionToIssue(ctx, newComponentVersionId, newIssueId)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("CreateIssueVariantObject", func() {
		It("should create an issue variant object", func() {
			secondaryName := gofakeit.AppName() + gofakeit.UUID()
			description := gofakeit.Product().Description

			issueRepositoryName := gofakeit.AppName() + gofakeit.UUID()
			issueRepositoryUrl := gofakeit.URL()

			newIssueRepositoryId, err := processor.CreateIssueRepositoryObject(*p, ctx, issueRepositoryName, issueRepositoryUrl)
			Expect(err).ToNot(HaveOccurred())
			Expect(newIssueRepositoryId).ToNot(BeEmpty())

			issuePrimaryName := gofakeit.AppName() + gofakeit.UUID()
			issueDesc := gofakeit.Product().Description

			newIssueId, err := processor.CreateIssueObject(*p, ctx, issuePrimaryName, issueDesc)
			Expect(err).ToNot(HaveOccurred())
			Expect(newIssueId).ToNot(BeEmpty())

			newIssueVariantId, err := processor.CreateIssueVariantObject(*p, ctx, secondaryName, description, newIssueRepositoryId, newIssueId)
			Expect(err).ToNot(HaveOccurred())
			Expect(newIssueVariantId).ToNot(BeEmpty())
		})
	})

	Describe("ProcessIssueVariant", func() {
		It("should process an issue variant object", func() {
			secondaryName := gofakeit.AppName() + gofakeit.UUID()
			description := gofakeit.Product().Description

			issueRepositoryName := gofakeit.AppName() + gofakeit.UUID()
			issueRepositoryUrl := gofakeit.URL()

			newIssueRepositoryId, err := processor.CreateIssueRepositoryObject(*p, ctx, issueRepositoryName, issueRepositoryUrl)
			Expect(err).ToNot(HaveOccurred())
			Expect(newIssueRepositoryId).ToNot(BeEmpty())

			issuePrimaryName := gofakeit.AppName() + gofakeit.UUID()
			issueDesc := gofakeit.Product().Description

			newIssueId, err := processor.CreateIssueObject(*p, ctx, issuePrimaryName, issueDesc)
			Expect(err).ToNot(HaveOccurred())
			Expect(newIssueId).ToNot(BeEmpty())

			issueVariantObj := processor.IssueVariantInfo{
				SecondaryName:     secondaryName,
				Description:       description,
				IssueRepositoryID: newIssueRepositoryId,
				IssueID:           newIssueId,
			}

			issueVariantId, err := p.ProcessIssueVariant(ctx, issueVariantObj)
			Expect(err).ToNot(HaveOccurred())
			Expect(issueVariantId).ToNot(BeEmpty())
		})
	})

})
