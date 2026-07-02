// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"embed"
	"fmt"
	"io"
	"io/fs"

	"github.com/cloudoperators/heureka/internal/util"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS
var MigrationFs fs.FS = &migrationFiles

func GetVersion(cfg util.Config) (string, error) {
	db, err := NewSqlDatabase(cfg)
	if err != nil {
		return "", fmt.Errorf("error while Creating Db")
	}

	v, err := db.getVersion()
	if err != nil {
		return "", fmt.Errorf("error while checking Db migration: %w", err)
	}

	return v, nil
}

func (s *SqlDatabase) getVersion() (string, error) {
	m, err := s.openMigration()
	if err != nil {
		return "", fmt.Errorf("could not open migration without source: %w", err)
	}

	defer func() {
		_, _ = m.Close()
	}()

	v, d, err := getMigrationVersion(m)
	if err != nil {
		return "", fmt.Errorf("could not get migration version: %w", err)
	}

	return versionToString(v, d), nil
}

func (s *SqlDatabase) openMigration() (*migrate.Migrate, error) {
	err := s.connectDB()
	if err != nil {
		return nil, fmt.Errorf("could not connect DB: %w", err)
	}

	d, err := iofs.New(MigrationFs, "migrations")
	if err != nil {
		return nil, err
	}

	driver, err := mysql.WithInstance(s.db.GetDbInstance(), &mysql.Config{DatabaseName: s.dbName})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance(
		"iofs", d,
		"mysql", driver,
	)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func getMigrationVersion(m *migrate.Migrate) (uint, bool, error) {
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, err
	}

	return version, dirty, nil
}

func versionToString(v uint, dirty bool) string {
	var dirtyStr string
	if dirty {
		dirtyStr = " (DIRTY)"
	}

	return fmt.Sprintf("%d%s", v, dirtyStr)
}

func RunMigrations(cfg util.Config) error {
	if err := runNewUpMigrations(cfg); err != nil {
		return err
	}

	StartMVEScheduler(cfg)

	return nil
}

func RunMigrationsSync(cfg util.Config) error {
	if err := runNewUpMigrations(cfg); err != nil {
		return err
	}

	if err := TriggerMVE(cfg); err != nil {
		return err
	}

	return nil
}

func (s *SqlDatabase) runUpMigrations() error {
	m, err := s.openMigration()
	if err != nil {
		return err
	}

	defer func() {
		_, _ = m.Close()
	}()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange && err != io.EOF {
		return err
	}

	return nil
}

func runNewUpMigrations(cfg util.Config) error {
	db, err := NewSqlDatabase(cfg)
	if err != nil {
		return fmt.Errorf("error while Creating Db: %w", err)
	}

	err = db.runUpMigrations()
	if err != nil {
		return fmt.Errorf("error while Migrating Db: %w", err)
	}

	return nil
}

func (s *SqlDatabase) WaitPostMigrations() {
	WaitMVEForFirstRun()
}
