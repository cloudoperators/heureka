// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/goark/go-cvss/v3/metric"
	"github.com/onsi/ginkgo/v2/dsl/core"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type SeedCollection struct {
	IssueVariantRows           []mariadb.IssueVariantRow
	IssueRepositoryRows        []mariadb.BaseIssueRepositoryRow
	UserRows                   []mariadb.UserRow
	IssueRows                  []mariadb.IssueRow
	IssueMatchRows             []mariadb.IssueMatchRow
	ActivityRows               []mariadb.ActivityRow
	EvidenceRows               []mariadb.EvidenceRow
	ComponentInstanceRows      []mariadb.ComponentInstanceRow
	ComponentVersionRows       []mariadb.ComponentVersionRow
	ComponentRows              []mariadb.ComponentRow
	ServiceRows                []mariadb.BaseServiceRow
	SupportGroupUserRows       []mariadb.SupportGroupUserRow
	SupportGroupRows           []mariadb.SupportGroupRow
	SupportGroupServiceRows    []mariadb.SupportGroupServiceRow
	OwnerRows                  []mariadb.OwnerRow
	ActivityHasServiceRows     []mariadb.ActivityHasServiceRow
	ActivityHasIssueRows       []mariadb.ActivityHasIssueRow
	ComponentVersionIssueRows  []mariadb.ComponentVersionIssueRow
	IssueMatchEvidenceRows     []mariadb.IssueMatchEvidenceRow
	IssueMatchChangeRows       []mariadb.IssueMatchChangeRow
	IssueRepositoryServiceRows []mariadb.IssueRepositoryServiceRow
}

func (s *SeedCollection) GetComponentInstanceById(id int64) *mariadb.ComponentInstanceRow {
	for _, ci := range s.ComponentInstanceRows {
		if ci.Id.Int64 == id {
			return &ci
		}
	}
	return nil
}

func (s *SeedCollection) GetIssueById(id int64) *mariadb.IssueRow {
	for _, issue := range s.IssueRows {
		if issue.Id.Int64 == id {
			return &issue
		}
	}
	return nil
}

func (s *SeedCollection) GetIssueVariantsByIssueId(id int64) []mariadb.IssueVariantRow {
	var r []mariadb.IssueVariantRow
	for _, iv := range s.IssueVariantRows {
		if iv.IssueId.Int64 == id {
			r = append(r, iv)
		}
	}
	return r
}

func (s *SeedCollection) GetIssueMatchesByServiceOwner(owner mariadb.OwnerRow) []mariadb.IssueMatchRow {
	serviceIds := lo.FilterMap(s.OwnerRows, func(o mariadb.OwnerRow, _ int) (int64, bool) {
		return o.ServiceId.Int64, o.UserId.Int64 == owner.UserId.Int64
	})

	ciIds := lo.FilterMap(s.ComponentInstanceRows, func(c mariadb.ComponentInstanceRow, _ int) (int64, bool) {
		return c.Id.Int64, lo.Contains(serviceIds, c.ServiceId.Int64)
	})

	return lo.Filter(s.IssueMatchRows, func(im mariadb.IssueMatchRow, _ int) bool {
		return lo.Contains(ciIds, im.ComponentInstanceId.Int64)

	})
}

func (s *SeedCollection) GetIssueVariantsByService(service *mariadb.BaseServiceRow) []mariadb.IssueVariantRow {
	var issueVariants []mariadb.IssueVariantRow
	for _, irs := range s.IssueRepositoryServiceRows {
		if irs.ServiceId.Valid && service.Id.Valid && irs.ServiceId.Int64 == service.Id.Int64 {
			for _, iv := range s.IssueVariantRows {
				if iv.IssueRepositoryId.Valid && irs.IssueRepositoryId.Int64 == iv.IssueRepositoryId.Int64 {
					issueVariants = append(issueVariants, iv)
				}
			}
		}
	}
	issueVariants = lo.UniqBy(issueVariants, func(a mariadb.IssueVariantRow) int64 { return a.Id.Int64 })
	return issueVariants
}

func (s *SeedCollection) GetIssueVariantsByIssueMatch(im *mariadb.IssueMatchRow) []mariadb.IssueVariantRow {
	var issueVariants []mariadb.IssueVariantRow
	for _, i := range s.IssueRows {
		if im.IssueId.Int64 == i.Id.Int64 {
			for _, iv := range s.IssueVariantRows {
				if iv.IssueId.Int64 == i.Id.Int64 {
					issueVariants = append(issueVariants, iv)
				}
			}
		}
	}
	issueVariants = lo.UniqBy(issueVariants, func(iv mariadb.IssueVariantRow) int64 { return iv.Id.Int64 })
	return issueVariants
}

func (s *SeedCollection) GetIssueByService(service *mariadb.BaseServiceRow) []mariadb.IssueRow {
	var issues []mariadb.IssueRow
	for _, ci := range s.ComponentInstanceRows {
		if ci.ServiceId.Valid && service.Id.Valid && ci.ServiceId.Int64 == service.Id.Int64 {
			for _, im := range s.IssueMatchRows {
				if im.ComponentInstanceId.Valid && ci.Id.Valid && im.ComponentInstanceId.Int64 == ci.Id.Int64 {
					for _, i := range s.IssueRows {
						if i.Id.Valid && im.IssueId.Valid && i.Id.Int64 == im.IssueId.Int64 {
							issues = append(issues, i)
						}
					}
				}
			}
		}
	}

	issues = lo.UniqBy(issues, func(i mariadb.IssueRow) int64 { return i.Id.Int64 })
	return issues
}

func (s *SeedCollection) GetComponentInstanceByIssueMatches(im []mariadb.IssueMatchRow) ([]mariadb.ComponentInstanceRow, []*int64) {
	ids := make([]*int64, len(im))
	var expectedComponentInstances []mariadb.ComponentInstanceRow
	for i, row := range im {
		x := row.ComponentInstanceId.Int64
		ci, found := lo.Find(s.ComponentInstanceRows, func(cir mariadb.ComponentInstanceRow) bool {
			return cir.Id.Int64 == x
		})
		if found && lo.ContainsBy(expectedComponentInstances, func(c mariadb.ComponentInstanceRow) bool {
			return c.Id.Int64 == ci.Id.Int64
		}) {
			expectedComponentInstances = append(expectedComponentInstances, ci)
			ids[i] = &x
		}
	}
	return expectedComponentInstances, ids
}

func (s *SeedCollection) GetComponentInstance() mariadb.ComponentInstanceRow {
	return s.ComponentInstanceRows[0]
}

func (s *SeedCollection) GetComponentInstanceWithPredicateVal(predicate func(picked, iter mariadb.ComponentInstanceRow) (string, bool)) (mariadb.ComponentInstanceRow, []string) {
	picked := s.ComponentInstanceRows[0]
	return picked, lo.FilterMap(
		s.ComponentInstanceRows,
		func(iter mariadb.ComponentInstanceRow, _ int) (string, bool) {
			return predicate(picked, iter)
		},
	)
}

func (s *SeedCollection) GetComponentInstanceVal(predicate func(cir mariadb.ComponentInstanceRow) string) []string {
	return lo.Map(s.ComponentInstanceRows, func(cir mariadb.ComponentInstanceRow, _ int) string {
		return predicate(cir)
	})
}

func (s *SeedCollection) GetValidComponentInstanceRows() []mariadb.ComponentInstanceRow {
	var valid []mariadb.ComponentInstanceRow
	var added []int64
	for _, ci := range s.ComponentInstanceRows {
		if ci.Id.Valid && !lo.Contains(added, ci.Id.Int64) {
			added = append(added, ci.Id.Int64)
			valid = append(valid, ci)
		}
	}
	return valid
}

func (s *SeedCollection) GetValidIssueMatchRows() []mariadb.IssueMatchRow {
	var valid []mariadb.IssueMatchRow
	var added []int64
	for _, r := range s.IssueMatchRows {
		if r.Id.Valid && !lo.Contains(added, r.Id.Int64) {
			added = append(added, r.Id.Int64)
			valid = append(valid, r)
		}
	}
	return valid
}

func (s *SeedCollection) GetValidEvidenceRows() []mariadb.EvidenceRow {
	var valid []mariadb.EvidenceRow
	var added []int64
	for _, e := range s.EvidenceRows {
		if e.Id.Valid && !lo.Contains(added, e.Id.Int64) {
			added = append(added, e.Id.Int64)
			valid = append(valid, e)
		}
	}
	return valid
}

type DatabaseSeeder struct {
	db *sqlx.DB
}

func NewDatabaseSeeder(cfg util.Config) (*DatabaseSeeder, error) {
	db, err := mariadb.Connect(cfg)

	if err != nil {
		return nil, err
	}

	return &DatabaseSeeder{
		db: db,
	}, nil

}

