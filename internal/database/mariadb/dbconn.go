// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/util"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func GetSqlxConnection(cfg util.Config) (*sqlx.DB, error) {
	return sqlx.Connect("mysql", buildUserDSN(cfg))
}

func GetSqlxRootConnection(cfg util.Config) (*sqlx.DB, error) {
	return sqlx.Connect("mysql", buildRootDSN(cfg))
}

func buildUserDSN(cfg util.Config) string {
	return buildDSN(cfg.DBUser, cfg.DBPassword, cfg)
}

func buildRootDSN(cfg util.Config) string {
	return buildDSN("root", cfg.DBRootPassword, cfg)
}

func buildDSN(user string, pass string, cfg util.Config) string {
	if cfg.DBAddress == "/var/run/mysqld/mysqld.sock" {
		return fmt.Sprintf("%s:%s@unix(%s)/%s?multiStatements=true&parseTime=true", user, pass, cfg.DBAddress, cfg.DBName)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true&parseTime=true", user, pass, cfg.DBAddress, cfg.DBPort, cfg.DBName)
}

func TestConnection(cfg util.Config, backOff int) error {
	if cfg.DBAddress == "/var/run/mysqld/mysqld.sock" {
		// No need to test local socket connection
		return nil
	}
	if backOff <= 0 {
		return fmt.Errorf("Unable to connect to Database, exceeded backoffs...")
	}

	db, err := GetSqlxConnection(cfg)
	if err != nil {
		fmt.Printf("Error connecting to DB: %s\n", err)
		return TestConnection(cfg, backOff-1)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		//before next try wait 100 milliseconds
		time.Sleep(100 * time.Millisecond)
		return TestConnection(cfg, backOff-1)
	}
	return nil
}

func Connect(cfg util.Config) (*sqlx.DB, error) {
	db, err := GetSqlxConnection(cfg)
	if err != nil {
		logrus.WithError(err).Error(err)
		return nil, err
	}

	db.SetConnMaxLifetime(time.Second * 5)
	db.SetMaxIdleConns(cfg.DBMaxIdleConnections)
	db.SetMaxOpenConns(cfg.DBMaxOpenConnections)
	return db, nil
}
