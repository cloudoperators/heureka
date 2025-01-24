// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner_test

import (
	"bytes"
	"compress/gzip"
	"net/http"

	. "github.com/cloudoperators/heureka/scanner/openstack/scanner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

// Setup helper functions and server routes for ghttp server
func setupAuthTokenRoute(server *Server, catalog string) {
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
						"type": catalog,
						"endpoints": []map[string]interface{}{
							{
								"interface": "public",
								"region":    "test-region",
								"url":       server.URL(),
							},
						},
					},
				},
			},
		}, http.Header{"X-Subject-Token": []string{"fake-token"}}),
	))
}

func setupComputeRoute(server *Server) {
	server.RouteToHandler("GET", "/v2.1/", RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{
		"version": map[string]interface{}{
			"id":      "v2.1",
			"status":  "CURRENT",
			"updated": "2023-01-01T00:00:00Z",
			"endpoints": []map[string]interface{}{
				{
					"url": server.URL() + "/v2.1/",
				},
			},
			"links": []map[string]interface{}{
				{"rel": "self", "href": server.URL() + "/v2.1/"},
			},
		},
	}))
}

func setupImageRoute(server *Server) {
	server.RouteToHandler("GET", "/v3/", RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{
		"version": map[string]interface{}{
			"id":      "v2.1",
			"status":  "CURRENT",
			"updated": "2023-01-01T00:00:00Z",
			"endpoints": []map[string]interface{}{
				{
					"url": server.URL() + "/v2.1/",
				},
			},
			"links": []map[string]interface{}{
				{"rel": "self", "href": server.URL() + "/v2.1/"},
			},
		},
	}))
}

func setupGetServersRoute(server *Server) {
	server.RouteToHandler("GET", "/servers/detail", RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{
		"servers": []interface{}{
			map[string]interface{}{
				"OS-DCF:diskConfig": "MANUAL",
				"ID":                "test-id1",
			},
			map[string]interface{}{
				"OS-DCF:diskConfig1": "MANUAL1",
				"ID":                 "test-id2",
			},
		},
	}))
}

func setupGetImage(server *Server) {
	server.RouteToHandler("GET", "/v2/images/test-id", func(w http.ResponseWriter, r *http.Request) {
		jsonData := []byte(`{
			"image": {
				"ID": "image-id",
				"name": "test-image",
				"status": "active",
				"minDisk": 10,
				"minRam": 2048,
				"links": [
					{
						"rel": "self",
						"href": "http://test.com/v2/images/image-id"
					}
				]
			}
		}`)

		gzippedData := gzipJSON(jsonData)

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write(gzippedData)
	})
}

func newTestScanner(serverURL string) *Scanner {
	return &Scanner{
		Username:         "test-user",
		Password:         "test-pass",
		AuthToken:        "",
		Region:           "test-region",
		Domain:           "test-domain",
		Project:          "test-project",
		ProjectId:        "test-projectId",
		ProjectDomain:    "test-domain",
		IdentityEndpoint: serverURL + "/v3/",
	}
}

func gzipJSON(data []byte) []byte {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err := gzipWriter.Write(data)
	if err != nil {
		panic(err)
	}
	gzipWriter.Close()
	return buf.Bytes()
}

var _ = Describe("OpenStack Scanner", func() {
	var (
		scanner *Scanner
		server  *Server
	)

	BeforeEach(func() {
		server = NewServer()
		scanner = newTestScanner(server.URL())
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Compute", func() {
		BeforeEach(func() {
			setupAuthTokenRoute(server, "compute")
			setupComputeRoute(server)
			setupGetServersRoute(server)
		})

		It("should run NewAuthenticatedProviderClient", func() {
			provider, err := scanner.NewAuthenticatedProviderClient()
			Expect(err).ToNot(HaveOccurred())
			Expect(provider).ToNot(BeNil())
			Expect(provider.IdentityEndpoint).To(Equal(server.URL() + "/v3/"))
		})

		It("should run CreateComputeClient", func() {
			service, err := scanner.CreateComputeClient()
			Expect(err).ToNot(HaveOccurred())
			Expect(service).ToNot(BeNil())
			Expect(service.Type).To(Equal("compute"))
		})

		It("should run GetServers", func() {
			service, _ := scanner.CreateComputeClient()
			servers, err := scanner.GetServers(service)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(servers)).To(BeEquivalentTo((2)))
		})
	})

	Describe("Image", func() {
		BeforeEach(func() {
			setupAuthTokenRoute(server, "image")
			setupImageRoute(server)
			setupGetImage(server)
		})

		It("should run CreateImageClient", func() {
			service, err := scanner.CreateImageClient()
			Expect(err).ToNot(HaveOccurred())
			Expect(service).ToNot(BeNil())
			Expect(service.Type).To(Equal("image"))
		})

		It("should run GetImage", func() {
			service, _ := scanner.CreateImageClient()
			images, err := scanner.GetImage(service, "test-id")
			Expect(err).ToNot(HaveOccurred())
			Expect(images.Name).ToNot(BeNil())
		})
	})
})
