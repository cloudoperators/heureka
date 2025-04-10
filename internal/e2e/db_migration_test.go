// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"embed"
	"io/fs"
	"regexp"

	"github.com/jmoiron/sqlx"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/server"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:embed migrations/*_A.*.sql
var migrationAFiles embed.FS

//go:embed migrations/*_B.*.sql
var migrationBFiles embed.FS

//go:embed migrations/*.sql
var migrationABFiles embed.FS

func setDbTestMigrationFiles(migrationFiles *embed.FS) {
	mariadb.MigrationFiles = *migrationFiles
}

func setDbAMigrationFiles() {
	setDbTestMigrationFiles(&migrationAFiles)
}

func setDbABMigrationFiles() {
	setDbTestMigrationFiles(&migrationABFiles)
}

func extractVersion(filename string) string {
	re := regexp.MustCompile(`^(\d+)_`)
	match := re.FindStringSubmatch(filename)
	Expect(len(match) >= 2).To(BeTrue())
	return match[1]
}

func getAMigrationVersion() string {
	entries, err := fs.ReadDir(migrationAFiles, "migrations")
	Expect(err).To(BeNil())
	Expect(len(entries) >= 1).To(BeTrue())
	return extractVersion(entries[0].Name())
}

func getBMigrationVersion() string {
	entries, err := fs.ReadDir(migrationBFiles, "migrations")
	Expect(err).To(BeNil())
	Expect(len(entries) >= 1).To(BeTrue())
	return extractVersion(entries[0].Name())
}

func dbVersionShouldBeZero(cfg *util.Config) {
	v, err := db.GetVersion()
	Expect(err).To(BeNil())
	Expect(v).To(Equal("0"))
	dbShouldNotContainAMigrationData(cfg)
	dbShouldNotContainBMigrationData(cfg)
}

func dbVersionShouldBeA(cfg *util.Config) {
	v, err := db.GetVersion()
	Expect(err).To(BeNil())
	Expect(v).To(Equal(getAMigrationVersion()))
	dbShouldNotContainBMigrationData(cfg)
}

func dbVersionShouldBeB(db *mariadb.SqlDatabase) {
	v, err := db.GetVersion()
	Expect(err).To(BeNil())
	Expect(v).To(Equal(getBMigrationVersion()))
}

func dbVersionIsZero(cfg *util.Config) {
	dbVersionShouldBeZero(cfg)
	setDbABMigrationFiles()
}

func dbVersionIsA(db *mariadb.SqlDatabase, cfg *util.Config) {
	dbVersionShouldBeZero(cfg)
	setDbAMigrationFiles()
	createHeurekaServer(cfg)
	dbShouldContainAMigrationData(cfg)
	dbVersionShouldBeA(cfg)
	setDbABMigrationFiles()
}

func dbVersionIsB(db *mariadb.SqlDatabase, cfg *util.Config) {
	dbVersionShouldBeZero(cfg)
	setDbABMigrationFiles()
	createHeurekaServer(cfg)
	dbShouldContainAMigrationData(cfg)
	dbShouldContainBMigrationData(cfg)
	dbVersionShouldBeB(db)
}

func createHeurekaServer(cfg *util.Config) {
	server.NewServer(*cfg)
}

func tableExists(db *sqlx.DB, tableName string) bool {
	var exists bool
	query := `
        SELECT COUNT(*) > 0
        FROM information_schema.tables
        WHERE table_schema = DATABASE() AND table_name = ?
    `
	err := db.Get(&exists, query, tableName)
	Expect(err).To(BeNil())
	return exists
}

func dbShouldContainAMigrationData(cfg *util.Config) {
	sx, err := mariadb.Connect(*cfg)
	Expect(err).To(BeNil())
	defer sx.Close()
	exists := tableExists(sx, "A_USER")
	Expect(exists).To(BeTrue())
}

func dbShouldNotContainAMigrationData(cfg *util.Config) {
	sx, err := mariadb.Connect(*cfg)
	Expect(err).To(BeNil())
	defer sx.Close()
	exists := tableExists(sx, "A_USER")
	Expect(exists).To(BeFalse())
}

func dbShouldContainBMigrationData(cfg *util.Config) {
	sx, err := mariadb.Connect(*cfg)
	Expect(err).To(BeNil())
	defer sx.Close()
	exists := tableExists(sx, "B_USER")
	Expect(exists).To(BeTrue())
}

func dbShouldNotContainBMigrationData(cfg *util.Config) {
	sx, err := mariadb.Connect(*cfg)
	Expect(err).To(BeNil())
	defer sx.Close()
	exists := tableExists(sx, "B_USER")
	Expect(exists).To(BeFalse())
}

func dbShouldContainAllMigrations(db *mariadb.SqlDatabase, cfg *util.Config) {
	dbShouldContainAMigrationData(cfg)
	dbShouldContainBMigrationData(cfg)
	dbVersionShouldBeB(db)
}

var _ = Describe("Proceeding migration on heureka startup", Label("e2e", "Migrations"), func() {
	var cfg util.Config

	BeforeEach(func() {
		db = dbm.NewTestSchema()
		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
	})
	When("creating app with zero version of db", func() {
		It("executes all available migrations", func() {
			dbVersionIsZero(&cfg)
			createHeurekaServer(&cfg)
			dbShouldContainAllMigrations(db, &cfg)
		})
	})
	When("creating app with some version of db", func() {
		It("exectues only new versions of migrations", func() {
			dbVersionIsA(db, &cfg)
			createHeurekaServer(&cfg)
			dbShouldContainAllMigrations(db, &cfg)
		})
	})
	When("creating app with newest version of db", func() {
		It("executes no migration", func() {
			dbVersionIsB(db, &cfg)
			createHeurekaServer(&cfg)
			dbShouldContainAllMigrations(db, &cfg)
		})
	})
})
