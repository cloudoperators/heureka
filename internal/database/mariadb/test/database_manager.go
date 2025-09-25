// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"
	osx "github.com/docker/docker-credential-helpers/client"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kelseyhightower/envconfig"
	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
)

const (
	MARIADB_DEFAULT_PORT = "3306/tcp"
)

func NewDatabaseManager() (TestDatabaseManager, error) {
	backOff := 20
	localTestDB := os.Getenv("LOCAL_TEST_DB")
	var tdm TestDatabaseManager

	if localTestDB != "true" {
		tdm = NewContainerizedTestDatabaseManager()
	} else {
		tdm = NewLocalTestDatabaseManager()
	}
	// We test the connection with n(backoff) amounts of tries in a 500ms interval
	if err := mariadb.TestConnection(tdm.DbConfig(), backOff); err != nil {
		return nil, fmt.Errorf("Database should be reachable within %d Seconds: %w", backOff/2, err)
	}
	return tdm, nil
}

type TestDatabaseManager interface {
	NewTestSchema() *mariadb.SqlDatabase
	NewTestSchemaWithoutMigration() *mariadb.SqlDatabase
	Setup() error
	TearDown() error
	TestTearDown(*mariadb.SqlDatabase) error
	DbConfig() util.Config
}
type LocalTestDatabaseConfig struct {
	util.Config
}

type LocalTestDataBaseManager struct {
	Config   *LocalTestDatabaseConfig
	Schemas  []string
	dbClient *mariadb.SqlDatabase
}

func NewLocalTestDatabaseManager() *LocalTestDataBaseManager {
	tdbm := LocalTestDataBaseManager{}
	loadConfig(&tdbm.Config)
	tdbm.Config.DBName = ""
	return &tdbm
}

func (dbm *LocalTestDataBaseManager) rootUserConfig() util.Config {
	cfg := dbm.Config.Config
	cfg.DBUser = "root"
	cfg.DBPassword = cfg.DBRootPassword
	return cfg
}

func (dbm *LocalTestDataBaseManager) loadDBClientIfNeeded() {
	if dbm.dbClient == nil {
		dbm.loadDBClient()
	}
}

func (dbm *LocalTestDataBaseManager) loadDBClient() {
	dbClient, err := mariadb.NewSqlDatabase(dbm.rootUserConfig())
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while DB Client init")
	}
	dbm.dbClient = dbClient
}

func loadConfig[T any](config **T) {
	var cfg T
	err := envconfig.Process("heureka", &cfg)
	if err != nil {
		panic(err)
	}
	*config = &cfg
}

func (dbm *LocalTestDataBaseManager) Setup() error {
	//first ensure that you can connect
	err := mariadb.TestConnection(dbm.Config.Config, 20)
	if err != nil {
		return err
	}

	//load the client
	dbm.loadDBClientIfNeeded()

	//setup base schema to ensure schema loading works
	err = dbm.dbClient.RunUpMigrations()
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while setting up migrations Schema")
		return err
	}

	dbm.loadDBClient()
	err = dbm.dbClient.RunPostMigrations()
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while calling post migration procedures")
		return err
	}

	return nil
}

func (dbm *LocalTestDataBaseManager) ResetSchema(dbName string) error {
	//first ensure that you can connect
	err := mariadb.TestConnection(dbm.Config.Config, 20)
	if err != nil {
		return err
	}

	//load the client
	dbm.loadDBClientIfNeeded()

	// Drop main heureka schema
	err = dbm.dbClient.DropSchema(dbName)
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while dropping schema")
		return err
	}

	err = dbm.prepareDb(dbName)
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failed to prepare DB schema")
		return err
	}

	err = dbm.dbClient.RunUpMigrations()
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while creating database")
		return err
	}
	dbm.loadDBClient()
	err = dbm.dbClient.RunPostMigrationsNoClose()
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while calling post migration procedures")
		return err
	}

	dbm.dbClient, err = dbm.createNewConnection()
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failed to create db connection")
		return err
	}
	return nil
}

func (dbm *LocalTestDataBaseManager) TearDown() error {
	return dbm.cleanupSchemas()
}

func removeValue[T comparable](s []T, value T) []T {
	for i, v := range s {
		if v == value {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s // value not found
}

func (dbm *LocalTestDataBaseManager) cleanupSchemas() error {
	var err error
	for _, schema := range dbm.Schemas {
		err = dbm.dbClient.DropSchema(schema)
	}
	dbm.Schemas = []string{}
	return err
}

func (dbm *LocalTestDataBaseManager) TestTearDown(dbClient *mariadb.SqlDatabase) error {
	err := dbm.cleanupSchemas()
	dbClient.CloseConnection()
	if dbm.dbClient != nil {
		dbm.dbClient.CloseConnection()
		dbm.dbClient = nil
	}
	return err
}

func (dbm *LocalTestDataBaseManager) DbConfig() util.Config {
	return dbm.Config.Config
}

func (dbm *LocalTestDataBaseManager) prepareTestDb() error {
	return dbm.prepareDb(fmt.Sprintf("heureka%s", util2.GenerateRandomString(15, util2.Ptr("abcdefghijklmnopqrstuvwxyz0123456789"))))
}

func (dbm *LocalTestDataBaseManager) prepareDb(dbName string) error {
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

func (dbm *LocalTestDataBaseManager) createNewConnection() (*mariadb.SqlDatabase, error) {
	dbClient, err := mariadb.NewSqlDatabase(dbm.Config.Config)
	if err != nil {
		return nil, fmt.Errorf("Failure while loading DB Client for new Schema")
	}

	err = mariadb.TestConnection(dbm.Config.Config, 10)
	if err != nil {
		return nil, fmt.Errorf("Failure while testing connection for new Schema")
	}

	return dbClient, nil
}

func (dbm *LocalTestDataBaseManager) NewTestSchema() *mariadb.SqlDatabase {
	err := dbm.prepareTestDb()
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failed to prepare DB schema")
	}

	err = dbm.dbClient.RunUpMigrations()
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure during DB migration")
	}

	dbClient, err := dbm.createNewConnection()
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failed to create db connection")
	}
	err = dbClient.RunPostMigrationsNoClose()
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while calling post migration procedures")
	}

	return dbClient
}

