// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb"
	"github.wdf.sap.corp/cc/heureka/internal/database/mariadb/test"
	util2 "github.wdf.sap.corp/cc/heureka/internal/util"
	"github.wdf.sap.corp/cc/heureka/pkg/util"
	"os"
	"testing"
)

var dbConfig util2.Config
var dbm test.TestDatabaseManager
var db *mariadb.SqlDatabase

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
	var err error
	backOff := 20

	localTestDB := os.Getenv("LOCAL_TEST_DB")

	if localTestDB != "true" {
		cDbm := test.NewContainerizedTestDatabaseManager()

		Expect(cDbm.Setup()).To(Succeed(), "Setup of containerized test database should work")
		//set dbConfig after Setup

		// We test the connection with n(backoff) amounts of tries in a 500ms interval
		Expect(mariadb.TestConnection(cDbm.Config.Config, backOff)).To(Succeed(), fmt.Sprintf("Database should be reachable within %d Seconds", backOff/2))

		Expect(err).To(BeNil(), "Expecting Containerized Database initialization to be completed")
		dbm = cDbm
	} else {
		lDbm := test.NewLocalTestDatabaseManager()

		Expect(lDbm.Setup()).To(Succeed(), "Setup of local test database should work")
		// We test the connection with n(backoff) amounts of tries in a 500ms interval
		Expect(mariadb.TestConnection(lDbm.Config.Config, backOff)).To(Succeed(), fmt.Sprintf("Database should be reachable within %d Seconds", backOff/2))

		Expect(err).To(BeNil(), "Expecting Containerized Database initialization to be completed")
		dbm = lDbm
	}
})

var _ = AfterSuite(func() {
	Expect(dbm.TearDown()).To(Succeed())
})
