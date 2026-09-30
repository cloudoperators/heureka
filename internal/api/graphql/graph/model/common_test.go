// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package model_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/entity"
)

func TestModel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Model Suite")
}

var _ = Describe("ComponentFriendlyName", func() {
	DescribeTable(
		"returns the correct user-facing name",
		func(componentType string, expected string) {
			Expect(model.ComponentFriendlyName(componentType)).To(Equal(expected))
		},

		Entry("containerImage maps to container image", "containerImage", "container image"),
		Entry("virtualMachineImage maps to virtual machine image", "virtualMachineImage", "virtual machine image"),
		Entry("repository maps to repository", "repository", "repository"),
		Entry("unknown type returns the raw type string", "someOtherType", "someOtherType"),
		Entry("empty type falls back to component", "", "component"),
	)
})

var _ = Describe("NewVulnerability", func() {
	Context("when the issue is in the CISA KEV catalog", func() {
		var (
			addedDate = time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC)
			dueDate   = time.Date(2024, 2, 7, 0, 0, 0, 0, time.UTC)
			issue     = entity.Issue{
				Id:                      42,
				PrimaryName:             "CVE-2024-0519",
				Description:             "Chrome V8 out-of-bounds memory access",
				KnownExploited:          true,
				KnownExploitedAddedDate: &addedDate,
				KnownExploitedDueDate:   &dueDate,
			}
		)

		It("sets KnownExploited to true", func() {
			v := model.NewVulnerability(&issue)
			Expect(v.KnownExploited).NotTo(BeNil())
			Expect(*v.KnownExploited).To(BeTrue())
		})

		It("formats KnownExploitedAddedDate as RFC3339", func() {
			v := model.NewVulnerability(&issue)
			Expect(v.KnownExploitedAddedDate).NotTo(BeNil())
			Expect(*v.KnownExploitedAddedDate).To(Equal("2024-01-17T00:00:00Z"))
		})

		It("formats KnownExploitedDueDate as RFC3339", func() {
			v := model.NewVulnerability(&issue)
			Expect(v.KnownExploitedDueDate).NotTo(BeNil())
			Expect(*v.KnownExploitedDueDate).To(Equal("2024-02-07T00:00:00Z"))
		})
	})

	Context("when the issue is not in the CISA KEV catalog", func() {
		issue := entity.Issue{
			Id:                      7,
			PrimaryName:             "CVE-2023-1234",
			Description:             "some vulnerability",
			KnownExploited:          false,
			KnownExploitedAddedDate: nil,
			KnownExploitedDueDate:   nil,
		}

		It("sets KnownExploited to false", func() {
			v := model.NewVulnerability(&issue)
			Expect(v.KnownExploited).NotTo(BeNil())
			Expect(*v.KnownExploited).To(BeFalse())
		})

		It("leaves KnownExploitedAddedDate nil", func() {
			v := model.NewVulnerability(&issue)
			Expect(v.KnownExploitedAddedDate).To(BeNil())
		})

		It("leaves KnownExploitedDueDate nil", func() {
			v := model.NewVulnerability(&issue)
			Expect(v.KnownExploitedDueDate).To(BeNil())
		})
	})
})
