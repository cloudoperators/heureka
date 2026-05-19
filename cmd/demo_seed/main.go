// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/cloudoperators/heureka/pkg/log"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

//go:embed seed.sql
var seedSQL string

func main() {
	fmt.Print(util.HeurekaFiglet)

	var cfg util.Config

	log.InitLog()

	if err := envconfig.Process("heureka", &cfg); err != nil {
		logrus.WithError(err).Fatal("Error while reading env config")
	}

	cfg.ConfigToConsole()

	db, err := mariadb.Connect(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}

	defer func() {
		if err := db.Close(); err != nil {
			logrus.Warn("Erorr while closing db connection")
		}
	}()

	logrus.Info("Applying demo seed data...")

	statements := splitStatements(seedSQL)
	applied := 0

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := db.Exec(stmt); err != nil {
			logrus.WithError(err).WithField("statement", truncate(stmt, 120)).Fatal("Failed to execute statement")
		}

		applied++
	}

	logrus.WithField("statements", applied).Info("Demo seed data applied successfully")
}

func splitStatements(sql string) []string {
	var (
		stmts []string
		buf   strings.Builder
	)

	inSingle := false

	for i := 0; i < len(sql); i++ {
		ch := sql[i]

		if ch == '\'' && (i == 0 || sql[i-1] != '\\') {
			inSingle = !inSingle
		}

		if ch == ';' && !inSingle {
			s := strings.TrimSpace(buf.String())
			if s != "" {
				stmts = append(stmts, s)
			}

			buf.Reset()

			continue
		}

		buf.WriteByte(ch)
	}

	if s := strings.TrimSpace(buf.String()); s != "" {
		stmts = append(stmts, s)
	}

	return stmts
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}

	return s[:n] + "..."
}
