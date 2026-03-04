// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"testing"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	util2 "github.com/cloudoperators/heureka/internal/util"
	"github.com/cloudoperators/heureka/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	dbConfig util2.Config
	dbm      *test.DatabaseManager
	db       *mariadb.SqlDatabase
)

func TestE2E(t *testing.T) {
	// Set the environment variables
	projectDir, err := util.GetProjectRoot()
	if err != nil {
		panic(err)
	}
	util.SetEnvVars(fmt.Sprintf("%s/%s", projectDir, ".test.env"))

	RegisterFailHandler(Fail)
	RunSpecs(t, "e2e Suite")
}

var _ = BeforeSuite(func() {
	dbm = test.NewDatabaseManager()
})

var _ = AfterSuite(func() {
	Expect(dbm.TearDown()).To(Succeed())
})