// Generate a random CVSS 3.1 vector
func GenerateRandomCVSS31Vector() string {
	avValues := []string{"N", "A", "L", "P"}
	acValues := []string{"L", "H"}
	prValues := []string{"N", "L", "H"}
	uiValues := []string{"N", "R"}
	sValues := []string{"U", "C"}
	cValues := []string{"N", "L", "H"}
	iValues := []string{"N", "L", "H"}
	aValues := []string{"N", "L", "H"}

	emValues := []string{"X", "U", "P", "F", "H"}
	rcValues := []string{"X", "U", "R", "C"}
	rlValues := []string{"X", "U", "O", "T", "W"}

	crValues := []string{"X", "L", "M", "H"}
	irValues := []string{"X", "L", "M", "H"}
	arValues := []string{"X", "L", "M", "H"}

	maValues := []string{"X", "N", "L", "H"}
	mavValues := []string{"X", "N", "A", "L", "P"}
	macValues := []string{"X", "L", "H"}
	mprValues := []string{"X", "N", "L", "H"}
	muiValues := []string{"X", "N", "R"}
	msValues := []string{"X", "U", "C"}

	mcValues := []string{"X", "N", "L", "H"}
	miValues := []string{"X", "N", "L", "H"}

	var baseVector []string

	//version
	baseVector = append(baseVector, "CVSS:3.1")

	// Attack Vector (AV)
	baseVector = append(baseVector, "AV:"+avValues[rand.Intn(len(avValues))])
	// Attack Complexity (AC)
	baseVector = append(baseVector, "AC:"+acValues[rand.Intn(len(acValues))])
	// Privileges Required (PR)
	baseVector = append(baseVector, "PR:"+prValues[rand.Intn(len(prValues))])
	// User Interaction (UI)
	baseVector = append(baseVector, "UI:"+uiValues[rand.Intn(len(uiValues))])
	// Scope (S)
	baseVector = append(baseVector, "S:"+sValues[rand.Intn(len(sValues))])
	// Confidentiality (C)
	baseVector = append(baseVector, "C:"+cValues[rand.Intn(len(cValues))])
	// Integrity (I)
	baseVector = append(baseVector, "I:"+iValues[rand.Intn(len(iValues))])
	// Availability (A)
	baseVector = append(baseVector, "A:"+aValues[rand.Intn(len(aValues))])
	var temporalVector []string
	// Temporal (E)
	temporalVector = append(temporalVector, "E:"+emValues[rand.Intn(len(emValues))])
	// Temporal (RL)
	temporalVector = append(temporalVector, "RL:"+rlValues[rand.Intn(len(rlValues))])
	// Temporal (RC)
	temporalVector = append(temporalVector, "RC:"+rcValues[rand.Intn(len(rcValues))])
	var environmentalVector []string
	// Environmental (CR)
	environmentalVector = append(environmentalVector, "CR:"+crValues[rand.Intn(len(crValues))])
	// Environmental (IR)
	environmentalVector = append(environmentalVector, "IR:"+irValues[rand.Intn(len(irValues))])
	// Environmental (AR)
	environmentalVector = append(environmentalVector, "AR:"+arValues[rand.Intn(len(arValues))])
	// Environmental (MA)
	environmentalVector = append(environmentalVector, "MA:"+maValues[rand.Intn(len(maValues))])
	// Environmental (MAV)
	environmentalVector = append(environmentalVector, "MAV:"+mavValues[rand.Intn(len(mavValues))])
	// Environmental (MAC)
	environmentalVector = append(environmentalVector, "MAC:"+macValues[rand.Intn(len(macValues))])
	// Environmental (MPR)
	environmentalVector = append(environmentalVector, "MPR:"+mprValues[rand.Intn(len(mprValues))])
	// Environmental (MUI)
	environmentalVector = append(environmentalVector, "MUI:"+muiValues[rand.Intn(len(muiValues))])
	// Environmental (MS)
	environmentalVector = append(environmentalVector, "MS:"+msValues[rand.Intn(len(msValues))])
	// Environmental (MC)
	environmentalVector = append(environmentalVector, "MC:"+mcValues[rand.Intn(len(mcValues))])
	// Environmental (MI)
	environmentalVector = append(environmentalVector, "MI:"+miValues[rand.Intn(len(miValues))])

	// Randomly append temporalVector and environmentalVector to baseVector per upper vector
	if rand.Intn(2) == 0 {
		baseVector = append(baseVector, temporalVector...)
	}
	if rand.Intn(2) == 0 {
		baseVector = append(baseVector, environmentalVector...)
	}
	url := strings.Join(baseVector, "/")
	return url
}

func (s *DatabaseSeeder) SeedDbForServer(n int) *SeedCollection {
	users := s.SeedUsers(n)
	supportGroupsMap := s.SeedRealSupportGroups()
	supportGroups := lo.Values[string, mariadb.SupportGroupRow](supportGroupsMap)
	servicesMap := s.SeedRealServices()
	services := lo.Values[string, mariadb.BaseServiceRow](servicesMap)
	components := s.SeedComponents(n)
	componentVersions := s.SeedComponentVersions(n, components)
	componentInstances := s.SeedComponentInstances(n, componentVersions, services)
	repos := s.SeedIssueRepositories()
	issues := s.SeedIssues(n)
	issueVariants := s.SeedIssueVariants(n, repos, issues)
	owners := s.SeedOwners(n, services, users)
	supportGroupServices := s.SeedRealSupportGroupService(servicesMap, supportGroupsMap)
	supportGroupUsers := s.SeedSupportGroupUsers(n, users, supportGroups)
	activities := s.SeedActivities(n)
	activityHasServices := s.SeedActivityHasServices(n, activities, services)
	activityHasIssues := s.SeedActivityHasIssues(n, activities, issues)
	evidences := s.SeedEvidences(n, activities, users)
	componentVersionIssues := s.SeedComponentVersionIssues(n, componentVersions, issues)
	issueMatches := s.SeedIssueMatches(n, issues, componentInstances, users)
	issueMatchEvidences := s.SeedIssueMatchEvidence(n, issueMatches, evidences)
	issueRepositoryServices := s.SeedIssueRepositoryServices(n, services, repos)
	issueMatchChanges := s.SeedIssueMatchChanges(n, issueMatches, activities)

	return &SeedCollection{
		IssueVariantRows:           issueVariants,
		IssueRepositoryRows:        repos,
		UserRows:                   users,
		IssueRows:                  issues,
		IssueMatchRows:             issueMatches,
		ActivityRows:               activities,
		EvidenceRows:               evidences,
		ComponentInstanceRows:      componentInstances,
		ComponentVersionRows:       componentVersions,
		ComponentRows:              components,
		ServiceRows:                services,
		SupportGroupUserRows:       supportGroupUsers,
		SupportGroupRows:           supportGroups,
		SupportGroupServiceRows:    supportGroupServices,
		OwnerRows:                  owners,
		ActivityHasServiceRows:     activityHasServices,
		ActivityHasIssueRows:       activityHasIssues,
		ComponentVersionIssueRows:  componentVersionIssues,
		IssueMatchEvidenceRows:     issueMatchEvidences,
		IssueRepositoryServiceRows: issueRepositoryServices,
		IssueMatchChangeRows:       issueMatchChanges,
	}
}

func (s *DatabaseSeeder) SeedDbWithNFakeData(n int) *SeedCollection {
	users := s.SeedUsers(n)
	supportGroups := s.SeedSupportGroups(n)
	services := s.SeedServices(n)
	components := s.SeedComponents(n)
	componentVersions := s.SeedComponentVersions(n, components)
	componentInstances := s.SeedComponentInstances(n, componentVersions, services)
	repos := s.SeedIssueRepositories()
	issues := s.SeedIssues(n)
	issueVariants := s.SeedIssueVariants(n, repos, issues)
	owners := s.SeedOwners(n, services, users)
	supportGroupServices := s.SeedSupportGroupServices(n/2, services, supportGroups)
	supportGroupUsers := s.SeedSupportGroupUsers(n/2, users, supportGroups)
	activities := s.SeedActivities(n)
	activityHasServices := s.SeedActivityHasServices(n/2, activities, services)
	activityHasIssues := s.SeedActivityHasIssues(n/2, activities, issues)
	evidences := s.SeedEvidences(n, activities, users)
	componentVersionIssues := s.SeedComponentVersionIssues(n/2, componentVersions, issues)
	issueMatches := s.SeedIssueMatches(n, issues, componentInstances, users)
	issueMatchEvidences := s.SeedIssueMatchEvidence(n/2, issueMatches, evidences)
	issueRepositoryServices := s.SeedIssueRepositoryServices(n/2, services, repos)
	issueMatchChanges := s.SeedIssueMatchChanges(n, issueMatches, activities)

	return &SeedCollection{
		IssueVariantRows:           issueVariants,
		IssueRepositoryRows:        repos,
		UserRows:                   users,
		IssueRows:                  issues,
		IssueMatchRows:             issueMatches,
		ActivityRows:               activities,
		EvidenceRows:               evidences,
		ComponentInstanceRows:      componentInstances,
		ComponentVersionRows:       componentVersions,
		ComponentRows:              components,
		ServiceRows:                services,
		SupportGroupUserRows:       supportGroupUsers,
		SupportGroupRows:           supportGroups,
		SupportGroupServiceRows:    supportGroupServices,
		OwnerRows:                  owners,
		ActivityHasServiceRows:     activityHasServices,
		ActivityHasIssueRows:       activityHasIssues,
		ComponentVersionIssueRows:  componentVersionIssues,
		IssueMatchEvidenceRows:     issueMatchEvidences,
		IssueRepositoryServiceRows: issueRepositoryServices,
		IssueMatchChangeRows:       issueMatchChanges,
	}
}

