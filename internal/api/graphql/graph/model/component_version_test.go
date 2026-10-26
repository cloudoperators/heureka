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

// isOutdated is now a persisted field: the mappers pass through the stored
// entity value rather than computing it. The semver computation itself is
// covered by entity.IsVersionOutdated (see internal/entity).
var _ = Describe("NewImageVersion isOutdated passthrough", func() {
	DescribeTable(
		"mirrors the stored entity.IsOutdated and latestVersion",
		func(stored *bool, latest *string) {
			cv := &entity.ComponentVersion{
				Tag:           "irrelevant",
				LatestVersion: latest,
				IsOutdated:    stored,
			}

			iv := model.NewImageVersion(cv)

			Expect(iv.IsOutdated).To(Equal(stored))
			Expect(iv.LatestVersion).To(Equal(latest))
		},

		Entry("stored true", lo.ToPtr(true), lo.ToPtr("2.0.0")),
		Entry("stored false", lo.ToPtr(false), lo.ToPtr("1.0.0")),
		Entry("stored nil (not comparable)", nil, lo.ToPtr("")),
	)
})

var _ = Describe("NewComponentVersion isOutdated passthrough", func() {
	It("mirrors the stored entity.IsOutdated and latestVersion", func() {
		cv := &entity.ComponentVersion{
			Tag:           "irrelevant",
			LatestVersion: lo.ToPtr("v2.0.0"),
			IsOutdated:    lo.ToPtr(true),
		}

		result := model.NewComponentVersion(cv)

		Expect(result.LatestVersion).To(Equal(lo.ToPtr("v2.0.0")))
		Expect(result.IsOutdated).To(Equal(lo.ToPtr(true)))
	})
})
