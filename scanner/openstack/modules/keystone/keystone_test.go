// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package keystone_test

import (
	"context"
	"os"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/scanner/openstack/modules/keystone"
	"github.com/cloudoperators/heureka/scanner/openstack/processor"
	"github.com/cloudoperators/heureka/scanner/openstack/scanner"
	"github.com/kelseyhightower/envconfig"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	log "github.com/sirupsen/logrus"
)

var _ = Describe("Keystone", func() {
	var (
		server        *ghttp.Server
		scannerCfg    scanner.Config
		processorsCfg processor.Config
		s             *scanner.Scanner
		p             *processor.Processor
		ctx           context.Context
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		os.Setenv("OS_AUTH_URL", server.Addr())

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
	})

	Describe("Policy2dot2Check", func() {
		Context("user is not admin nor technical user", func() {
			It("should return true", func() {
				username := gofakeit.Name()
				roles := []string{"role1", "role2"}

				result := keystone.Policy2dot2Check(username, roles)
				Expect(result).To(BeTrue())
			})
		})

		Context("user is admin but not technical user", func() {
			It("should return true", func() {
				username := gofakeit.Name()
				roles := []string{"registry_admin", "role2"}

				result := keystone.Policy2dot2Check(username, roles)
				Expect(result).To(BeTrue())
			})
		})

		Context("user is technical user but not admin", func() {
			It("should return true", func() {
				username := "TM3" + gofakeit.Name()
				roles := []string{"role1", "role2"}

				result := keystone.Policy2dot2Check(username, roles)
				Expect(result).To(BeTrue())
			})
		})

		Context("user is admin and technical user", func() {
			It("should return false", func() {
				username := "TM3" + gofakeit.Name()
				roles := []string{"registry_admin", "role2"}

				result := keystone.Policy2dot2Check(username, roles)
				Expect(result).To(BeFalse())
			})
		})
	})

	Describe("ComputeUserRoleCompliance", func() {
		Context("when the user role is not compliant", func() {
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

				keystone.ComputeUserRoleCompliance(s, p, ctx, serviceId, serviceCCRN, issueRepositoryId)

			})

		})

	})

})
