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

type TestDatabaseManager interface {
	NewTestSchema() *mariadb.SqlDatabase
	Setup() error
	TearDown() error
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
		dbClient, err := mariadb.NewSqlDatabase(dbm.rootUserConfig())
		if err != nil {
			ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while DB Client init")
		}
		dbm.dbClient = dbClient
	}
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
	err = dbm.dbClient.SetupSchema(dbm.Config.Config)
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while setting up management Schema")
		return err
	}

	return nil
}

func (dbm *LocalTestDataBaseManager) ResetSchema() error {
	//first ensure that you can connect
	err := mariadb.TestConnection(dbm.Config.Config, 20)
	if err != nil {
		return err
	}

	//load the client
	dbm.loadDBClientIfNeeded()

	// Drop main heureka schema
	err = dbm.dbClient.DropSchema()

	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while dropping schema")
		return err
	}

	err = dbm.dbClient.SetupSchema(dbm.Config.Config)
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while creating database")
		return err
	}

	return nil
}

func (dbm *LocalTestDataBaseManager) TearDown() error {
	var err error
	for _, schema := range dbm.Schemas {
		err = dbm.dbClient.DropSchemaByName(schema)
	}
	return err
}

func (dbm *LocalTestDataBaseManager) DbConfig() util.Config {
	return dbm.Config.Config
}

func (dbm *LocalTestDataBaseManager) NewTestSchema() *mariadb.SqlDatabase {
	dbm.loadDBClientIfNeeded()

	// using only lowercase characters as in local scenarios the schema name is case-insensitive but the db file names are not leading to errors
	schemaName := fmt.Sprintf("heureka%s", util2.GenerateRandomString(15, util2.Ptr("abcdefghijklmnopqrstuvwxyz0123456789")))
	dbm.Schemas = append(dbm.Schemas, schemaName)
	dbm.Config.DBName = schemaName

	err := dbm.dbClient.SetupSchema(dbm.Config.Config)
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while setting up new Schema")
	}

	err = dbm.dbClient.GrantAccess(dbm.Config.DBUser, schemaName, "%")
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while granting privileges for new Schema ")
	}

	dbClient, err := mariadb.NewSqlDatabase(dbm.Config.Config)
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while loading DB Client for new Schema")
	}

	err = mariadb.TestConnection(dbm.Config.Config, 10)
	if err != nil {
		ginkgo.GinkgoLogr.WithCallDepth(5).Error(err, "Failure while testing connection for new Schema")
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
		Volumes: map[string]struct{}{
			"./internal/database/mariadb/init/schema.sql": struct{}{},
		},

		Env: []string{
			fmt.Sprintf("%s=%s", "MARIADB_USER", dbm.Config.DBUser),
			fmt.Sprintf("%s=%s", "MARIADB_PASSWORD", dbm.Config.DBPassword),
			fmt.Sprintf("%s=%s", "MARIADB_DATABASE", dbm.Config.DBName),
			fmt.Sprintf("%s=%s", "MARIADB_DATABASE", dbm.Config.DBName),
			fmt.Sprintf("%s=%s", "MARIADB_ROOT_PASSWORD", dbm.Config.DBRootPassword),
		},
	}, &container.HostConfig{
		// Other host configuration options...
		Binds: []string{
			fmt.Sprintf("%s:%s", dbm.Config.DBSchema, "/docker-entrypoint-initdb.d/schema.sql"),
		},
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
