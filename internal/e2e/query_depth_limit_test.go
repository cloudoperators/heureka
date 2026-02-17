// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/server"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	queryWithDepthExceeded = `{
  Services {
    edges {
      node {
        supportGroups {
          edges {
            node {
              services {
                edges {
                  node {
                    id
                  }
                }
              }
              users {
                edges {
                  node {
                    id
                    supportGroups {
                      edges {
                        node {
                          id
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`
	queryWithAllowedDepth = `{
  Services {
    edges {
      node {
        supportGroups {
          edges {
            node {
              services {
                edges {
                  node {
                    id
                  }
                }
              }
              users {
                edges {
                  node {
                    id
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`
)

var _ = Describe("Getting data via API", Label("e2e", "Depth Limiting"), func() {
	var s *server.Server
	var cfg util.Config
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchemaWithoutMigration()
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		// Set depth limit as 10 for testing pourpose
		cfg.GQLDepthLimit = 10
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		dbm.TestTearDown(db)
	})

	When("Request with depth exceeding limit", func() {
		It("returns an error", func() {
			_, err := e2e_common.ExecuteGqlQuery[struct {
				Services model.ServiceConnection `json:"Services"`
			}](cfg.Port, queryWithDepthExceeded, nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("operation exceeds the depth limit"))
		})
	})

	When("Request with allowed depth", func() {
		It("doesn't return an error", func() {
			_, err := e2e_common.ExecuteGqlQuery[struct {
				Services model.ServiceConnection `json:"Services"`
			}](cfg.Port, queryWithAllowedDepth, nil)

			Expect(err).ToNot(HaveOccurred())
		})
	})
})
