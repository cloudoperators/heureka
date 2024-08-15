// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.wdf.sap.corp/cc/heureka/scanners/nvd/models"
)

var _ = Describe("Submitting Issues", Ordered, Label("e2e", "Issues", "create"), func() {
	When("no issues exist", func() {
		Context("and the CVE is valid", func () {
		    It("creates a new Issue", func() {
				newCve := models.Cve{}
				cveJson := `
				{
					"id": "CVE-2023-0001",
					"descriptions": [
						{
							"lang": "en",
							"value": "Sample description of the vulnerability in English."
						},
						{
							"lang": "es",
							"value": "Descripción de ejemplo de la vulnerabilidad en español."
						}
					]
				}
				`
				err := json.Unmarshal([]byte(cveJson), &newCve)
				Expect(err).To(BeNil())

				// Create new Issue
				_, err = cveProcessor.CreateIssue(&newCve)
				Expect(err).To(BeNil())
            })
        })
    })

	When("issues exist", func () {
		Context("and the CVE is valid", func() {
			var issueId string

			It("returns a valid issue_id", func () {
				cve := models.Cve{}
				cveJson := `
				{
					"id": "CVE-2023-0001",
					"descriptions": [
						{
							"lang": "en",
							"value": "Sample description of the vulnerability in English."
						},
						{
							"lang": "es",
							"value": "Descripción de ejemplo de la vulnerabilidad en español."
						}
					],
					"metrics": {},
				}
				`
				err := json.Unmarshal([]byte(cveJson), &cve)
				Expect(err).To(BeNil())

				// Create new IssueVariant
				issueId, err := cveProcesor.GetIssueId(cve)
				Expect(err).To(BeNil())
			})
		})
	})
})