func (s *DatabaseSeeder) SeedDbWithFakeData() {
	s.SeedDbWithNFakeData(100)
}

func (s *DatabaseSeeder) SeedDbForNestedIssueVariantTest() *SeedCollection {
	users := s.SeedUsers(1)
	services := s.SeedServices(1)
	components := s.SeedComponents(1)
	componentVersions := s.SeedComponentVersions(1, components)
	componentInstances := s.SeedComponentInstances(1, componentVersions, services)
	repos := s.SeedIssueRepositories()
	issues := s.SeedIssues(1)
	issueVariants := s.SeedIssueVariants(100, repos, issues)
	issueMatches := s.SeedIssueMatches(1, issues, componentInstances, users)
	issueRepositoryServices := s.SeedIssueRepositoryServices(len(repos), services, repos)
	return &SeedCollection{
		IssueVariantRows:           issueVariants,
		IssueRepositoryRows:        repos,
		UserRows:                   users,
		IssueRows:                  issues,
		IssueMatchRows:             issueMatches,
		ActivityRows:               nil,
		EvidenceRows:               nil,
		ComponentInstanceRows:      componentInstances,
		ComponentVersionRows:       componentVersions,
		ComponentRows:              components,
		ServiceRows:                services,
		SupportGroupUserRows:       nil,
		SupportGroupRows:           nil,
		SupportGroupServiceRows:    nil,
		OwnerRows:                  nil,
		ActivityHasServiceRows:     nil,
		ActivityHasIssueRows:       nil,
		ComponentVersionIssueRows:  nil,
		IssueMatchEvidenceRows:     nil,
		IssueRepositoryServiceRows: issueRepositoryServices,
	}
}

func (s *DatabaseSeeder) SeedForIssueCounts() (*SeedCollection, error) {
	issueRepositories := s.SeedIssueRepositories()
	supportGroups := s.SeedSupportGroups(2)
	issues := s.SeedIssues(10)
	components := s.SeedComponents(1)
	componentVersions := s.SeedComponentVersions(10, components)
	services := s.SeedServices(5)
	issueVariants, err := LoadIssueVariants(GetTestDataPath("../testdata/component_version_order/issue_variant.json"))
	if err != nil {
		return nil, err
	}
	cvIssueRows, err := LoadComponentVersionIssues(GetTestDataPath("../testdata/service_order/component_version_issue.json"))
	if err != nil {
		return nil, err
	}
	componentInstances, err := LoadComponentInstances(GetTestDataPath("../testdata/service_order/component_instance.json"))
	if err != nil {
		return nil, err
	}
	issueMatches, err := LoadIssueMatches(GetTestDataPath("../testdata/service_order/issue_match.json"))
	if err != nil {
		return nil, err
	}
	supportGroupServices, err := LoadSupportGroupServices(GetTestDataPath("../testdata/issue_counts/support_group_service.json"))
	if err != nil {
		return nil, err
	}

	// Important: the order need to be preserved
	for _, iv := range issueVariants {
		_, err := s.InsertFakeIssueVariant(iv)
		if err != nil {
			return nil, err
		}
	}
	for _, cvi := range cvIssueRows {
		_, err := s.InsertFakeComponentVersionIssue(cvi)
		if err != nil {
			return nil, err
		}
	}
	for _, ci := range componentInstances {
		_, err := s.InsertFakeComponentInstance(ci)
		if err != nil {
			return nil, err
		}
	}
	for _, im := range issueMatches {
		_, err := s.InsertFakeIssueMatch(im)
		if err != nil {
			return nil, err
		}
	}
	for _, sgs := range supportGroupServices {
		_, err := s.InsertFakeSupportGroupService(sgs)
		if err != nil {
			return nil, err
		}
	}
	return &SeedCollection{
		IssueVariantRows:           issueVariants,
		IssueRepositoryRows:        issueRepositories,
		UserRows:                   nil,
		IssueRows:                  issues,
		IssueMatchRows:             issueMatches,
		ActivityRows:               nil,
		EvidenceRows:               nil,
		ComponentInstanceRows:      componentInstances,
		ComponentVersionRows:       componentVersions,
		ComponentRows:              components,
		ServiceRows:                services,
		SupportGroupUserRows:       nil,
		SupportGroupRows:           supportGroups,
		SupportGroupServiceRows:    supportGroupServices,
		OwnerRows:                  nil,
		ActivityHasServiceRows:     nil,
		ActivityHasIssueRows:       nil,
		ComponentVersionIssueRows:  cvIssueRows,
		IssueMatchEvidenceRows:     nil,
		IssueRepositoryServiceRows: nil,
		IssueMatchChangeRows:       nil,
	}, nil
}

func (s *DatabaseSeeder) SeedIssueRepositories() []mariadb.BaseIssueRepositoryRow {
	variants := []string{
		"Converged Cloud",
		"SAP Global",
		"RedHat Advisory",
		"Github Advisory",
		"Nist",
	}
	repos := make([]mariadb.BaseIssueRepositoryRow, len(variants))
	i := 0
	for _, name := range variants {
		row := mariadb.BaseIssueRepositoryRow{
			Name:      sql.NullString{String: fmt.Sprintf("%s-%s", name, gofakeit.UUID()), Valid: true},
			Url:       sql.NullString{String: gofakeit.URL(), Valid: true},
			CreatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
			UpdatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
		}
		id, err := s.InsertFakeBaseIssueRepository(row)
		if err != nil {
			logrus.WithField("seed_type", "IssueRepository").Debug(err)
			continue
		}
		row.Id = sql.NullInt64{Int64: id, Valid: true}
		repos[i] = row
		i = i + 1
	}
	return repos
}

func (s *DatabaseSeeder) SeedIssueMatches(num int, issues []mariadb.IssueRow, componentInstances []mariadb.ComponentInstanceRow, users []mariadb.UserRow) []mariadb.IssueMatchRow {
	var issueMatches []mariadb.IssueMatchRow
	for i := 0; i < num; i++ {
		im := NewFakeIssueMatch()

		randomUserIndex := rand.Intn(len(users))
		im.UserId = users[randomUserIndex].Id

		randomIIndex := rand.Intn(len(issues))
		i := issues[randomIIndex]
		im.IssueId = i.Id

		randomCiIndex := rand.Intn(len(componentInstances))
		ci := componentInstances[randomCiIndex]
		im.ComponentInstanceId = ci.Id

		imId, err := s.InsertFakeIssueMatch(im)
		if err != nil {
			core.GinkgoLogr.WithValues("im", im).Error(err, "Error while inserting IssueMatch")
			logrus.WithField("seed_type", "IssueMatches").Debug(err)
		} else {
			im.Id = sql.NullInt64{Int64: imId, Valid: true}
			issueMatches = append(issueMatches, im)
		}
	}
	return issueMatches
}

func (s *DatabaseSeeder) SeedIssues(num int) []mariadb.IssueRow {
	var issues []mariadb.IssueRow
	for i := 0; i < num; i++ {
		issue := NewFakeIssue()
		iId, err := s.InsertFakeIssue(issue)
		if err != nil {
			logrus.WithField("seed_type", "Issues").Debug(err)
		} else {
			issue.Id = sql.NullInt64{Int64: iId, Valid: true}
			issues = append(issues, issue)
		}
	}
	return issues
}

func (s *DatabaseSeeder) SeedIssueVariants(num int, repos []mariadb.BaseIssueRepositoryRow, issues []mariadb.IssueRow) []mariadb.IssueVariantRow {
	var variants []mariadb.IssueVariantRow
	for i := 0; i < num; i++ {
		variant := NewFakeIssueVariant(repos, issues)
		variantId, err := s.InsertFakeIssueVariant(variant)
		if err != nil {
			logrus.WithField("seed_type", "IssueVariants").Debug(err)
		} else {
			variant.Id = sql.NullInt64{Int64: variantId, Valid: true}
			variants = append(variants, variant)
		}
	}
	return variants
}

func (s *DatabaseSeeder) SeedSupportGroups(num int) []mariadb.SupportGroupRow {
	var supportGroups []mariadb.SupportGroupRow
	for i := 0; i < num; i++ {
		supportGroup := NewFakeSupportGroup()
		supportGroupId, err := s.InsertFakeSupportGroup(supportGroup)
		if err != nil {
			logrus.WithField("seed_type", "SupportGroups").Debug(err)
		} else {
			supportGroup.Id = sql.NullInt64{Int64: supportGroupId, Valid: true}
			supportGroups = append(supportGroups, supportGroup)
		}
	}
	return supportGroups
}

func (s *DatabaseSeeder) SeedSupportGroupServices(num int, services []mariadb.BaseServiceRow, supportGroups []mariadb.SupportGroupRow) []mariadb.SupportGroupServiceRow {
	var rows []mariadb.SupportGroupServiceRow
	for i := 0; i < num; i++ {
		sgs := NewFakeSupportGroupService()
		randomSIndex := rand.Intn(len(services))
		service := services[randomSIndex]
		sgs.ServiceId = service.Id
		randomSGIndex := rand.Intn(len(supportGroups))
		sg := supportGroups[randomSGIndex]
		sgs.SupportGroupId = sg.Id
		_, err := s.InsertFakeSupportGroupService(sgs)
		if err != nil {
			logrus.WithField("seed_type", "SupportGroupServices").Debug(err)
		} else {
			rows = append(rows, sgs)
		}
	}
	return rows
}

