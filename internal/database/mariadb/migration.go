// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
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
		if serr, derr := m.Close(); serr != nil || derr != nil {
			log.Printf("error during closing migration: source - %s, database - %s", serr, derr)
		}
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
		"mysql", driver)
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
	err := runNewUpMigrations(cfg)
	if err != nil {
		return err
	}
	err = runNewPostMigrationsAsync(cfg)
	if err != nil {
		return err
	}
	err = enableScheduler(cfg)
	if err != nil {
		return fmt.Errorf("error while Enabling Scheduler Db: %w", err)
	}
	return nil
}

func enableScheduler(cfg util.Config) error {
	db, err := GetSqlxRootConnection(cfg)
	if err != nil {
		logrus.WithError(err).Error(err)
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			logrus.WithError(err).Error(err)
		}
	}()

	_, err = db.Exec("SET GLOBAL event_scheduler = ON;")
	if err != nil {
		logrus.WithError(err).Error(err)
		return err
	}
	return nil
}

func runNewPostMigrationsAsync(cfg util.Config) error {
	_, err := runNewPostMigrations(cfg)
	if err != nil {
		return err
	}
	return nil
}

func RunMigrationsSync(cfg util.Config) error {
	err := runNewUpMigrations(cfg)
	if err != nil {
		return err
	}
	err = runNewPostMigrationsSync(cfg)
	if err != nil {
		return err
	}
	return nil
}

func runNewPostMigrationsSync(cfg util.Config) error {
	db, err := runNewPostMigrations(cfg)
	if err != nil {
		return err
	}
	err = db.WaitPostMigrations()
	if err != nil {
		return fmt.Errorf("error while waiting for post migration procedures: %w", err)
	}
	return nil
}

func (s *SqlDatabase) runUpMigrations() error {
	m, err := s.openMigration()
	if err != nil {
		return err
	}
	defer func() {
		if serr, derr := m.Close(); serr != nil || derr != nil {
			log.Printf("error during closing migration: source - %s, database - %s", serr, derr)
		}
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange && err != io.EOF {
		return err
	}

	return nil
}

func runNewPostMigrations(cfg util.Config) (*SqlDatabase, error) {
	db, err := NewSqlDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("error while Creating Db: %w", err)
	}
	err = db.runPostMigrations()
	if err != nil {
		return nil, fmt.Errorf("error while starting Post Migration procedures: %w", err)
	}
	return db, nil
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
		return procs, fmt.Errorf("could not check if table exists: %w", err)
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

func (pmc *postMigrationContext) hasError() bool {
	return len(pmc.errs) > 0
}

func (pmc *postMigrationContext) getError() error {
	return fmt.Errorf("error when exeute joined callers: [%s]", strings.Join(pmc.errs, "; "))
}

func (s *SqlDatabase) runPostMigrations() error {
	procs, err := s.getPostMigrationProcedures()
	if err != nil {
		return fmt.Errorf("failed to get post migration procedures: %w", err)
	}

	if err := s.checkProceduresExist(procs); err != nil {
		return err
	}
	s.runPostMigrationProcessInBackground(procs)
	return nil
}

func (s *SqlDatabase) checkProceduresExist(procs []string) error {
	exists, err := s.proceduresExist(procs)
	if err != nil {
		return fmt.Errorf("could not check procedures exist: %w", err)
	} else if !exists {
		return fmt.Errorf("some procedures [%s] do not exist", strings.Join(procs, ", "))
	}
	return nil
}

func (s *SqlDatabase) proceduresExist(procedures []string) (bool, error) {
	if len(procedures) == 0 {
		return true, nil
	}

	placeholders := make([]string, len(procedures))
	args := make([]interface{}, len(procedures))
	for i, p := range procedures {
		placeholders[i] = "?"
		args[i] = p
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM information_schema.routines
		WHERE routine_schema = DATABASE()
		  AND routine_type = 'PROCEDURE'
		  AND routine_name IN (%s);
	`, strings.Join(placeholders, ","))

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("could not check if procedures exist: %w", err)
	}

	if count == len(procedures) {
		return true, nil
	}
	return false, nil
}

func (s *SqlDatabase) runPostMigrationProcessInBackground(procs []string) {
	s.runPostMigrationProceduresInBackground(procs)
	s.runPostMigrationCleanupRoutineInBackground(procs)
}

func (s *SqlDatabase) runPostMigrationProceduresInBackground(procs []string) {
	for _, p := range procs {
		s.postMigrationCtx.wg.Add(1)
		go func() {
			defer s.postMigrationCtx.wg.Done()
			if _, err := s.db.Exec(fmt.Sprintf("CALL %s();", p)); err != nil {
				s.postMigrationCtx.appendErrorMessage(fmt.Sprintf("%s: %v", p, err))
			}
		}()
	}
}

func (s *SqlDatabase) runPostMigrationCleanupRoutineInBackground(procs []string) {
	go func() {
		if err := s.WaitPostMigrations(); err != nil {
			logrus.WithError(err).Error(err)
		}

		if err := s.CloseConnection(); err != nil {
			logrus.Warnf("error during closing connection: %s", err)
		}
	}()
}

func (s *SqlDatabase) WaitPostMigrations() error {
	s.postMigrationCtx.wg.Wait()
	if s.postMigrationCtx.hasError() {
		return s.postMigrationCtx.getError()
	}
	return nil
}
