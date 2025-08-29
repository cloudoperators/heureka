// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package openfga_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/cloudoperators/heureka/internal/util"
)

const (
	testModelFilePath = "./testdata/testModel.fga"
	openfgaApiUrl     = "http://localhost:8080"

	documentType = "document"
	readerRel    = "reader"
	writerRel    = "writer"
	ownerRel     = "owner"
)

var _ = Describe("Authz", func() {
	var (
		cfg   *util.Config
		authz openfga.Authorization
	)

	BeforeEach(func() {
		enableLogs := true
		cfg = &util.Config{
			AuthzEnabled:      true,
			AuthModelFilePath: testModelFilePath,
			OpenFGApiUrl:      openfgaApiUrl,
		}
		authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
	})

	Describe("NewAuthz", func() {
		It("should create a new Authz instance", func() {
			Expect(authz).NotTo(BeNil())
		})
	})

	Describe("CheckPermission", func() {
		It("should return false with no relations added", func() {
			ok, err := authz.CheckPermission("user1", "document1", documentType, readerRel)
			Expect(ok).To(BeFalse())
			Expect(err).To(BeNil())
		})

		It("should return an error for invalid resource type", func() {
			ok, err := authz.CheckPermission("user1", "document1", "invalid_type", readerRel)
			Expect(ok).To(BeFalse())
			Expect(err).NotTo(BeNil())
		})

		It("should return true after adding relation", func() {
			err := authz.AddRelation("user1", "document1", documentType, readerRel)
			Expect(err).To(BeNil())

			ok, err := authz.CheckPermission("user1", "document1", documentType, readerRel)
			Expect(ok).To(BeTrue())
			Expect(err).To(BeNil())
		})

		It("should return false after adding wrong relation", func() {
			err := authz.AddRelation("user1", "document1", documentType, readerRel)
			Expect(err).To(BeNil())

			ok, err := authz.CheckPermission("user1", "document1", documentType, ownerRel)
			Expect(ok).To(BeFalse())
			Expect(err).To(BeNil())
		})
	})

	Describe("AddRelation", func() {
		It("should return no error", func() {
			err := authz.AddRelation("user1", "document1", documentType, writerRel)
			Expect(err).To(BeNil())
		})

		It("should return error when adding invalid relation", func() {
			err := authz.AddRelation("user1", "document1", documentType, "invalid_relation")
			Expect(err).NotTo(BeNil())
		})

		It("should return error when adding relation to invalid resource type", func() {
			err := authz.AddRelation("user1", "document1", "invalid_type", writerRel)
			Expect(err).NotTo(BeNil())
		})

		It("should return error when adding relation with empty userId", func() {
			err := authz.AddRelation("", "document1", documentType, writerRel)
			Expect(err).NotTo(BeNil())
		})

		It("should be able to access resource after adding relation", func() {
			err := authz.AddRelation("user1", "document1", documentType, ownerRel)
			Expect(err).To(BeNil())

			ok, err := authz.CheckPermission("user1", "document1", documentType, ownerRel)
			Expect(ok).To(BeTrue())
			Expect(err).To(BeNil())
		})
	})

	Describe("RemoveRelation", func() {
		It("should return error when removing non-existing relation", func() {
			err := authz.RemoveRelation("user1", "document1", documentType, readerRel)
			Expect(err).NotTo(BeNil())
		})

		It("should return no error when removing existing relation", func() {
			err := authz.AddRelation("user1", "document1", documentType, ownerRel)
			Expect(err).To(BeNil())

			err = authz.RemoveRelation("user1", "document1", documentType, ownerRel)
			Expect(err).To(BeNil())
		})

		It("should not be able to access resource after removing relation", func() {
			err := authz.AddRelation("user1", "document1", documentType, ownerRel)
			Expect(err).To(BeNil())

			ok, err := authz.CheckPermission("user1", "document1", documentType, ownerRel)
			Expect(ok).To(BeTrue())
			Expect(err).To(BeNil())

			err = authz.RemoveRelation("user1", "document1", documentType, ownerRel)
			Expect(err).To(BeNil())

			ok, err = authz.CheckPermission("user1", "document1", documentType, ownerRel)
			Expect(ok).To(BeFalse())
			Expect(err).To(BeNil())
		})
	})

	Describe("ListAccessibleResources", func() {
		It("should return an empty slice and no error", func() {
			resources, err := authz.ListAccessibleResources("user1", documentType, "read", ownerRel)
			Expect(err).To(BeNil())
			Expect(resources).To(BeEmpty())
		})

		It("should return an empty slice for invalid resource type", func() {
			err := authz.AddRelation("user1", "document1", documentType, ownerRel)
			Expect(err).To(BeNil())

			resources, err := authz.ListAccessibleResources("user1", "invalid_type", "read", ownerRel)
			Expect(err).NotTo(BeNil())
			Expect(resources).To(BeEmpty())
		})

		It("should return a list with one resource after adding relation", func() {
			err := authz.AddRelation("user1", "document1", documentType, ownerRel)
			Expect(err).To(BeNil())

			expectedResult := []string{documentType + ":document1"}

			resources, err := authz.ListAccessibleResources("user1", documentType, "read", ownerRel)
			Expect(err).To(BeNil())
			Expect(resources).To(Equal(expectedResult))
		})

		It("should return a list with multiple resources after adding relations", func() {
			err := authz.AddRelation("user1", "document1", documentType, ownerRel)
			Expect(err).To(BeNil())
			err = authz.AddRelation("user1", "document2", documentType, ownerRel)
			Expect(err).To(BeNil())
			err = authz.AddRelation("user1", "document3", documentType, ownerRel)
			Expect(err).To(BeNil())

			expectedResult := []string{
				documentType + ":document1",
				documentType + ":document2",
				documentType + ":document3",
			}

			resources, err := authz.ListAccessibleResources("user1", documentType, "read", ownerRel)
			Expect(err).To(BeNil())
			Expect(resources).To(ConsistOf(expectedResult))
		})

		It("should return an empty slice after removing all relations", func() {
			err := authz.AddRelation("user1", "document1", documentType, ownerRel)
			Expect(err).To(BeNil())
			err = authz.AddRelation("user1", "document2", documentType, ownerRel)
			Expect(err).To(BeNil())

			err = authz.RemoveRelation("user1", "document1", documentType, ownerRel)
			Expect(err).To(BeNil())
			err = authz.RemoveRelation("user1", "document2", documentType, ownerRel)
			Expect(err).To(BeNil())

			resources, err := authz.ListAccessibleResources("user1", documentType, "read", ownerRel)
			Expect(err).To(BeNil())
			Expect(resources).To(BeEmpty())
		})
	})
})