func (s *DatabaseSeeder) SeedSupportGroupUsers(num int, users []mariadb.UserRow, supportGroups []mariadb.SupportGroupRow) []mariadb.SupportGroupUserRow {
	var rows []mariadb.SupportGroupUserRow
	for i := 0; i < num; i++ {
		sgu := NewFakeSupportGroupUser()
		randomUIndex := rand.Intn(len(users))
		user := users[randomUIndex]
		sgu.UserId = user.Id
		randomSGIndex := rand.Intn(len(supportGroups))
		sg := supportGroups[randomSGIndex]
		sgu.SupportGroupId = sg.Id
		_, err := s.InsertFakeSupportGroupUser(sgu)
		if err != nil {
			logrus.WithField("seed_type", "SupportGroupUsers").Debug(err)
		} else {
			rows = append(rows, sgu)
		}
	}
	return rows
}

func (s *DatabaseSeeder) SeedServices(num int) []mariadb.BaseServiceRow {
	var services []mariadb.BaseServiceRow
	for i := 0; i < num; i++ {
		service := NewFakeBaseService()
		serviceId, err := s.InsertFakeBaseService(service)
		if err != nil {
			logrus.WithField("seed_type", "Services").Debug(err)
		} else {
			service.Id = sql.NullInt64{Int64: serviceId, Valid: true}
			services = append(services, service)
		}
	}
	return services
}

func (s *DatabaseSeeder) SeedComponents(num int) []mariadb.ComponentRow {
	var components []mariadb.ComponentRow
	for i := 0; i < num; i++ {
		component := NewFakeComponent()
		componentId, err := s.InsertFakeComponent(component)
		if err != nil {
			logrus.WithField("seed_type", "Components").Debug(err)
		} else {
			component.Id = sql.NullInt64{Int64: componentId, Valid: true}
			components = append(components, component)
		}

	}
	return components
}

func (s *DatabaseSeeder) SeedComponentVersions(num int, components []mariadb.ComponentRow) []mariadb.ComponentVersionRow {
	var componentVersions []mariadb.ComponentVersionRow
	for i := 0; i < num; i++ {
		componentVersion := NewFakeComponentVersion()
		randomIndex := rand.Intn(len(components))
		component := components[randomIndex]
		componentVersion.ComponentId = component.Id
		componentVersionId, err := s.InsertFakeComponentVersion(componentVersion)
		if err != nil {
			logrus.WithField("seed_type", "ComponentVersions").Debug(err)
		} else {
			componentVersion.Id = sql.NullInt64{Int64: componentVersionId, Valid: true}
			componentVersions = append(componentVersions, componentVersion)
		}

	}
	return componentVersions
}

func (s *DatabaseSeeder) SeedComponentInstances(num int, componentVersions []mariadb.ComponentVersionRow, services []mariadb.BaseServiceRow) []mariadb.ComponentInstanceRow {
	var componentInstances []mariadb.ComponentInstanceRow
	for i := 0; i < num; i++ {
		componentInstance := NewFakeComponentInstance()
		randomCvIndex := rand.Intn(len(componentVersions))
		componentVersion := componentVersions[randomCvIndex]
		componentInstance.ComponentVersionId = componentVersion.Id
		randomSIndex := rand.Intn(len(services))
		service := services[randomSIndex]
		componentInstance.ServiceId = service.Id
		componentInstanceId, err := s.InsertFakeComponentInstance(componentInstance)
		if err != nil {
			logrus.WithField("seed_type", "ComponentInstances").Debug(err)
		} else {
			componentInstance.Id = sql.NullInt64{Int64: componentInstanceId, Valid: true}
			componentInstances = append(componentInstances, componentInstance)
		}
	}
	return componentInstances
}

func (s *DatabaseSeeder) SeedUsers(num int) []mariadb.UserRow {
	var users []mariadb.UserRow
	for i := 0; i < num; i++ {
		user := NewFakeUser()
		userId, err := s.InsertFakeUser(user)
		if err != nil {
			logrus.WithField("seed_type", "Users").Debug(err)
		} else {
			user.Id = sql.NullInt64{Int64: userId, Valid: true}
			users = append(users, user)
		}
	}
	return users
}

func (s *DatabaseSeeder) SeedOwners(num int, services []mariadb.BaseServiceRow, users []mariadb.UserRow) []mariadb.OwnerRow {
	var owners []mariadb.OwnerRow
	for i := 0; i < num; i++ {
		owner := NewFakeOwner()
		randomSIndex := rand.Intn(len(services))
		service := services[randomSIndex]
		owner.ServiceId = service.Id
		randomUIndex := rand.Intn(len(users))
		user := users[randomUIndex]
		owner.UserId = user.Id
		_, err := s.InsertFakeOwner(owner)
		if err != nil {
			logrus.WithField("seed_type", "Owners").Debug(err)
		} else {
			owners = append(owners, owner)
		}
	}
	return owners
}

func (s *DatabaseSeeder) SeedActivities(num int) []mariadb.ActivityRow {
	var activities []mariadb.ActivityRow
	for i := 0; i < num; i++ {
		activity := NewFakeActivity()
		activityId, err := s.InsertFakeActivity(activity)
		if err != nil {
			logrus.WithField("seed_type", "Activities").Debug(err)
		} else {
			activity.Id = sql.NullInt64{Int64: activityId, Valid: true}
			activities = append(activities, activity)
		}
	}
	return activities
}

func (s *DatabaseSeeder) SeedActivityHasServices(num int, activities []mariadb.ActivityRow, services []mariadb.BaseServiceRow) []mariadb.ActivityHasServiceRow {
	var ahsList []mariadb.ActivityHasServiceRow
	for i := 0; i < num; i++ {
		ahs := NewFakeActivityHasService()
		randomSIndex := rand.Intn(len(services))
		service := services[randomSIndex]
		ahs.ServiceId = service.Id
		randomAIndex := rand.Intn(len(activities))
		activity := activities[randomAIndex]
		ahs.ActivityId = activity.Id
		_, err := s.InsertFakeActivityHasService(ahs)
		if err != nil {
			logrus.WithField("seed_type", "ActivityHasServices").Debug(err)
		} else {
			ahsList = append(ahsList, ahs)
		}
	}
	return ahsList
}

func (s *DatabaseSeeder) SeedIssueMatchEvidence(num int, im []mariadb.IssueMatchRow, e []mariadb.EvidenceRow) []mariadb.IssueMatchEvidenceRow {
	var imeList []mariadb.IssueMatchEvidenceRow
	for i := 0; i < num; i++ {
		ime := NewFakeIssueMatchEvidence()
		randomImIndex := rand.Intn(len(im))
		im := im[randomImIndex]
		ime.IssueMatchId = im.Id
		randomEIndex := rand.Intn(len(e))
		evidence := e[randomEIndex]
		ime.EvidenceId = evidence.Id
		_, err := s.InsertFakeIssueMatchEvidence(ime)
		if err != nil {
			core.GinkgoLogr.WithValues("IssueMatchEvidence", ime).Error(err, "Error while creating IssueMatchEvidence")
			logrus.WithField("seed_type", "IssueMatchEvidences").Debug(err)
		} else {
			imeList = append(imeList, ime)
		}
	}
	return imeList
}

func (s *DatabaseSeeder) SeedActivityHasIssues(num int, activities []mariadb.ActivityRow, issues []mariadb.IssueRow) []mariadb.ActivityHasIssueRow {
	ahiList := make([]mariadb.ActivityHasIssueRow, num)
	for i := 0; i < num; i++ {
		ahi := NewFakeActivityHasIssue()
		randomIIndex := rand.Intn(len(issues))
		issue := issues[randomIIndex]
		ahi.IssueId = issue.Id
		randomAIndex := rand.Intn(len(activities))
		activity := activities[randomAIndex]
		ahi.ActivityId = activity.Id
		_, err := s.InsertFakeActivityHasIssue(ahi)
		if err != nil {
			logrus.WithField("seed_type", "ActivityHasIssues").Debug(err)
		}
		ahiList[i] = ahi
	}
	return ahiList
}

func (s *DatabaseSeeder) SeedEvidences(num int, activities []mariadb.ActivityRow, users []mariadb.UserRow) []mariadb.EvidenceRow {
	var evidences []mariadb.EvidenceRow
	for i := 0; i < num; i++ {
		evidence := NewFakeEvidence()
		randomAIndex := rand.Intn(len(activities))
		activity := activities[randomAIndex]
		evidence.ActivityId = activity.Id
		randomUIndex := rand.Intn(len(users))
		user := users[randomUIndex]
		evidence.UserId = user.Id
		evidenceId, err := s.InsertFakeEvidence(evidence)
		if err != nil {
			logrus.WithField("seed_type", "Evidences").Debug(err)
		} else {
			evidence.Id = sql.NullInt64{Int64: evidenceId, Valid: true}
			evidences = append(evidences, evidence)
		}
	}
	return evidences
}

