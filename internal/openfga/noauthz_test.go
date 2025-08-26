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
	)

	BeforeEach(func() {
		cfg = &util.Config{
			AuthzEnabled: false,
		}
		authz = openfga.NewNoAuthz(cfg)
	})

	Describe("NewNoAuthz", func() {
		It("should create a new NoAuthz instance", func() {
			Expect(authz).NotTo(BeNil())
		})
	})

	Describe("CheckPermission", func() {
		It("should always return true and no error", func() {
			ok, err := authz.CheckPermission("user1", "resource1", "read")
			Expect(ok).To(BeTrue())
			Expect(err).To(BeNil())
		})
	})

	Describe("AddRelation", func() {
		It("should always return no error", func() {
			err := authz.AddRelation("user1", "resource1", "member")
			Expect(err).To(BeNil())
		})
	})

	Describe("RemoveRelation", func() {
		It("should always return no error", func() {
			err := authz.RemoveRelation("user1", "resource1", "member")
			Expect(err).To(BeNil())
		})
	})

	Describe("ListAccessibleResources", func() {
		It("should always return an empty slice and no error", func() {
			resources, err := authz.ListAccessibleResources("user1", "resourceType", "read", "member")
			Expect(err).To(BeNil())
			Expect(resources).To(BeEmpty())
		})
	})
})
