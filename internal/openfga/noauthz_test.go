// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package openfga_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/cloudoperators/heureka/internal/util"
)

var _ = Describe("NoAuthz", func() {
	var (
		cfg   *util.Config
		authz openfga.Authorization
		r     openfga.RelationInput
	)

	BeforeEach(func() {
		cfg = &util.Config{
			AuthzOpenFgaApiUrl: "",
		}
		enableLogs := true
		r = openfga.RelationInput{
			UserType:   "user",
			UserId:     "user1",
			Relation:   "read",
			ObjectType: "doocument",
			ObjectId:   "document1",
		}

		authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
	})

	Describe("NewNoAuthz", func() {
		It("should create a new NoAuthz instance", func() {
			Expect(authz).NotTo(BeNil())
		})
	})

	Describe("CheckPermission", func() {
		It("should always return true and no error", func() {
			ok, err := authz.CheckPermission(r)
			Expect(ok).To(BeTrue())
			Expect(err).To(BeNil())
		})
	})

	Describe("AddRelation", func() {
		It("should always return no error", func() {
			err := authz.AddRelation(r)
			Expect(err).To(BeNil())
		})
	})

	Describe("RemoveRelation", func() {
		It("should always return no error", func() {
			r.Relation = "member"
			err := authz.RemoveRelation(r)
			Expect(err).To(BeNil())
		})
	})

	Describe("ListAccessibleResources", func() {
		It("should always return an empty slice and no error", func() {
			r.Relation = "member"
			resources, err := authz.ListAccessibleResources(r)
			Expect(err).To(BeNil())
			Expect(resources).To(BeEmpty())
		})
	})
})