func (s *DatabaseSeeder) SeedComponentVersionIssues(num int, componentVersions []mariadb.ComponentVersionRow, issues []mariadb.IssueRow) []mariadb.ComponentVersionIssueRow {
	cviList := make([]mariadb.ComponentVersionIssueRow, num)
	for i := 0; i < num; i++ {
		cvi := NewFakeComponentVersionIssue()
		randomIIndex := rand.Intn(len(issues))
		issue := issues[randomIIndex]
		cvi.IssueId = issue.Id
		randomCIndex := rand.Intn(len(componentVersions))
		componentVersion := componentVersions[randomCIndex]
		cvi.ComponentVersionId = componentVersion.Id
		_, err := s.InsertFakeComponentVersionIssue(cvi)
		if err != nil {
			logrus.WithField("seed_type", "ComponentVersionIssues").Debug(err)
		} else {
			cviList[i] = cvi
		}
	}
	return cviList
}

func (s *DatabaseSeeder) SeedIssueRepositoryServices(num int, services []mariadb.BaseServiceRow, issueRepositories []mariadb.BaseIssueRepositoryRow) []mariadb.IssueRepositoryServiceRow {
	var rows []mariadb.IssueRepositoryServiceRow
	for i := 0; i < num; i++ {
		irs := NewFakeIssueRepositoryService()
		irs.Priority = sql.NullInt64{Int64: int64(i), Valid: true}
		randomSIndex := rand.Intn(len(services))
		service := services[randomSIndex]
		irs.ServiceId = service.Id
		randomIRIndex := rand.Intn(len(issueRepositories))
		ir := issueRepositories[randomIRIndex]
		irs.IssueRepositoryId = ir.Id
		_, err := s.InsertFakeIssueRepositoryService(irs)
		if err != nil {
			logrus.WithField("seed_type", "IssueRepositoryServices").Debug(err)
		} else {
			rows = append(rows, irs)
		}
	}
	return rows
}

func (s *DatabaseSeeder) SeedIssueMatchChanges(num int, issueMatches []mariadb.IssueMatchRow, activities []mariadb.ActivityRow) []mariadb.IssueMatchChangeRow {
	var rows []mariadb.IssueMatchChangeRow
	for i := 0; i < num; i++ {
		imc := NewFakeIssueMatchChange()
		randomIMIndex := rand.Intn(len(issueMatches))
		issueMatch := issueMatches[randomIMIndex]
		imc.IssueMatchId = issueMatch.Id
		randomAIndex := rand.Intn(len(activities))
		activity := activities[randomAIndex]
		imc.ActivityId = activity.Id
		id, err := s.InsertFakeIssueMatchChange(imc)
		imc.Id = sql.NullInt64{Int64: id, Valid: true}
		if err != nil {
			logrus.WithField("seed_type", "IssueMatchChanges").Debug(err)
		} else {
			rows = append(rows, imc)
		}
	}
	return rows
}

func (s *DatabaseSeeder) InsertFakeIssueMatchEvidence(ime mariadb.IssueMatchEvidenceRow) (int64, error) {
	query := `
		INSERT INTO IssueMatchEvidence (
			issuematchevidence_evidence_id,
			issuematchevidence_issue_match_id
		) VALUES (
			:issuematchevidence_evidence_id,
			:issuematchevidence_issue_match_id
		)`
	return s.ExecPreparedNamed(query, ime)
}

func (s *DatabaseSeeder) InsertFakeIssue(issue mariadb.IssueRow) (int64, error) {
	query := `
		INSERT INTO Issue (
			issue_primary_name,
			issue_type,
			issue_description,
			issue_created_by,
			issue_updated_by
		) VALUES (
			:issue_primary_name,
			:issue_type,
			:issue_description,
			:issue_created_by,
			:issue_updated_by
		)`
	return s.ExecPreparedNamed(query, issue)
}

func (s *DatabaseSeeder) InsertFakeIssueVariant(issueVariant mariadb.IssueVariantRow) (int64, error) {
	query := `
		INSERT INTO IssueVariant (
			issuevariant_secondary_name,
			issuevariant_vector,
			issuevariant_rating,
			issuevariant_issue_id,
			issuevariant_repository_id,
			issuevariant_description,
			issuevariant_external_url,
			issuevariant_created_by,
			issuevariant_updated_by
		) VALUES (
			:issuevariant_secondary_name,
			:issuevariant_vector,
			:issuevariant_rating,
			:issuevariant_issue_id,
			:issuevariant_repository_id,
			:issuevariant_description,
			:issuevariant_external_url,
			:issuevariant_created_by,
			:issuevariant_updated_by
		)`
	return s.ExecPreparedNamed(query, issueVariant)
}

func (s *DatabaseSeeder) InsertFakeBaseIssueRepository(irr mariadb.BaseIssueRepositoryRow) (int64, error) {
	query := `
		INSERT INTO IssueRepository (
			issuerepository_name, 
			issuerepository_url,
			issuerepository_created_by,
			issuerepository_updated_by
		) VALUES (
			:issuerepository_name, 
			:issuerepository_url,
			:issuerepository_created_by,
			:issuerepository_updated_by
		)`
	return s.ExecPreparedNamed(query, irr)
}

func (s *DatabaseSeeder) InsertFakeIssueMatch(im mariadb.IssueMatchRow) (int64, error) {
	query := `
		INSERT INTO IssueMatch (
			issuematch_status,
			issuematch_component_instance_id,
			issuematch_vector,
			issuematch_rating,
			issuematch_issue_id,
			issuematch_user_id,
			issuematch_remediation_date,
			issuematch_target_remediation_date,
			issuematch_created_by,
			issuematch_updated_by
		) VALUES (
			:issuematch_status,
			:issuematch_component_instance_id,
			:issuematch_vector,
			:issuematch_rating,
			:issuematch_issue_id,
			:issuematch_user_id,
			:issuematch_remediation_date,
			:issuematch_target_remediation_date,
			:issuematch_created_by,
			:issuematch_updated_by
		)`
	return s.ExecPreparedNamed(query, im)
}

func (s *DatabaseSeeder) InsertFakeComponentInstance(ci mariadb.ComponentInstanceRow) (int64, error) {
	query := `
		INSERT INTO ComponentInstance (
			componentinstance_ccrn,
			componentinstance_region,
			componentinstance_cluster,
			componentinstance_namespace,
			componentinstance_domain,
			componentinstance_project,
			componentinstance_pod,
			componentinstance_container,
			componentinstance_type,
			componentinstance_parent_id,
			componentinstance_context,
			componentinstance_count,
			componentinstance_component_version_id,
			componentinstance_service_id,
			componentinstance_created_by,
			componentinstance_updated_by
		) VALUES (
			:componentinstance_ccrn,
			:componentinstance_region,
			:componentinstance_cluster,
			:componentinstance_namespace,
			:componentinstance_domain,
			:componentinstance_project,
			:componentinstance_pod,
			:componentinstance_container,
			:componentinstance_type,
			:componentinstance_parent_id,
			:componentinstance_context,
			:componentinstance_count,
			:componentinstance_component_version_id,
			:componentinstance_service_id,
			:componentinstance_created_by,
			:componentinstance_updated_by
		)`
	return s.ExecPreparedNamed(query, ci)
}

func (s *DatabaseSeeder) InsertFakeBaseService(service mariadb.BaseServiceRow) (int64, error) {
	query := `
		INSERT INTO Service (
			service_ccrn,
			service_created_by,
			service_updated_by
		) VALUES (
			:service_ccrn,
			:service_created_by,
			:service_updated_by
		)`
	return s.ExecPreparedNamed(query, service)
}

func (s *DatabaseSeeder) InsertFakeSupportGroup(sg mariadb.SupportGroupRow) (int64, error) {
	query := `
		INSERT INTO SupportGroup (
			supportgroup_ccrn,
			supportgroup_created_by,
			supportgroup_updated_by
		) VALUES (
			:supportgroup_ccrn,
			:supportgroup_created_by,
			:supportgroup_updated_by
		)`
	return s.ExecPreparedNamed(query, sg)
}

func (s *DatabaseSeeder) InsertFakeComponent(component mariadb.ComponentRow) (int64, error) {
	query := `
		INSERT INTO Component (
			component_ccrn,
			component_type,
			component_created_by,
			component_updated_by
		) VALUES (
			:component_ccrn,
			:component_type,
			:component_created_by,
			:component_updated_by
		)`
	return s.ExecPreparedNamed(query, component)
}

func (s *DatabaseSeeder) InsertFakeComponentVersion(cv mariadb.ComponentVersionRow) (int64, error) {
	query := `
		INSERT INTO ComponentVersion (
			componentversion_version,
			componentversion_component_id,
            componentversion_tag,
            componentversion_repository,
            componentversion_organization,
			componentversion_created_by,
			componentversion_updated_by
		) VALUES (
			:componentversion_version,
			:componentversion_component_id,
            :componentversion_tag,
            :componentversion_repository,
            :componentversion_organization,
			:componentversion_created_by,
			:componentversion_updated_by
		)`
	return s.ExecPreparedNamed(query, cv)
}

