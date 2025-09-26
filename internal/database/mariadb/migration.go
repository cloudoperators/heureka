// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"embed"
	"io/fs"

	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/cloudoperators/heureka/internal/util"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/sirupsen/logrus"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS
var Migration fs.FS = &migrationFiles

func (s *SqlDatabase) GetVersion() (string, error) {
	m, err := s.openMigration()
	if err != nil {
		return "", fmt.Errorf("Could not open migration without source: %w", err)
	}
	defer m.Close()

	v, d, err := getMigrationVersion(m)
	if err != nil {
		return "", fmt.Errorf("Could not get migration version: %w", err)
	}

	return versionToString(v, d), nil
}

func GetVersion(cfg util.Config) (string, error) {
	db, err := NewSqlDatabase(cfg)
	if err != nil {
		return "", fmt.Errorf("Error while Creating Db")
	}

	v, err := db.GetVersion()
	if err != nil {
		return "", fmt.Errorf("Error while checking Db migration: %w", err)
	}
	return v, nil
}

func (s *SqlDatabase) RunUpMigrations() error {
	m, err := s.openMigration()
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange && err != io.EOF {
		return err
	}

	return nil
}

func versionToString(v uint, dirty bool) string {
	var dirtyStr string
	if dirty {
		dirtyStr = " (DIRTY)"
	}
	return fmt.Sprintf("%d%s", v, dirtyStr)
}

func getMigrationVersion(m *migrate.Migrate) (uint, bool, error) {
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, err
	}
	return version, dirty, nil
}

func (s *SqlDatabase) openMigration() (*migrate.Migrate, error) {
	err := s.connectDB()
	if err != nil {
		return nil, fmt.Errorf("Could not connect DB: %w", err)
	}

	d, err := iofs.New(Migration, "migrations")
	if err != nil {
		return nil, err
	}

	driver, err := mysql.WithInstance(s.db.GetDbInstance(), &mysql.Config{DatabaseName: s.dbName})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance(
		"iofs", d,
		"mysql", driver)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func RunMigrations(cfg util.Config) error {
	db, err := NewSqlDatabase(cfg)
	if err != nil {
		return fmt.Errorf("Error while Creating Db: %w", err)
	}
	err = db.RunUpMigrations()
	if err != nil {
		return fmt.Errorf("Error while Migrating Db: %w", err)
	}

	db, err = NewSqlDatabase(cfg)
	if err != nil {
		return fmt.Errorf("Error while Creating Db: %w", err)
	}
	err = db.RunPostMigrations()
	if err != nil {
		return fmt.Errorf("Error while starting Post Migration procedures: %w", err)
	}

	err = EnableScheduler(cfg)
	if err != nil {
		return fmt.Errorf("Error while Enabling Scheduler Db: %w", err)
	}
	return nil
}

func (s *SqlDatabase) procedureExists(procedure string) (bool, error) {
	var count int
	err := s.db.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*)
		FROM information_schema.routines
		WHERE routine_schema = DATABASE()
		  AND routine_type = 'PROCEDURE'
		  AND routine_name = '%s';
	`, procedure)).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("Could not check if procedure exists '%s', %w", procedure, err)
	}

	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (s *SqlDatabase) tableExists(table string) (bool, error) {
	var count int
	err := s.db.Get(&count, `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
		  AND table_name = ?
	`, table)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

const PostMigrationProcedureRegistryTable = "post_migration_procedure_registry"

func (s *SqlDatabase) getPostMigrationProcedures() ([]string, error) {
	var procs []string

	exists, err := s.tableExists(PostMigrationProcedureRegistryTable)
	if err != nil {
		return procs, fmt.Errorf("Could not check if table exists: %w", err)
	} else if !exists {
		return procs, nil
	}

	err = s.db.Select(&procs, fmt.Sprintf("SELECT name FROM %s", PostMigrationProcedureRegistryTable))
	if err != nil {
		return nil, err
	}
	return procs, nil
}

type postMigrationContext struct {
	wg   sync.WaitGroup
	mu   sync.Mutex
	errs []string
}

func (pmc *postMigrationContext) appendErrorMessage(msg string) {
	pmc.mu.Lock()
	pmc.errs = append(pmc.errs, msg)
	pmc.mu.Unlock()
}

func (pmc postMigrationContext) hasError() bool {
	return len(pmc.errs) > 0
}

func (pmc postMigrationContext) getError() error {
	return fmt.Errorf("Error when exeute joined callers: [%s]", strings.Join(pmc.errs, "; "))
}

func (s *SqlDatabase) callPostMigrationProcedure(proc string, pmCtx *postMigrationContext) error {
	exists, err := s.procedureExists(proc)
	if err != nil {
		return fmt.Errorf("Could not check caller procedure exists: %w", err)
	} else if !exists {
		return fmt.Errorf("Caller procedure '%s' does not exist", proc)
	}

	pmCtx.wg.Add(1)
	go func() {
		defer pmCtx.wg.Done()
		if _, err := s.db.Exec(fmt.Sprintf("CALL %s();", proc)); err != nil {
			pmCtx.appendErrorMessage(fmt.Sprintf("%s: %v", proc, err))
		}
	}()

	return nil
}

func (s *SqlDatabase) RunPostMigrations() error {
	procs, err := s.getPostMigrationProcedures()
	if err != nil {
		return fmt.Errorf("failed to get post migration procedures: %w", err)
	}
	for _, p := range procs {
		if err := s.callPostMigrationProcedure(p, &s.postMigrationCtx); err != nil {
			return fmt.Errorf("Failed to call post migration procedure: %w", err)
		}
	}
	go func() {
		s.postMigrationCtx.wg.Wait()
		s.CloseConnection()
	}()
	return nil
}

func (s *SqlDatabase) WaitPostMigrations() error {
	s.postMigrationCtx.wg.Wait()
	if s.postMigrationCtx.hasError() {
		return s.postMigrationCtx.getError()
	}
	return nil
}

func EnableScheduler(cfg util.Config) error {
	db, err := getSqlxRootConnection(cfg)
	if err != nil {
		logrus.WithError(err).Error(err)
		return err
	}
	defer db.Close()

	_, err = db.Exec("SET GLOBAL event_scheduler = ON;")
	if err != nil {
		logrus.WithError(err).Error(err)
		return err
	}
	return nil
}
