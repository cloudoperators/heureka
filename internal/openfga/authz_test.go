// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package openfga_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/cloudoperators/heureka/internal/util"
)

const (
	testModelFilePath = "./testdata/testModel.fga"
	testStoreName     = "heureka_test_store1"

	documentType = "document"
	userType     = "user"
	readerRel    = "reader"
	writerRel    = "writer"
	ownerRel     = "owner"
)

var (
	cfg   *util.Config
	authz openfga.Authorization
	p     openfga.PermissionInput
	r     openfga.RelationInput
)

var _ = BeforeSuite(func() {
	if os.Getenv("AUTHZ_FGA_API_URL") == "" {
		Skip("Skipping OpenFGA authz tests because AUTHZ_FGA_API_URL is not set")
	}

	cfg = &util.Config{
		AuthzOpenFgaApiUrl:    os.Getenv("AUTHZ_FGA_API_URL"),
		AuthzOpenFgaApiToken:  os.Getenv("AUTHZ_FGA_API_TOKEN"),
		AuthzModelFilePath:    testModelFilePath,
		AuthzOpenFgaStoreName: testStoreName,
	}
})

var _ = Describe("Authz", func() {
	BeforeEach(func() {
		p = openfga.PermissionInput{
			UserType:   userType,
			UserId:     "user1",
			Relation:   readerRel,
			ObjectType: documentType,
			ObjectId:   "document1",
		}
		r = openfga.RelationInput{
			UserType:   userType,
			UserId:     "user1",
			Relation:   readerRel,
			ObjectType: documentType,
			ObjectId:   "document1",
		}
		authz = openfga.NewAuthorizationHandler(cfg, true)
	})

	Describe("NewAuthz", func() {
		It("should create a new Authz instance", func() {
			Expect(authz).NotTo(BeNil())
		})
	})

	Describe("CheckPermission", func() {
		It("should return false with no relations added", func() {
			ok, err := authz.CheckPermission(p)
			Expect(ok).To(BeFalse())
			Expect(err).To(BeNil())
		})

		It("should return an error for invalid resource type", func() {
			ok, err := authz.CheckPermission(p)
			Expect(ok).To(BeFalse())
			Expect(err).To(BeNil())
		})

		It("should return true after adding relation", func() {
			err := authz.AddRelation(r)
			Expect(err).To(BeNil())

			ok, err := authz.CheckPermission(p)
			Expect(ok).To(BeTrue())
			Expect(err).To(BeNil())
		})

		It("should return false after adding wrong relation", func() {
			err := authz.AddRelation(r)
			Expect(err).To(BeNil())

			p.Relation = ownerRel
			ok, err := authz.CheckPermission(p)
			Expect(ok).To(BeFalse())
			Expect(err).To(BeNil())
		})
	})

	Describe("AddRelation", func() {
		It("should return no error", func() {
			r.Relation = writerRel
			err := authz.AddRelation(r)
			Expect(err).To(BeNil())
		})

		It("should return error when adding invalid relation", func() {
			r.Relation = "invalid_relation"
			err := authz.AddRelation(r)
			Expect(err).NotTo(BeNil())
		})

		It("should return error when adding relation to invalid object type", func() {
			r.Relation = writerRel
			r.ObjectType = "invalid_type"
			err := authz.AddRelation(r)
			Expect(err).NotTo(BeNil())
		})

		It("should return error when adding relation with empty userId", func() {
			r.UserId = ""
			r.Relation = writerRel
			err := authz.AddRelation(r)
			Expect(err).NotTo(BeNil())
		})

		It("should be able to access resource after adding relation", func() {
			r.Relation = ownerRel
			err := authz.AddRelation(r)
			Expect(err).To(BeNil())

			p.Relation = ownerRel
			ok, err := authz.CheckPermission(p)
			Expect(ok).To(BeTrue())
			Expect(err).To(BeNil())
		})
	})

	Describe("RemoveRelation", func() {
		It("should do nothing when removing non-existing relation", func() {
			err := authz.RemoveRelation(r)
			Expect(err).To(BeNil())
		})

		It("should return no error when removing existing relation", func() {
			r.Relation = ownerRel
			err := authz.AddRelation(r)
			Expect(err).To(BeNil())

			err = authz.RemoveRelation(r)
			Expect(err).To(BeNil())
		})

		It("should not be able to access resource after removing relation", func() {
			r.Relation = ownerRel
			err := authz.AddRelation(r)
			Expect(err).To(BeNil())

			p.Relation = ownerRel
			ok, err := authz.CheckPermission(p)
			Expect(ok).To(BeTrue())
			Expect(err).To(BeNil())

			err = authz.RemoveRelation(r)
			Expect(err).To(BeNil())

			ok, err = authz.CheckPermission(p)
			Expect(ok).To(BeFalse())
			Expect(err).To(BeNil())
		})
	})

	Describe("ListAccessibleResources", func() {
		It("should return an empty slice and no error", func() {
			p.ObjectId = "read"
			resources, err := authz.ListAccessibleResources(p)
			Expect(err).To(BeNil())
			Expect(resources).To(BeEmpty())
		})

		It("should return an empty slice for invalid resource type", func() {
			r.Relation = ownerRel
			err := authz.AddRelation(r)
			Expect(err).To(BeNil())

			p.ObjectType = "invalid_type"
			p.ObjectId = "read"
			p.Relation = ownerRel
			resources, err := authz.ListAccessibleResources(p)
			Expect(err).NotTo(BeNil())
			Expect(resources).To(BeEmpty())
		})

		It("should return a list with one resource after adding relation", func() {
			r.Relation = ownerRel
			err := authz.AddRelation(r)
			Expect(err).To(BeNil())

			expectedResult := []openfga.AccessibleResource{
				{ObjectType: documentType, ObjectId: "document1"},
			}

			p.Relation = ownerRel
			p.ObjectId = "read"
			resources, err := authz.ListAccessibleResources(p)
			Expect(err).To(BeNil())
			Expect(resources).To(Equal(expectedResult))
		})

		It("should return a list with multiple resources after adding relations", func() {
			r.Relation = ownerRel
			r.ObjectId = "document1"
			err := authz.AddRelation(r)
			Expect(err).To(BeNil())
			r.ObjectId = "document2"
			err = authz.AddRelation(r)
			Expect(err).To(BeNil())
			r.ObjectId = "document3"
			err = authz.AddRelation(r)
			Expect(err).To(BeNil())

			expectedResult := []openfga.AccessibleResource{
				{ObjectType: documentType, ObjectId: "document1"},
				{ObjectType: documentType, ObjectId: "document2"},
				{ObjectType: documentType, ObjectId: "document3"},
			}

			p.Relation = ownerRel
			p.ObjectId = "read"
			resources, err := authz.ListAccessibleResources(p)
			Expect(err).To(BeNil())
			Expect(resources).To(ConsistOf(expectedResult))
		})

		It("should return an empty slice after removing all relations", func() {
			r.Relation = ownerRel
			err := authz.AddRelation(r)
			Expect(err).To(BeNil())
			r.ObjectId = "document2"
			err = authz.AddRelation(r)
			Expect(err).To(BeNil())
			r.ObjectId = "document3"
			err = authz.AddRelation(r)
			Expect(err).To(BeNil())

			r.ObjectId = "document1"
			err = authz.RemoveRelation(r)
			Expect(err).To(BeNil())
			r.ObjectId = "document2"
			err = authz.RemoveRelation(r)
			Expect(err).To(BeNil())
			r.ObjectId = "document3"
			err = authz.RemoveRelation(r)
			Expect(err).To(BeNil())

			p.Relation = ownerRel
			p.ObjectId = "read"
			resources, err := authz.ListAccessibleResources(p)
			Expect(err).To(BeNil())
			Expect(resources).To(BeEmpty())
		})
	})
})
