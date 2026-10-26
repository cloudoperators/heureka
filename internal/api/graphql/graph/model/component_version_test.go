// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package model_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
)

var _ = Describe("NewImageVersion isOutdated", func() {
	DescribeTable(
		"derives isOutdated from the running tag and the latest registry version",
		func(tag string, latest *string, expected *bool) {
			cv := &entity.ComponentVersion{
				Tag:           tag,
				LatestVersion: latest,
			}

			iv := model.NewImageVersion(cv)

			if expected == nil {
				Expect(iv.IsOutdated).To(BeNil())
			} else {
				Expect(iv.IsOutdated).ToNot(BeNil())
				Expect(*iv.IsOutdated).To(Equal(*expected))
			}
			// latestVersion is passed through unchanged
			Expect(iv.LatestVersion).To(Equal(latest))
		},

		Entry("running tag behind latest -> outdated", "v1.2.3", lo.ToPtr("v1.3.0"), lo.ToPtr(true)),
		Entry("running tag equals latest -> not outdated", "v1.3.0", lo.ToPtr("v1.3.0"), lo.ToPtr(false)),
		Entry("running tag ahead of latest -> not outdated", "v1.3.0", lo.ToPtr("v1.2.0"), lo.ToPtr(false)),
		Entry("patch bump -> outdated", "1.2.3", lo.ToPtr("1.2.4"), lo.ToPtr(true)),
		Entry("non-semver running tag -> unknown", "latest", lo.ToPtr("v1.3.0"), nil),
		Entry("non-semver latest -> unknown", "v1.2.3", lo.ToPtr("rolling"), nil),
		Entry("no latest version known -> unknown", "v1.2.3", nil, nil),
		Entry("empty latest version -> unknown", "v1.2.3", lo.ToPtr(""), nil),
		Entry("empty running tag -> unknown", "", lo.ToPtr("v1.3.0"), nil),
	)
})

var _ = Describe("NewComponentVersion isOutdated", func() {
	It("computes isOutdated and passes through latestVersion", func() {
		cv := &entity.ComponentVersion{
			Tag:           "v1.2.3",
			LatestVersion: lo.ToPtr("v2.0.0"),
		}

		result := model.NewComponentVersion(cv)

		Expect(result.LatestVersion).To(Equal(lo.ToPtr("v2.0.0")))
		Expect(result.IsOutdated).ToNot(BeNil())
		Expect(*result.IsOutdated).To(BeTrue())
	})
})
