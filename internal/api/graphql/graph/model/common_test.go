// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package model_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
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