func (s *DatabaseSeeder) InsertFakeUser(user mariadb.UserRow) (int64, error) {
	query := `
		INSERT INTO User (
			user_name,
			user_unique_user_id,
			user_type,
			user_email,
			user_created_by,
			user_updated_by
		) VALUES (
			:user_name,
			:user_unique_user_id,
			:user_type,
			:user_email,
			:user_created_by,
			:user_updated_by
		)`
	return s.ExecPreparedNamed(query, user)
}

func (s *DatabaseSeeder) InsertFakeOwner(owner mariadb.OwnerRow) (int64, error) {
	query := `
		INSERT INTO Owner (
			owner_service_id,
			owner_user_id
		) VALUES (
			:owner_service_id,
			:owner_user_id
		)`
	return s.ExecPreparedNamed(query, owner)
}

func (s *DatabaseSeeder) InsertFakeSupportGroupUser(sgu mariadb.SupportGroupUserRow) (int64, error) {
	query := `
		INSERT INTO SupportGroupUser (
			supportgroupuser_user_id,
			supportgroupuser_support_group_id
		) VALUES (
			:supportgroupuser_user_id,
			:supportgroupuser_support_group_id
		)`
	return s.ExecPreparedNamed(query, sgu)
}

func (s *DatabaseSeeder) InsertFakeSupportGroupService(sgs mariadb.SupportGroupServiceRow) (int64, error) {
	query := `
		INSERT INTO SupportGroupService (
			supportgroupservice_service_id,
			supportgroupservice_support_group_id
		) VALUES (
			:supportgroupservice_service_id,
			:supportgroupservice_support_group_id
		)`
	return s.ExecPreparedNamed(query, sgs)
}

func (s *DatabaseSeeder) InsertFakeActivity(activity mariadb.ActivityRow) (int64, error) {
	query := `
		INSERT INTO Activity (
			activity_status,
			activity_created_by,
			activity_updated_by
		) VALUES (
			:activity_status,
			:activity_created_by,
			:activity_updated_by
		)`
	return s.ExecPreparedNamed(query, activity)
}

func (s *DatabaseSeeder) InsertFakeActivityHasService(ahs mariadb.ActivityHasServiceRow) (int64, error) {
	query := `
		INSERT INTO ActivityHasService (
			activityhasservice_activity_id,
			activityhasservice_service_id
		) VALUES (
			:activityhasservice_activity_id,
			:activityhasservice_service_id
		)`
	return s.ExecPreparedNamed(query, ahs)
}

func (s *DatabaseSeeder) InsertFakeActivityHasIssue(ahi mariadb.ActivityHasIssueRow) (int64, error) {
	query := `
		INSERT INTO ActivityHasIssue (
			activityhasissue_activity_id,
			activityhasissue_issue_id
		) VALUES (
			:activityhasissue_activity_id,
			:activityhasissue_issue_id
		)`
	return s.ExecPreparedNamed(query, ahi)
}

func (s *DatabaseSeeder) InsertFakeEvidence(evidence mariadb.EvidenceRow) (int64, error) {
	query := `
		INSERT INTO Evidence (
			evidence_description,
			evidence_type,
			evidence_vector,
			evidence_rating,
			evidence_raa_end,
			evidence_author_id,
			evidence_activity_id,
			evidence_created_by,
			evidence_updated_by
		) VALUES (
			:evidence_description,
			:evidence_type,
			:evidence_vector,
			:evidence_rating,
			:evidence_raa_end,
			:evidence_author_id,
			:evidence_activity_id,
			:evidence_created_by,
			:evidence_updated_by
		)`
	return s.ExecPreparedNamed(query, evidence)
}

func (s *DatabaseSeeder) InsertFakeComponentVersionIssue(cvi mariadb.ComponentVersionIssueRow) (int64, error) {
	query := `
		INSERT INTO ComponentVersionIssue (
			componentversionissue_component_version_id,
			componentversionissue_issue_id
		) VALUES (
			:componentversionissue_component_version_id,
			:componentversionissue_issue_id
		)`
	return s.ExecPreparedNamed(query, cvi)
}

func (s *DatabaseSeeder) InsertFakeIssueRepositoryService(sgs mariadb.IssueRepositoryServiceRow) (int64, error) {
	query := `
		INSERT INTO IssueRepositoryService (
			issuerepositoryservice_service_id,
			issuerepositoryservice_issue_repository_id,
			issuerepositoryservice_priority
		) VALUES (
			:issuerepositoryservice_service_id,
			:issuerepositoryservice_issue_repository_id,
			:issuerepositoryservice_priority
		)`
	return s.ExecPreparedNamed(query, sgs)
}

func (s *DatabaseSeeder) InsertFakeIssueMatchChange(vmc mariadb.IssueMatchChangeRow) (int64, error) {
	query := `
		INSERT INTO IssueMatchChange (
			issuematchchange_activity_id,
			issuematchchange_issue_match_id,
			issuematchchange_action,
			issuematchchange_created_by,
			issuematchchange_updated_by
		) VALUES (
			:issuematchchange_activity_id,
			:issuematchchange_issue_match_id,
			:issuematchchange_action,
			:issuematchchange_created_by,
			:issuematchchange_updated_by
		)`
	return s.ExecPreparedNamed(query, vmc)
}