func (dbm *LocalTestDataBaseManager) NewTestSchemaWithoutMigration() *mariadb.SqlDatabase {
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

type TestContainerizedDataBaseConfig struct {
	util.Config
	DBImage               string `envconfig:"DB_CONTAINER_IMAGE" required:"true" json:"dbContainerImage"`
	DockerImageRegistry   string `envconfig:"DOCKER_IMAGE_REGISTRY" required:"true" json:"DockerImageRegistry"`
	DockerCredentialStore string `envconfig:"DOCKER_CREDENTIAL_STORE" required:"true" json:"dockerCredentialStore"`
}

type ContainerizedTestDataBaseManager struct {
	*LocalTestDataBaseManager
	Config      *TestContainerizedDataBaseConfig
	Cli         *client.Client
	ContainerId string
}

func NewContainerizedTestDatabaseManager() *ContainerizedTestDataBaseManager {
	tdbm := &ContainerizedTestDataBaseManager{}
	loadConfig(&tdbm.Config)
	tdbm.Config.DBName = ""
	tdbm.LocalTestDataBaseManager = NewLocalTestDatabaseManager()
	return tdbm
}

func (dbm *ContainerizedTestDataBaseManager) getDockerAuthString() (string, error) {
	p := osx.NewShellProgramFunc(dbm.Config.DockerCredentialStore)

	creds, err := osx.Get(p, dbm.Config.DockerImageRegistry)
	if err != nil {
		return "", err
	}

	authConfig := registry.AuthConfig{
		Username: creds.Username,
		Password: creds.Secret,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	return authStr, nil
}

func (dbm *ContainerizedTestDataBaseManager) Setup() error {
	name := util2.GenerateRandomString(10, nil)
	l := logrus.WithField("name", name)
	authStr, err := dbm.getDockerAuthString()
	if err != nil {
		l.WithError(err).Error("Error while setting up Authstring")
		return err
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		l.WithError(err).Error("Error while initiating client")
		return err
	}
	defer cli.Close()

	imgPull := image.PullOptions{}

	if authStr != "" {
		imgPull.RegistryAuth = authStr
	}

	reader, err := cli.ImagePull(ctx, dbm.Config.DBImage, imgPull)
	if err != nil {
		l.WithError(err).Error("Error while pulling image")
		return err
	}

	dbm.Config.DBPort = util2.GetRandomFreePort()

	defer reader.Close()
	io.Copy(os.Stdout, reader)
	containerCfg := &container.Config{}
	containerCfg.ExposedPorts = nat.PortSet{}
	createResp, err := cli.ContainerCreate(ctx, &container.Config{
		Labels: map[string]string{
			"TestSuite": "MariaDb",
		},
		Image: dbm.Config.DBImage,
		ExposedPorts: nat.PortSet{
			MARIADB_DEFAULT_PORT: struct{}{},
		},
		Env: []string{
			fmt.Sprintf("%s=%s", "MARIADB_USER", dbm.Config.DBUser),
			fmt.Sprintf("%s=%s", "MARIADB_PASSWORD", dbm.Config.DBPassword),
			fmt.Sprintf("%s=%s", "MARIADB_DATABASE", dbm.Config.DBName),
			fmt.Sprintf("%s=%s", "MARIADB_ROOT_PASSWORD", dbm.Config.DBRootPassword),
		},
	}, &container.HostConfig{
		// Other host configuration options...
		PortBindings: nat.PortMap{
			MARIADB_DEFAULT_PORT: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",         // Bind to all available network interfaces
					HostPort: dbm.Config.DBPort, // Local port to bind --- need to replace with random port
				},
			},
		},
	}, nil, nil, fmt.Sprintf("Test_MariaDB_%s", name))
	if err != nil {
		l.WithError(err).Error("Error while creating container")
		return err
	}
	dbm.ContainerId = createResp.ID

	if err := cli.ContainerStart(ctx, dbm.ContainerId, container.StartOptions{}); err != nil {
		return err
	}

	//Waiting at least 3 seconds for DB to spawn and beeing reachable
	time.Sleep(time.Second * 3)
	ginkgo.GinkgoLogr.Info("Waiting 3 Seconds for DB to Init")

	//replace with updated values
	dbm.LocalTestDataBaseManager.Config.Config = dbm.Config.Config
	return dbm.LocalTestDataBaseManager.Setup()
}

func (dbm *ContainerizedTestDataBaseManager) TearDown() error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	noWaitTimeout := 0 // to not wait for the container to exit gracefully
	if err := cli.ContainerStop(ctx, dbm.ContainerId, container.StopOptions{Timeout: &noWaitTimeout}); err != nil {
		return err
	}

	err = cli.ContainerRemove(ctx, dbm.ContainerId, container.RemoveOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (dbm *ContainerizedTestDataBaseManager) DbConfig() util.Config {
	return dbm.LocalTestDataBaseManager.Config.Config
}
