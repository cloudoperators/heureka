package mariadb

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4/source"
)

//go:embed migrations/*.sql
var MigrationFiles embed.FS

func loadMigrations() (map[uint]string, error) {
	migs := make(map[uint]string)

	// Pattern: e.g., "1_init.up.sql"
	re := regexp.MustCompile(`^(\d+)_.*\.up\.sql$`)

	err := fs.WalkDir(MigrationFiles, "migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		name := filepath.Base(path)
		match := re.FindStringSubmatch(name)
		if match == nil {
			return nil // skip non-up files
		}

		version, err := strconv.ParseUint(match[1], 10, 64)
		if err != nil {
			return err
		}

		data, err := MigrationFiles.ReadFile(path)
		if err != nil {
			return err
		}

		sql := string(data)

		migs[uint(version)] = sql
		return nil
	})

	return migs, err
}

type stringMigrationSource struct {
	migrations map[uint]string
}

func (s *stringMigrationSource) Open(url string) (source.Driver, error) {
	return s, nil
}

func (s *stringMigrationSource) Close() error { return nil }

func (s *stringMigrationSource) First() (uint, error) {
	for v := range s.migrations {
		return v, nil
	}
	return 0, fmt.Errorf("no migrations")
}

func (s *stringMigrationSource) Prev(version uint) (uint, error) {
	return 0, fmt.Errorf("not implemented")
}

func (s *stringMigrationSource) Next(version uint) (uint, error) {
	for v := range s.migrations {
		if v > version {
			return v, nil
		}
	}
	return 0, io.EOF
}

func (s *stringMigrationSource) ReadUp(version uint) (io.ReadCloser, string, error) {
	if sql, ok := s.migrations[version]; ok {
		return io.NopCloser(strings.NewReader(sql)), "text/sql", nil
	}
	return nil, "", fmt.Errorf("no such version")
}

func (s *stringMigrationSource) ReadDown(version uint) (io.ReadCloser, string, error) {
	return nil, "", fmt.Errorf("down migration not supported")
}

func (s *stringMigrationSource) List() ([]*source.Migration, error) {
	var list []*source.Migration
	for v := range s.migrations {
		list = append(list, &source.Migration{
			Version:    v,
			Identifier: fmt.Sprintf("mem_%d", v),
		})
	}
	return list, nil
}

func getLowestVersionOfMigration(migs map[uint]string) (uint, bool) {
	if len(migs) == 0 {
		return 0, false // no versions available
	}

	var min uint = ^uint(0) // max uint value
	for version := range migs {
		if version < min {
			min = version
		}
	}
	return min, true
}
