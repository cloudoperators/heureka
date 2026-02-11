// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"fmt"
	"testing"

	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var dbm *test.DatabaseManager

func TestMariadb(t *testing.T) {
	// Set the environment variables
	projectDir, err := util.GetProjectRoot()
	if err != nil {
		panic(err)
	}
	_ = util.SetEnvVars(fmt.Sprintf("%s/%s", projectDir, ".test.env"))

	RegisterFailHandler(Fail)
	RunSpecs(t, "MariaDB Suite")
}

var _ = BeforeSuite(func() {
	dbm = test.NewDatabaseManager()
})

var _ = AfterSuite(func() {
	Expect(dbm.TearDown()).To(Succeed())
})
