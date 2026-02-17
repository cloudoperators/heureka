// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/server"
	"github.com/machinebox/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	queryWithBatchLimitExceeded = `{
  Services {
    edges {
      node {
        ccrn
      }
    }
  }
  Components {
    edges {
      node {
        ccrn
      }
    }
  }
  ComponentInstances {
    edges {
      node {
        ccrn
      }
    }
  }
}`
	queryWithAllowedBatchLimit = `{
  Services {
    edges {
      node {
        ccrn
      }
    }
  }
  Components {
    edges {
      node {
        ccrn
      }
    }
  }
}`
)

var _ = Describe("Getting data via API", Label("e2e", "Batch Limiting"), func() {
	var s *server.Server
	var cfg util.Config
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchemaWithoutMigration()
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		// Set batch limit as 2 for testing pourpose
		cfg.GQLBatchLimit = 2
		s = e2e_common.NewRunningServer(cfg)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		dbm.TestTearDown(db)
	})

	When("Request with batch exceeding limit", func() {
		It("returns an error", func() {
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))
			req := graphql.NewRequest(queryWithBatchLimitExceeded)
			req.Header.Set("Cache-Control", "no-cache")
			_, err := e2e_common.ExecuteGqlQuery[struct {
				Services           model.ServiceConnection           `json:"Services"`
				Components         model.ComponentConnection         `json:"Components"`
				ComponentInstances model.ComponentInstanceConnection `json:"ComponentInstances"`
			}](client, req)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the limit for sending batches has been exceeded"))
		})
	})

	When("Request with allowed batch limit", func() {
		It("doesn't return an error", func() {
			client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", cfg.Port))
			req := graphql.NewRequest(queryWithAllowedBatchLimit)

			req.Header.Set("Cache-Control", "no-cache")

			_, err := e2e_common.ExecuteGqlQuery[struct {
				Services   model.ServiceConnection   `json:"Services"`
				Components model.ComponentConnection `json:"Components"`
			}](client, req)

			Expect(err).ToNot(HaveOccurred())
		})
	})
})
