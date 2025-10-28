// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"
	"github.com/kelseyhightower/envconfig"
	"github.com/onsi/ginkgo/v2"
)

const (
	MARIADB_DEFAULT_PORT = "3306/tcp"
)

type DatabaseManager struct {
	Config   util.Config
	Schemas  []string
	dbClient *mariadb.SqlDatabase
}

func NewDatabaseManager() *DatabaseManager {
	dbm := DatabaseManager{}
	err := envconfig.Process("heureka", &dbm.Config)
	if err != nil {
		panic(err)
	}
	dbm.Config.DBName = ""
	return &dbm
}

func (dbm *DatabaseManager) rootUserConfig() util.Config {
	cfg := dbm.Config
	cfg.DBUser = "root"
	cfg.DBPassword = cfg.DBRootPassword
	cfg.DBName = ""
	return cfg
}

func (dbm *DatabaseManager) loadDBClientIfNeeded() {
	if dbm.dbClient == nil {
		dbm.loadDBClient()
	}
}

func (dbm *DatabaseManager) loadDBClient() {
	dbClient, err := mariadb.NewSqlDatabase(dbm.rootUserConfig())
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while DB Client init")
	}
	dbm.dbClient = dbClient
}

func (dbm *DatabaseManager) Setup() error {
	err := mariadb.RunMigrationsSync(dbm.rootUserConfig())
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while setting migrations schema")
		return err
	}

	return nil
}

func (dbm *DatabaseManager) ResetSchema(dbName string) error {
	//load the client
	dbm.loadDBClientIfNeeded()

	// Drop main heureka schema
	err := dbm.dbClient.DropSchema(dbName)
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while dropping schema")
		return err
	}

	err = dbm.prepareDb(dbName)
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failed to prepare DB schema")
		return err
	}

	err = mariadb.RunMigrationsSync(dbm.DbConfig())
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while resetting migrations schema")
		return err
	}

	dbm.dbClient, err = dbm.createNewConnection()
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failed to create db connection")
		return err
	}
	return nil
}

func (dbm *DatabaseManager) TearDown() error {
	return dbm.cleanupSchemas()
}

func (dbm *DatabaseManager) cleanupSchemas() error {
	var err error
	for _, schema := range dbm.Schemas {
		err = dbm.dbClient.DropSchema(schema)
	}
	dbm.Schemas = []string{}
	return err
}

func (dbm *DatabaseManager) TestTearDown(dbClient *mariadb.SqlDatabase) error {
	err := dbm.cleanupSchemas()
	dbClient.CloseConnection()
	if dbm.dbClient != nil {
		dbm.dbClient.CloseConnection()
		dbm.dbClient = nil
	}
	return err
}

func (dbm *DatabaseManager) DbConfig() util.Config {
	return dbm.Config
}

func (dbm *DatabaseManager) prepareTestDb() error {
	return dbm.prepareDb(fmt.Sprintf("heureka%s", util2.GenerateRandomString(15, util2.Ptr("abcdefghijklmnopqrstuvwxyz0123456789"))))
}

func (dbm *DatabaseManager) prepareDb(dbName string) error {
	dbm.loadDBClientIfNeeded()

	// using only lowercase characters as in local scenarios the schema name is case-insensitive but the db file names are not leading to errors
	dbm.Config.DBName = dbName
	dbm.Schemas = append(dbm.Schemas, dbm.Config.DBName)

	err := dbm.dbClient.ConnectDB(dbm.Config.DBName)
	if err != nil {
		return fmt.Errorf("Failure while connecting to DB")
	}
	err = dbm.dbClient.GrantAccess(dbm.Config.DBUser, dbm.Config.DBName, "%")
	if err != nil {
		return fmt.Errorf("Failure while granting privileges for new Schema")
	}
	return nil
}

func (dbm *DatabaseManager) createNewConnection() (*mariadb.SqlDatabase, error) {
	dbClient, err := mariadb.NewSqlDatabase(dbm.Config)
	if err != nil {
		return nil, fmt.Errorf("Failure while loading DB Client for new Schema")
	}

	return dbClient, nil
}

func (dbm *DatabaseManager) NewTestSchema() *mariadb.SqlDatabase {
	err := dbm.prepareTestDb()
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failed to prepare DB schema")
	}

	err = mariadb.RunMigrationsSync(dbm.DbConfig())
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while resetting migrations schema")
	}

	resultDbClient, err := dbm.createNewConnection()
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failed to create db connection")
	}
	return resultDbClient
}

func (dbm *DatabaseManager) NewTestSchemaWithoutMigration() *mariadb.SqlDatabase {
	err := dbm.prepareTestDb()
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failed to prepare db schema")
	}

	err = dbm.dbClient.CloseConnection()
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while closing DB connection")
	}

	dbClient, err := dbm.createNewConnection()
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failed to create db connection")
	}
	return dbClient
}
