// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/server"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/cloudoperators/heureka/pkg/log"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

var (
	mode string
)

func main() {
	fmt.Println(util.HeurekaFiglet)
	var cfg util.Config
	log.InitLog()

	err := envconfig.Process("heureka", &cfg)
	if err != nil {
		logrus.WithField("error", err).Fatal("Error while reading env config %s", "test")
		return
	}
	cfg.ConfigToConsole()

	if cfg.SeedMode {
		dbManager := test.NewLocalTestDatabaseManager()

		err = dbManager.ResetSchema()
		if err != nil {
			logrus.WithError(err).Fatalln("Error while resetting database schema.")
		}

		err = dbManager.Setup()
		if err != nil {
			logrus.WithError(err).Fatalln("Error while setting up database.")
		}

		seedDb, err := test.NewDatabaseSeeder(cfg)
		if err != nil {
			logrus.WithError(err).Fatalln("Error while initializing database seeder.")
		}

		seedDb.SeedDbForServer(100)
	}

	s := server.NewServer(cfg)
	s.Start()
}
