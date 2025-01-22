// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner_test

import (
	"net/http"
	"testing"

	. "github.com/cloudoperators/heureka/scanner/openstack/scanner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

func TestScanner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scanner.Openstack Scanner Suite")
}

var _ = Describe("NewAuthenticatedProviderClient", func() {
	var (
		scanner *Scanner
		server  *Server
	)

	BeforeEach(func() {
		server = NewServer()
		scanner = &Scanner{
			Username:         "test-user",
			Password:         "test-pass",
			AuthToken:        "",
			Region:           "test-region",
			Domain:           "test-domain",
			Project:          "test-project",
			ProjectId:        "test-projectId",
			ProjectDomain:    "test-domain",
			IdentityEndpoint: server.URL() + "/v3/",
		}
	})

	AfterEach(func() {
		server.Close()
	})

	It("should create a new authenticated provider client", func() {
		// Create test ProviderClient used for service clients like compute to run authenticated requests
		server.RouteToHandler("POST", "/v3/auth/tokens", CombineHandlers(
			VerifyJSON(`{
                "auth": {
                    "identity": {
                        "methods": ["password"],
                        "password": {
                            "user": {
                                "name": "test-user",
                                "password": "test-pass",
                                "domain": { "name": "test-domain" }
                            }
                        }
                    },
                    "scope": {
                        "project": {
                            "name": "test-project",
                            "domain": { "name": "test-domain" }
                        }
                    }
                }
            }`),
			RespondWithJSONEncoded(201, map[string]interface{}{
				"token": map[string]interface{}{
					"catalog": []map[string]interface{}{
						{
							"type": "compute",
							"endpoints": []map[string]interface{}{
								{
									"interface": "public",
									"region":    "test-region",
									// Static ghttp server URL
									"url": server.URL(),
								},
							},
						},
					},
				},
			}, http.Header{"X-Subject-Token": []string{"fake-token"}}),
		))

		provider, err := scanner.NewAuthenticatedProviderClient()
		Expect(err).ToNot(HaveOccurred())
		Expect(provider).ToNot(BeNil())
		Expect(provider.IdentityEndpoint).To(Equal(server.URL() + "/v3/"))
	})
})
