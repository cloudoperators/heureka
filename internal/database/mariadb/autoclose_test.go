// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Autoclose", Label("database", "Autoclose"), func() {
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		var err error
		db = dbm.NewTestSchema()
		Expect(err).To(BeNil(), "Database Seeder Setup should work")
	})

	When("Running a Canary", Label("Autoclose"), func() {
		Context("and the database is empty", func() {
			It("Autoclose should return false and no error", func() {
				res, err := db.Autoclose()
				Expect(err).To(BeNil())
				Expect(res).To(BeFalse())
			})
		})
	})
})
