// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/event/nats"
	"os"
	"os/signal"
	"syscall"

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

	// initialize the database
	db, err := mariadb.NewSqlDatabase(cfg)
	if err != nil {
		logrus.WithError(err).Fatalln("Error while Creating Db")
	}
	defer db.CloseConnection()

	er := nats.NewEventRegistry(db)
	defer er.Shutdown()

	// initialize the application
	application := app.NewHeurekaApp(db, er)
	defer application.Shutdown()

	application.SubscribeHandlers()

	// Create a channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-sigChan
	fmt.Printf("Received signal: %s, shutting down...\n", sig)
}
