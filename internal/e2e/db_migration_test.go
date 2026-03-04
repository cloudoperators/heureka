// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"testing/fstest"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/server"
	"github.com/cloudoperators/heureka/internal/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:embed migrations/*_A.*.sql
var migrationAFiles embed.FS

//go:embed migrations/*_B.*.sql
var migrationBFiles embed.FS

//go:embed migrations/*_A.*.sql migrations/*_B.*.sql
var migrationABFiles embed.FS

//go:embed migrations/*_mvTestTable.*.sql
var migrationMvTestTableMigrationFiles embed.FS

//go:embed migrations/*_add_post_migration.*.sql
var migrationAddPostMigrationMigrationFiles embed.FS

// Merge multiple fs.FS into one MapFS with "migrations/" prefix
func MergeToMapFS(sources ...fs.FS) (fstest.MapFS, error) {
	merged := fstest.MapFS{}

	for _, source := range sources {
		err := fs.WalkDir(source, ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			data, err := fs.ReadFile(source, p)
			if err != nil {
				return err
			}

			// Ensure everything goes under "migrations/"
			key := path.Join("migrations", path.Base(p))
			merged[key] = &fstest.MapFile{Data: data}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return merged, nil
}

func setDbTestMigration(migrationFiles fs.FS) {
	mariadb.MigrationFs = migrationFiles
}

func setDbAMigration() {
	setDbTestMigration(&migrationAFiles)
}

func setDbABMigration() {
	setDbTestMigration(&migrationABFiles)
}

func addDbMvTestTableMigration() {
	mapFS, err := MergeToMapFS(mariadb.MigrationFs, migrationMvTestTableMigrationFiles)
	Expect(err).To(BeNil())
	setDbTestMigration(&mapFS)
}

func addDbMvTestTableAndAddPostMigrationMigration() {
	mapFS, err := MergeToMapFS(mariadb.MigrationFs, migrationMvTestTableMigrationFiles, migrationAddPostMigrationMigrationFiles)
	Expect(err).To(BeNil())
	setDbTestMigration(&mapFS)
}

func extractVersion(filename string) string {
	re := regexp.MustCompile(`^(\d+)_`)
	match := re.FindStringSubmatch(filename)
	Expect(len(match) >= 2).To(BeTrue())
	return match[1]
}

func getFirstMigrationVersionFromFiles(files *embed.FS) string {
	entries, err := fs.ReadDir(files, "migrations")
	Expect(err).To(BeNil())
	Expect(len(entries) >= 1).To(BeTrue())
	return extractVersion(entries[0].Name())
}

func getAMigrationVersion() string {
	return getFirstMigrationVersionFromFiles(&migrationAFiles)
}

func getBMigrationVersion() string {
	return getFirstMigrationVersionFromFiles(&migrationBFiles)
}

func getMvTestTableMigrationVersion() string {
	return getFirstMigrationVersionFromFiles(&migrationMvTestTableMigrationFiles)
}

func (dbmt *DbMigrationTest) tableExists(tableName string) bool {
	db := dbmt.dbConnect()
	defer db.Close()

	var exists bool
	query := `
        SELECT COUNT(*) > 0
        FROM information_schema.tables
        WHERE table_schema = DATABASE() AND table_name = ?
    `
	err := db.Get(&exists, query, tableName)
	Expect(err).To(BeNil())

	rows, err := db.Queryx(`SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE()`)
	for rows.Next() {
		var name string
		_ = rows.Scan(&name)
	}

	return exists
}

func (dbmt *DbMigrationTest) countRows(tableName string) int {
	db := dbmt.dbConnect()
	defer db.Close()

	query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	var count int
	err := db.Get(&count, query)
	Expect(err).To(BeNil())
	return count
}

type DbMigrationTest struct {
	cfg              util.Config
	db               *mariadb.SqlDatabase
	heurekaMigration fs.FS
}

func (dbmt *DbMigrationTest) setup() {
	dbmt.heurekaMigration = mariadb.MigrationFs
	dbmt.db = dbm.NewTestSchemaWithoutMigration()
	dbmt.cfg = dbm.DbConfig()
	dbmt.cfg.Port = e2e_common.GetRandomFreePort()
}

func (dbmt *DbMigrationTest) teardown() {
	dbm.TestTearDown(dbmt.db)
	mariadb.MigrationFs = dbmt.heurekaMigration
}

func (dbmt *DbMigrationTest) dbVersionIsZero() {
	dbmt.dbVersionShouldBeZero()
	setDbABMigration()
}

func (dbmt *DbMigrationTest) dbVersionIsA() {
	dbmt.dbVersionShouldBeZero()
	setDbAMigration()
	dbmt.createHeurekaServer()
	dbmt.dbShouldContainAMigrationTable()
	dbmt.dbVersionShouldBeA()
	setDbABMigration()
}

func (dbmt *DbMigrationTest) dbVersionIsB() {
	dbmt.dbVersionShouldBeZero()
	setDbABMigration()
	dbmt.createHeurekaServer()
	dbmt.dbShouldContainAMigrationTable()
	dbmt.dbShouldContainBMigrationTable()
	dbmt.dbVersionShouldBeB()
}

func (dbmt *DbMigrationTest) dbVersionIsMvTestTable() {
	dbmt.dbVersionShouldBeZero()
	dbmt.dbShouldNotContainPostMigrationProcedureRefreshTable()
	addDbMvTestTableMigration()
	dbmt.createHeurekaServer()
	dbmt.dbShouldNotContainPostMigrationProcedureData()
	dbmt.dbVersionShouldBeMvTestTable()
	addDbMvTestTableAndAddPostMigrationMigration()
}

func (dbmt *DbMigrationTest) createHeurekaServer() {
	server.NewServer(dbmt.cfg)
}

func (dbmt *DbMigrationTest) dbShouldContainABMigrations() {
	dbmt.dbShouldContainAMigrationTable()
	dbmt.dbShouldContainBMigrationTable()
	dbmt.dbVersionShouldBeB()
}

func (dbmt *DbMigrationTest) dbShouldContainPostMigrationProcedureData() {
	dbmt.dbExpectRows("mvTestData", 1)
}

func (dbmt *DbMigrationTest) dbShouldNotContainPostMigrationProcedureData() {
	dbmt.dbExpectRows("mvTestData", 0)
}

func (dbmt *DbMigrationTest) dbExpectRows(tablename string, rows int) {
	Expect(dbmt.countRows(tablename)).To(Equal(rows))
}

func (dbmt *DbMigrationTest) dbConnect() *sqlx.DB {
	sx, err := mariadb.Connect(dbmt.cfg)
	Expect(err).To(BeNil())
	return sx
}

func (dbmt *DbMigrationTest) dbVersionShouldBeZero() {
	v, err := mariadb.GetVersion(dbmt.cfg)
	Expect(err).To(BeNil())
	Expect(v).To(Equal("0"))
	dbmt.dbShouldNotContainAMigrationTable()
	dbmt.dbShouldNotContainBMigrationTable()
}

func (dbmt *DbMigrationTest) dbVersionShouldBeA() {
	v, err := mariadb.GetVersion(dbmt.cfg)
	Expect(err).To(BeNil())
	Expect(v).To(Equal(getAMigrationVersion()))
	dbmt.dbShouldNotContainBMigrationTable()
}

func (dbmt *DbMigrationTest) dbVersionShouldBeB() {
	v, err := mariadb.GetVersion(dbmt.cfg)
	Expect(err).To(BeNil())
	Expect(v).To(Equal(getBMigrationVersion()))
}

func (dbmt *DbMigrationTest) dbVersionShouldBeMvTestTable() {
	v, err := mariadb.GetVersion(dbmt.cfg)
	Expect(err).To(BeNil())
	Expect(v).To(Equal(getMvTestTableMigrationVersion()))
}

func (dbmt *DbMigrationTest) dbShouldContainAMigrationTable() {
	dbmt.dbShouldContainTable("A_USER")
}

func (dbmt *DbMigrationTest) dbShouldNotContainAMigrationTable() {
	dbmt.dbShouldNotContainTable("A_USER")
}

func (dbmt *DbMigrationTest) dbShouldContainBMigrationTable() {
	dbmt.dbShouldContainTable("B_USER")
}

func (dbmt *DbMigrationTest) dbShouldNotContainBMigrationTable() {
	dbmt.dbShouldNotContainTable("B_USER")
}

func (dbmt *DbMigrationTest) dbShouldNotContainPostMigrationProcedureRefreshTable() {
	dbmt.dbShouldNotContainTable("mvTestData")
}

func (dbmt *DbMigrationTest) dbShouldContainTable(tablename string) {
	Expect(dbmt.tableExists(tablename)).To(BeTrue())
}

func (dbmt *DbMigrationTest) dbShouldNotContainTable(tablename string) {
	Expect(dbmt.tableExists(tablename)).To(BeFalse())
}

func waitForPostMigration() {
	time.Sleep(100 * time.Millisecond)
}

var _ = Describe("Proceeding migration on heureka startup", Label("e2e", "Migrations"), func() {
	var migrationTest DbMigrationTest
	BeforeEach(func() {
		migrationTest.setup()
	})
	AfterEach(func() {
		migrationTest.teardown()
	})
	When("creating app with zero version of db", func() {
		It("executes all available migrations", func() {
			migrationTest.dbVersionIsZero()
			migrationTest.createHeurekaServer()
			migrationTest.dbShouldContainABMigrations()
		})
	})
	When("creating app with some version of db", func() {
		It("exectues only new versions of migrations", func() {
			migrationTest.dbVersionIsA()
			migrationTest.createHeurekaServer()
			migrationTest.dbShouldContainABMigrations()
		})
	})
	When("creating app with newest version of db", func() {
		It("executes no migration", func() {
			migrationTest.dbVersionIsB()
			migrationTest.createHeurekaServer()
			migrationTest.dbShouldContainABMigrations()
		})
	})
})

var _ = Describe("Proceeding migration on heureka startup", Label("e2e", "Migrations"), Label("e2e", "PostMigration"), func() {
	var migrationTest DbMigrationTest
	BeforeEach(func() {
		migrationTest.setup()
	})
	AfterEach(func() {
		migrationTest.teardown()
	})
	When("creating app with mvTestTable migration applied and append procedure to post migration migration to be applied", func() {
		It("executes post migration procedure after successful migration", func() {
			migrationTest.dbVersionIsMvTestTable()
			migrationTest.createHeurekaServer()
			waitForPostMigration()
			migrationTest.dbShouldContainPostMigrationProcedureData()
		})
	})
})
