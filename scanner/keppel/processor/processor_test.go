// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudoperators/heureka/scanners/keppel/models"
	"github.com/cloudoperators/heureka/scanners/keppel/processor"
)

func TestProcessor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Processor.Keppel Processor Suite")
}

func tags(names ...string) []models.Tag {
	t := make([]models.Tag, 0, len(names))
	for _, n := range names {
		t = append(t, models.Tag{Name: n})
	}
	return t
}

var _ = Describe("HighestSemverTag", func() {
	DescribeTable(
		"returns the highest semantic-version tag, ignoring non-semver tags",
		func(input []models.Tag, expected string) {
			Expect(processor.HighestSemverTag(input)).To(Equal(expected))
		},

		Entry("picks the highest of several semver tags",
			tags("v1.2.3", "v1.3.0", "v1.2.9"), "v1.3.0"),
		Entry("ignores floating and digest tags",
			tags("latest", "v1.2.3", "sha-abc123", "v1.4.0"), "v1.4.0"),
		Entry("handles mixed v-prefixed and bare semver",
			tags("1.2.3", "v1.2.4"), "v1.2.4"),
		Entry("returns empty when no tag parses as semver",
			tags("latest", "stable", "edge"), ""),
		Entry("returns empty for no tags",
			tags(), ""),
		Entry("single semver tag",
			tags("v2.0.0"), "v2.0.0"),
	)
})