func (s *DatabaseSeeder) ExecPreparedNamed(query string, obj any) (int64, error) {
	stmt, err := s.db.PrepareNamed(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(obj)
	if err != nil {
		return 0, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastId, nil
}

func NewFakeIssueMatch() mariadb.IssueMatchRow {
	v := GenerateRandomCVSS31Vector()
	cvss, _ := metric.NewEnvironmental().Decode(v)
	rating := cvss.Severity().String()
	return mariadb.IssueMatchRow{
		Status:                sql.NullString{String: gofakeit.RandomString(entity.AllIssueMatchStatusValues), Valid: true},
		Vector:                sql.NullString{String: v, Valid: true},
		Rating:                sql.NullString{String: rating, Valid: true},
		RemediationDate:       sql.NullTime{Time: gofakeit.Date(), Valid: true},
		TargetRemediationDate: sql.NullTime{Time: gofakeit.Date(), Valid: true},
		CreatedBy:             sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
		UpdatedBy:             sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
	}
}

func NewFakeIssue() mariadb.IssueRow {
	return mariadb.IssueRow{
		PrimaryName: sql.NullString{String: fmt.Sprintf("CVE-%d-%d", gofakeit.Year(), gofakeit.Number(100, 9999999)), Valid: true},
		Description: sql.NullString{String: gofakeit.HackerPhrase(), Valid: true},
		Type:        sql.NullString{String: gofakeit.RandomString(entity.AllIssueTypes), Valid: true},
		CreatedBy:   sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
		UpdatedBy:   sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
	}
}

func NewFakeIssueVariant(repos []mariadb.BaseIssueRepositoryRow, disc []mariadb.IssueRow) mariadb.IssueVariantRow {
	variants := []string{"GHSA", "RHSA", "VMSA"}
	v := GenerateRandomCVSS31Vector()
	cvss, _ := metric.NewEnvironmental().Decode(v)
	rating := cvss.Severity().String()
	if rating == "" {
		rating = "None"
	}
	externalUrl := gofakeit.URL()
	return mariadb.IssueVariantRow{
		SecondaryName: sql.NullString{String: fmt.Sprintf("%s-%d-%d", gofakeit.RandomString(variants), gofakeit.Year(), gofakeit.Number(1000, 9999999)), Valid: true},
		Description:   sql.NullString{String: gofakeit.HackerPhrase(), Valid: true},
		Vector:        sql.NullString{String: v, Valid: true},
		Rating:        sql.NullString{String: rating, Valid: true},
		ExternalUrl:   sql.NullString{String: externalUrl, Valid: true},
		IssueRepositoryId: sql.NullInt64{
			Int64: repos[rand.Intn(len(repos))].Id.Int64,
			Valid: true,
		},
		IssueId: sql.NullInt64{
			Int64: disc[rand.Intn(len(disc))].Id.Int64,
			Valid: true,
		},
		CreatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
		UpdatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
	}
}

func NewFakeIssueRepository() mariadb.IssueRepositoryRow {
	return mariadb.IssueRepositoryRow{
		BaseIssueRepositoryRow: mariadb.BaseIssueRepositoryRow{
			Name:      sql.NullString{String: fmt.Sprintf("%s-%s", gofakeit.AppName(), gofakeit.UUID()), Valid: true},
			Url:       sql.NullString{String: gofakeit.URL(), Valid: true},
			CreatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
			UpdatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
		},
	}
}

func NewFakeBaseService() mariadb.BaseServiceRow {
	return mariadb.BaseServiceRow{
		CCRN:      sql.NullString{String: fmt.Sprintf("%s-%s", gofakeit.AppName(), gofakeit.UUID()), Valid: true},
		CreatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
		UpdatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
	}
}

func NewFakeService() mariadb.ServiceRow {
	baseServiceRow := NewFakeBaseService()
	issueServiceRow := NewFakeIssueRepositoryService()
	return mariadb.ServiceRow{
		BaseServiceRow:            baseServiceRow,
		IssueRepositoryServiceRow: issueServiceRow,
	}
}

func NewFakeSupportGroup() mariadb.SupportGroupRow {
	return mariadb.SupportGroupRow{
		CCRN:      sql.NullString{String: fmt.Sprintf("%s-%s", gofakeit.AppName(), gofakeit.UUID()), Valid: true},
		CreatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
		UpdatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
	}
}

func NewFakeComponent() mariadb.ComponentRow {
	types := []string{"containerImage", "virtualMachineImage", "repository"}
	ccrn := fmt.Sprintf("%s-%d", gofakeit.AppName(), gofakeit.UUID())
	return mariadb.ComponentRow{
		CCRN:      sql.NullString{String: ccrn, Valid: true},
		Type:      sql.NullString{String: gofakeit.RandomString(types), Valid: true},
		CreatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
		UpdatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
	}
}

func NewFakeComponentVersion() mariadb.ComponentVersionRow {
	return mariadb.ComponentVersionRow{
		Version:      sql.NullString{String: gofakeit.Regex("^sha:[a-fA-F0-9]{64}$"), Valid: true},
		Tag:          sql.NullString{String: gofakeit.AppVersion(), Valid: true},
		Repository:   sql.NullString{String: gofakeit.AppName(), Valid: true},
		Organization: sql.NullString{String: gofakeit.Username(), Valid: true},
		CreatedBy:    sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
		UpdatedBy:    sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
	}
}

func GenerateFakeCcrn(cluster string, namespace string) string {
	return fmt.Sprintf("ccrn: apiVersion=k8s-registry.ccrn.sap.cloud/v1, kind=container, cluster=%s, namespace=%s, name=audit-logger-xe9mtzmq8l-cmbp6", cluster, namespace)
}

func NewFakeComponentInstance() mariadb.ComponentInstanceRow {
	n := gofakeit.Int16()
	if n < 0 {
		n = n * -1
	}
	region := gofakeit.RandomString([]string{"test-de-1", "test-de-2", "test-us-1", "test-jp-2", "test-jp-1"})
	cluster := gofakeit.RandomString([]string{"test-de-1", "test-de-2", "test-us-1", "test-jp-2", "test-jp-1", "a-test-de-1", "a-test-de-2", "a-test-us-1", "a-test-jp-2", "a-test-jp-1", "v-test-de-1", "v-test-de-2", "v-test-us-1", "v-test-jp-2", "v-test-jp-1", "s-test-de-1", "s-test-de-2", "s-test-us-1", "s-test-jp-2", "s-test-jp-1"})
	//make lower case to avoid conflicts in different lexicographical ordering between sql and golang due to collation
	namespace := strings.ToLower(gofakeit.ProductName())
	domain := strings.ToLower(gofakeit.SongName())
	project := strings.ToLower(gofakeit.BeerName())
	pod := strings.ToLower(gofakeit.UUID())
	container := strings.ToLower(gofakeit.UUID())
	t := gofakeit.RandomString(entity.AllComponentInstanceType)
	context := entity.Json{
		"timeout_nbd":               gofakeit.Float32(),
		"remove_unused_base_images": gofakeit.Bool(),
		"my_ip":                     gofakeit.IPv4Address(),
	}
	return mariadb.ComponentInstanceRow{
		CCRN:      sql.NullString{String: GenerateFakeCcrn(cluster, namespace), Valid: true},
		Region:    sql.NullString{String: region, Valid: true},
		Cluster:   sql.NullString{String: cluster, Valid: true},
		Namespace: sql.NullString{String: namespace, Valid: true},
		Domain:    sql.NullString{String: domain, Valid: true},
		Project:   sql.NullString{String: project, Valid: true},
		Pod:       sql.NullString{String: pod, Valid: true},
		Container: sql.NullString{String: container, Valid: true},
		Type:      sql.NullString{String: t, Valid: true},
		Context:   sql.NullString{String: context.String(), Valid: true},
		Count:     sql.NullInt16{Int16: n, Valid: true},
		CreatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
		UpdatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
	}
}

var nextUserType entity.UserType = entity.HumanUserType

func getNextUserType() int64 {
	userType := nextUserType
	if userType == entity.HumanUserType {
		nextUserType = entity.TechnicalUserType
	} else {
		nextUserType = entity.HumanUserType
	}
	return int64(userType)
}

func NewFakeUser() mariadb.UserRow {
	uniqueUserId := fmt.Sprintf("I%d", gofakeit.IntRange(100000, 999999))
	return mariadb.UserRow{
		Name:         sql.NullString{String: gofakeit.Name(), Valid: true},
		UniqueUserID: sql.NullString{String: uniqueUserId, Valid: true},
		Type:         sql.NullInt64{Int64: getNextUserType(), Valid: true},
		Email:        sql.NullString{String: gofakeit.Email(), Valid: true},
		CreatedBy:    sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
		UpdatedBy:    sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
	}
}

func NewFakeOwner() mariadb.OwnerRow {
	return mariadb.OwnerRow{}
}

func NewFakeSupportGroupService() mariadb.SupportGroupServiceRow {
	return mariadb.SupportGroupServiceRow{}
}

func NewFakeSupportGroupUser() mariadb.SupportGroupUserRow {
	return mariadb.SupportGroupUserRow{}
}

func NewFakeActivity() mariadb.ActivityRow {
	status := []string{"open", "closed", "in_progress"}
	return mariadb.ActivityRow{
		Status:    sql.NullString{String: gofakeit.RandomString(status), Valid: true},
		CreatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
		UpdatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
	}
}

func NewFakeActivityHasService() mariadb.ActivityHasServiceRow {
	return mariadb.ActivityHasServiceRow{}
}

func NewFakeIssueMatchEvidence() mariadb.IssueMatchEvidenceRow {
	return mariadb.IssueMatchEvidenceRow{}
}

func NewFakeActivityHasIssue() mariadb.ActivityHasIssueRow {
	return mariadb.ActivityHasIssueRow{}
}

func NewFakeEvidence() mariadb.EvidenceRow {
	types := []string{"risk_accepted", "mitigated", "severity_adjustment", "false_positive", "reopen"}
	v := GenerateRandomCVSS31Vector()
	cvss, _ := metric.NewEnvironmental().Decode(v)
	rating := cvss.Severity().String()
	return mariadb.EvidenceRow{
		Description: sql.NullString{String: gofakeit.Sentence(10), Valid: true},
		Type: sql.NullString{
			String: gofakeit.RandomString(types),
			Valid:  true,
		},
		Vector:    sql.NullString{String: v, Valid: true},
		Rating:    sql.NullString{String: rating, Valid: true},
		RAAEnd:    sql.NullTime{Time: gofakeit.Date(), Valid: true},
		CreatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
		UpdatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
	}
}

func NewFakeComponentVersionIssue() mariadb.ComponentVersionIssueRow {
	return mariadb.ComponentVersionIssueRow{}
}

func NewFakeIssueRepositoryService() mariadb.IssueRepositoryServiceRow {
	return mariadb.IssueRepositoryServiceRow{}
}

func NewFakeIssueMatchChange() mariadb.IssueMatchChangeRow {
	return mariadb.IssueMatchChangeRow{
		Action: sql.NullString{
			String: gofakeit.RandomString(entity.AllIssueMatchChangeActions),
			Valid:  true,
		},
		CreatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
		UpdatedBy: sql.NullInt64{Int64: e2e_common.SystemUserId, Valid: true},
	}
}

func (s *DatabaseSeeder) SeedRealSupportGroups() map[string]mariadb.SupportGroupRow {
	supportGroups := map[string]mariadb.SupportGroupRow{}

	sgs := []string{
		"compute",
		"compute-storage-api",
		"containers",
		"email",
		"foundation",
		"identity",
		"network-api",
		"network-data",
		"network-lb",
		"network-security",
		"observability",
		"src",
		"storage",
	}

	for _, sg := range sgs {
		supportGroup := mariadb.SupportGroupRow{
			CCRN: sql.NullString{String: sg, Valid: true},
		}
		supportGroupId, err := s.InsertFakeSupportGroup(supportGroup)
		if err != nil {
			logrus.WithField("seed_type", "SupportGroups").Debug(err)
		} else {
			supportGroup.Id = sql.NullInt64{Int64: supportGroupId, Valid: true}
			supportGroups[sg] = supportGroup
		}
	}

	return supportGroups
}

func (s *DatabaseSeeder) SeedRealServices() map[string]mariadb.BaseServiceRow {
	services := map[string]mariadb.BaseServiceRow{}

	svs := []string{
		"aci",
		"alerting",
		"alertmanager",
		"apic",
		"arista",
		"asa",
		"asr",
		"audit",
		"barbican",
		"baremetal",
		"castellum",
		"certificates",
		"cinder",
		"compute",
		"concourse",
		"designate",
		"exporter",
		"f5",
		"gatekeeper",
		"glance",
		"go-pmtud",
		"ground",
		"hammertime",
		"hermes",
		"ironic",
		"k8s",
		"keppel",
		"keystone",
		"kube-parrot",
		"kubelet",
		"kubernikus",
		"limes",
		"lyra",
		"maia",
		"manila",
		"metrics",
		"n9k",
		"nanny",
		"netbox",
		"network",
		"neutron",
		"node",
		"nova",
		"octavia",
		"prometheus",
		"px",
		"replicant",
		"resources",
		"servicing",
		"snmp",
		"storage",
		"swift",
		"tailscale",
		"webshell",
	}

	for _, sv := range svs {
		service := mariadb.BaseServiceRow{
			CCRN: sql.NullString{String: sv, Valid: true},
		}
		serviceId, err := s.InsertFakeBaseService(service)
		if err != nil {
			logrus.WithField("seed_type", "Services").Debug(err)
		} else {
			service.Id = sql.NullInt64{Int64: serviceId, Valid: true}
			services[sv] = service
		}
	}

	return services
}

func (s *DatabaseSeeder) SeedRealSupportGroupService(services map[string]mariadb.BaseServiceRow, supportGroups map[string]mariadb.SupportGroupRow) []mariadb.SupportGroupServiceRow {
	sgs := []mariadb.SupportGroupServiceRow{}
	mapping := map[string]string{
		"aci":          "observability",
		"alerting":     "observability",
		"alertmanager": "observability",
		"apic":         "network-api",
		"arista":       "foundation",
		"asa":          "observability",
		"asr":          "network-data",
		"audit":        "observability",
		"barbican":     "foundation",
		"baremetal":    "compute",
		"castellum":    "containers",
		"certificates": "containers",
		"cinder":       "compute",
		"compute":      "compute",
		"concourse":    "containers",
		"designate":    "network-api",
		"exporter":     "compute-storage-api",
		"f5":           "network-lb",
		"gatekeeper":   "containers",
		"glance":       "compute-storage-api",
		"go-pmtud":     "containers",
		"ground":       "containers",
		"hammertime":   "containers",
		"hermes":       "identity",
		"ironic":       "foundation",
		"k8s":          "containers",
		"keppel":       "containers",
		"keystone":     "identity",
		"kube-parrot":  "containers",
		"kubelet":      "containers",
		"kubernikus":   "containers",
		"limes":        "containers",
		"lyra":         "containers",
		"maia":         "observability",
		"manila":       "compute-storage-api",
		"metrics":      "observability",
		"n9k":          "observability",
		"nanny":        "",
		"netbox":       "foundation",
		"network":      "compute",
		"neutron":      "network-api",
		"node":         "containers",
		"nova":         "compute-storage-api",
		"octavia":      "network-api",
		"prometheus":   "observability",
		"px":           "network-api",
		"replicant":    "src",
		"resources":    "containers",
		"servicing":    "containers",
		"snmp":         "observability",
		"storage":      "compute",
		"swift":        "storage",
		"tailscale":    "containers",
		"webshell":     "containers",
	}

	for service, sg := range mapping {
		if sg == "" {
			continue
		}

		sgsr := mariadb.SupportGroupServiceRow{
			SupportGroupId: sql.NullInt64{Int64: supportGroups[sg].Id.Int64, Valid: true},
			ServiceId:      sql.NullInt64{Int64: services[service].Id.Int64, Valid: true},
		}
		s.InsertFakeSupportGroupService(sgsr)
		sgs = append(sgs, sgsr)
	}
	return sgs
}

type ScannerRunDef struct {
	Tag                  string
	IsCompleted          bool
	Timestamp            time.Time
	Issues               []string
	Components           []string
	IssueMatchComponents []string // WARNING: This needs pairs of Issue name and compoenent name
}

func (s *DatabaseSeeder) SeedScannerRuns(scannerRunDefs ...ScannerRunDef) error {
	var err error

	insertScannerRun := `
		INSERT INTO ScannerRun (
			scannerrun_uuid,
			scannerrun_tag,
			scannerrun_start_run,
			scannerrun_end_run,
			scannerrun_is_completed
		) VALUES (
			?,
			?,
			?,
			?,
			?
		)
	`

	insertIssue := `
		INSERT INTO Issue (
			issue_type,
			issue_primary_name,
			issue_description
		) VALUES (
			'Vulnerability',
			?,
			?
		)
	`

	insertScannerRunIssueTracker := `
		INSERT INTO ScannerRunIssueTracker (
			scannerrunissuetracker_scannerrun_run_id,
			scannerrunissuetracker_issue_id
		) VALUES (
			?,
			?
		)
	`

	insertIntoService := `
	INSERT INTO Service (
			service_ccrn
		) VALUES (
			?
		)
	`

	insertIntoComponent := `
	INSERT INTO Component (
			component_ccrn,
			component_type
		) VALUES (
			?,
			'floopy disk'
		)
	`

	insertIntoComponentVersion := `
		INSERT INTO ComponentVersion (
			componentversion_version,
			componentversion_component_id,
			componentversion_created_by
		) VALUES (
			?,
			1,
			1
		)
	`

	insertIntoComponentInstance := `
		INSERT INTO ComponentInstance (
			componentinstance_ccrn,
			componentinstance_count,
			componentinstance_component_version_id,
			componentinstance_service_id,
			componentinstance_created_by
		) VALUES (
			?,
			1,
			1,
			1,
			1
		)
	`

	insertIntoIssueMatchComponent := `
		INSERT INTO IssueMatch (
			issuematch_status,
			issuematch_rating,
			issuematch_target_remediation_date,
			issuematch_user_id,
			issuematch_issue_id,
			issuematch_component_instance_id
		) VALUES (
			'new',
			'CRITICAL',
			current_timestamp(),
			1,
			?,
			?
		)
	`

	knownIssues := make(map[string]int)
	knownComponentInstance := make(map[string]int)
	serviceCounter := 0
	componentCounter := 0
	componentVersionCounter := 0

	for _, srd := range scannerRunDefs {
		res, err := s.db.Exec(insertScannerRun, gofakeit.UUID(), srd.Tag, srd.Timestamp, srd.Timestamp, srd.IsCompleted)

		if err != nil {
			return err

		}

		scannerrunId, err := res.LastInsertId()
		if err != nil {
			return err
		}

		for _, issue := range srd.Issues {

			if _, ok := knownIssues[issue]; !ok {
				res, err := s.db.Exec(insertIssue, issue, issue)
				if err != nil {
					return err

				}
				issueId, err := res.LastInsertId()
				if err != nil {
					return err
				}

				knownIssues[issue] = int(issueId)
			}

			if err != nil {
				return err
			}

			_, err = s.db.Exec(insertScannerRunIssueTracker, scannerrunId, knownIssues[issue])
			if err != nil {
				return err

			}
		}

		if len(srd.Components) > 0 {
			_, err = s.db.Exec(insertIntoService, fmt.Sprintf("service-%d", serviceCounter))
			if err != nil {
				return fmt.Errorf("InsertIntoService failed: %v", err)
			}
			serviceCounter++
			_, err = s.db.Exec(insertIntoComponent, fmt.Sprintf("component-%d", componentCounter))
			if err != nil {
				return fmt.Errorf("InsertIntoComponent failed: %v", err)
			}
			componentCounter++
			_, err = s.db.Exec(insertIntoComponentVersion, fmt.Sprintf("version-%d", componentVersionCounter))
			if err != nil {
				return fmt.Errorf("InsertIntoComponentVersion failed: %v", err)
			}
			componentVersionCounter++
			for _, component := range srd.Components {
				if _, ok := knownComponentInstance[component]; ok {
					continue
				}
				res, err = s.db.Exec(insertIntoComponentInstance, component)
				if err != nil {
					return fmt.Errorf("bad things insertintocomponentinstance: %v", err)
				}
				if resId, err := res.LastInsertId(); err != nil {
					return fmt.Errorf("bad things insertintocomponentinstance get lastInsertId %v", err)
				} else {
					knownComponentInstance[component] = int(resId)
				}
			}
		}

		if len(srd.IssueMatchComponents) > 0 {
			for i, _ := range srd.IssueMatchComponents {
				if i%2 != 0 {
					continue
				}

				issueName := srd.IssueMatchComponents[i]
				componentName := srd.IssueMatchComponents[i+1]
				_, err = s.db.Exec(insertIntoIssueMatchComponent, knownIssues[issueName], knownComponentInstance[componentName])
				if err != nil {
					return fmt.Errorf("InsertIntoIssueMatchComponent failed: %v", err)
				}
			}
		}
	}
	return err
}
func (s *DatabaseSeeder) SeedScannerRunInstances(uuids ...string) error {
	insertScannerRun := `
		INSERT INTO ScannerRun (
			scannerrun_uuid,
			scannerrun_tag,
			scannerrun_start_run,
			scannerrun_end_run,
			scannerrun_is_completed
		) VALUES (
			?,
			?,
			?,
			?,
			?
		)
	`
	for _, uuid := range uuids {
		_, err := s.db.Exec(insertScannerRun, uuid, "tag", time.Now(), time.Now(), false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *DatabaseSeeder) Clear() error {
	rows, err := s.db.Query(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var table string
	for rows.Next() {
		if err := rows.Scan(&table); err != nil {
			return err
		}
		_, err := s.db.Exec(fmt.Sprintf("SET FOREIGN_KEY_CHECKS = 0;TRUNCATE TABLE `%s`; SET FOREIGN_KEY_CHECKS = 1", table))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *DatabaseSeeder) RefreshServiceIssueCounters() error {
	rows, err := s.db.Query(`
		CALL refresh_mvServiceIssueCounts_proc();
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}
