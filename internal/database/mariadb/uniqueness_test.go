// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type uniquenessTestItemTemplate interface {
	setup(*mariadb.SqlDatabase)
	teardown()
	createItem()
	deleteItem()
	expectItemCount(int64)
	expectDeletedItemCount(int64)
	expectNoError()
	expectDuplicationError()
}

var testTemplates = map[string]uniquenessTestItemTemplate{
	"UserUniqueness":              &uniquenessUserTemplate{},
	"ComponentUniqueness":         &uniquenessComponentTemplate{},
	"ComponentVersionUniqueness":  &uniquenessComponentVersionTemplate{},
	"ServiceUniqueness":           &uniquenessServiceTemplate{},
	"ComponentInstanceUniqueness": &uniquenessComponentInstanceTemplate{},
	"IssueRepositoryUniqueness":   &uniquenessIssueRepositoryTemplate{},
	"IssueUniqueness":             &uniquenessIssueTemplate{},
	"IssueVariantUniqueness":      &uniquenessIssueVariantTemplate{},
}

var _ = Describe("Delete uniqueness", Ordered, Label("database", "Uniqueness"), func() {
	var db *mariadb.SqlDatabase
	BeforeAll(func() {
		db = dbm.NewTestSchema()
	})

	AfterAll(func() {
		dbm.TestTearDown(db)
	})

	for label, uut := range testTemplates {
		Context(label, Label(label), func() {
			BeforeEach(func() {
				uut.setup(db)
			})
			AfterEach(func() {
				uut.teardown()
			})
			When("Denies creating two active items with the same unique value", func() {
				BeforeEach(func() {
					// GIVEN Item with test unique values
					uut.createItem()
					uut.expectNoError()
				})
				Context("Another item with test unique values is created", func() {
					BeforeEach(func() {
						// WHEN Another item with test unique values is created
						uut.createItem()
					})
					It("Creates no new item and duplication error is returned", func() {
						// THEN No new item should be created
						uut.expectItemCount(1)

						// AND Error should be raised
						uut.expectDuplicationError()
					})
				})
			})
			When("Allows creating a new item with the same unique value after the previous one is soft-deleted", func() {
				BeforeEach(func() {
					// GIVEN Item with test unique values is created and soft-deleted
					uut.createItem()
					uut.expectNoError()
					uut.deleteItem()
					uut.expectNoError()
				})
				Context("Another item with test unique value is created", func() {
					BeforeEach(func() {
						// WHEN Another item with test unique value is created
						uut.createItem()
					})
					It("Creates new item and no error is returned", func() {
						// THEN New item should be created
						uut.expectItemCount(2)

						// AND Error should be nil
						uut.expectNoError()
					})
				})
			})
			When("Allows deletion of a new item with the same unique value after the previous one is soft-deleted", func() {
				BeforeEach(func() {
					// GIVEN Item with test unique value is created and soft-deleted
					uut.createItem()
					uut.expectNoError()
					uut.deleteItem()
					uut.expectNoError()

					// AND Another item with test unique value is created
					uut.createItem()
					uut.expectNoError()
				})
				Context("Delete new item", func() {
					BeforeEach(func() {
						// WHEN New item is soft-deleted
						uut.deleteItem()
					})
					It("Deletes new item and no error is returned", func() {
						// THEN Both items should be soft-deleted
						uut.expectItemCount(2)
						uut.expectDeletedItemCount(2)

						// AND Error should be nil
						uut.expectNoError()
					})
				})
			})

			When("Allows creating a new item with the same unique value after the previous two were soft-deleted", func() {
				BeforeEach(func() {
					// GIVEN Item with test unique value is created and soft-deleted
					uut.createItem()
					uut.expectNoError()
					uut.deleteItem()
					uut.expectNoError()

					// AND Another item with test unique value is created and soft-deleted
					uut.createItem()
					uut.expectNoError()
					uut.deleteItem()
					uut.expectNoError()
				})
				Context("Another item with test unique value is created", func() {
					BeforeEach(func() {
						// WHEN Another item with test unique value is created
						uut.createItem()
					})
					It("Creates new item and no error is returned", func() {
						// THEN New item should be created
						uut.expectItemCount(3)

						// AND Error should be nil
						uut.expectNoError()

						// AND Database should contain two soft-deleted items with test unique value
						uut.expectDeletedItemCount(2)
					})
				})
			})
		})
	}
})

// -- BASE
type uniquenessTestItemTemplateBase struct {
	lastErr error
	db      *mariadb.SqlDatabase
}

func (utitb *uniquenessTestItemTemplateBase) setup(db *mariadb.SqlDatabase) {
	utitb.db = db
	// utitb.db = dbm.NewTestSchema()
}

func (utitb *uniquenessTestItemTemplateBase) teardown() {
	// dbm.TestTearDown(utitb.db)
}

func (utitb *uniquenessTestItemTemplateBase) expectNoError() {
	Expect(utitb.lastErr).To(BeNil())
}

// -- User
type uniquenessUserTemplate struct {
	uniquenessTestItemTemplateBase
	testUser entity.User
}

func (uut *uniquenessUserTemplate) setup(db *mariadb.SqlDatabase) {
	uut.uniquenessTestItemTemplateBase.setup(db)
	ur := test.NewFakeUser()
	uut.testUser = ur.AsUser()
}

func (uut *uniquenessUserTemplate) createItem() {
	_, uut.lastErr = uut.db.CreateUser(&uut.testUser)
}

func (uut *uniquenessUserTemplate) deleteItem() {
	uut.lastErr = uut.db.DeleteUser(uut.testUser.Id, util.SystemUserId)
}

func (uut *uniquenessUserTemplate) expectItemCount(cnt int64) {
	uut.expectUserCountForUniqueUserId(uut.testUser.UniqueUserID, cnt)
}

func (uut *uniquenessUserTemplate) expectDeletedItemCount(cnt int64) {
	uut.expectDeletedUserCountForUniqueUserId(uut.testUser.UniqueUserID, cnt)
}

func (uut *uniquenessUserTemplate) expectDuplicationError() {
	Expect(uut.lastErr).To(HaveOccurred())
	Expect(uut.lastErr.Error()).To(ContainSubstring("Duplicate entry"))
	Expect(uut.lastErr.Error()).To(ContainSubstring("user_unique_active_unique_user_id"))
}

// -- User non-interface
func (uut *uniquenessUserTemplate) expectUserCountForUniqueUserId(uniqueUserId string, cnt int64) {
	uf := entity.UserFilter{UniqueUserID: []*string{&uniqueUserId}, State: []entity.StateFilterType{entity.Active, entity.Deleted}}
	uut.expectUserCount(&uf, cnt)
}

func (uut *uniquenessUserTemplate) expectDeletedUserCountForUniqueUserId(uniqueUserId string, cnt int64) {
	uf := entity.UserFilter{UniqueUserID: []*string{&uniqueUserId}, State: []entity.StateFilterType{entity.Deleted}}
	uut.expectUserCount(&uf, cnt)
}

func (uut *uniquenessUserTemplate) expectUserCount(uf *entity.UserFilter, cnt int64) {
	userCount, err := uut.db.CountUsers(uf)
	Expect(err).To(BeNil())
	Expect(userCount).To(BeEquivalentTo(cnt))
}

// -- Component
type uniquenessComponentTemplate struct {
	uniquenessTestItemTemplateBase
	testComponent entity.Component
}

func (uct *uniquenessComponentTemplate) setup(db *mariadb.SqlDatabase) {
	uct.uniquenessTestItemTemplateBase.setup(db)
	cr := test.NewFakeComponent()
	uct.testComponent = cr.AsComponent()
}

func (uct *uniquenessComponentTemplate) createItem() {
	_, uct.lastErr = uct.db.CreateComponent(&uct.testComponent)
}

func (uct *uniquenessComponentTemplate) deleteItem() {
	uct.lastErr = uct.db.DeleteComponent(uct.testComponent.Id, util.SystemUserId)
}

func (uct *uniquenessComponentTemplate) expectItemCount(cnt int64) {
	uct.expectComponentCountForCCRN(uct.testComponent.CCRN, cnt)
}

func (uct *uniquenessComponentTemplate) expectDeletedItemCount(cnt int64) {
	uct.expectDeletedComponentCountForCCRN(uct.testComponent.CCRN, cnt)
}

func (uct *uniquenessComponentTemplate) expectDuplicationError() {
	Expect(uct.lastErr).To(HaveOccurred())
	Expect(uct.lastErr.Error()).To(ContainSubstring("Duplicate entry"))
	Expect(uct.lastErr.Error()).To(ContainSubstring("component_unique_active_ccrn"))
}

// -- Component non-interface
func (uct *uniquenessComponentTemplate) expectComponentCountForCCRN(ccrn string, cnt int64) {
	cf := entity.ComponentFilter{CCRN: []*string{&ccrn}, State: []entity.StateFilterType{entity.Active, entity.Deleted}}
	uct.expectComponentCount(&cf, cnt)
}

func (uct *uniquenessComponentTemplate) expectDeletedComponentCountForCCRN(ccrn string, cnt int64) {
	cf := entity.ComponentFilter{CCRN: []*string{&ccrn}, State: []entity.StateFilterType{entity.Deleted}}
	uct.expectComponentCount(&cf, cnt)
}

func (uct *uniquenessComponentTemplate) expectComponentCount(cf *entity.ComponentFilter, cnt int64) {
	componentCount, err := uct.db.CountComponents(cf)
	Expect(err).To(BeNil())
	Expect(componentCount).To(BeEquivalentTo(cnt))
}

// -- Component Version
type uniquenessComponentVersionTemplate struct {
	uniquenessTestItemTemplateBase
	testComponentVersion entity.ComponentVersion
}

func (ucvt *uniquenessComponentVersionTemplate) setup(db *mariadb.SqlDatabase) {
	ucvt.uniquenessTestItemTemplateBase.setup(db)

	seeder, err := test.NewDatabaseSeeder(dbm.DbConfig())
	Expect(err).To(BeNil())
	components := seeder.SeedComponents(1)

	cvr := test.NewFakeComponentVersion()
	cvr.ComponentId = components[0].Id

	ucvt.testComponentVersion = cvr.AsComponentVersion()
}

func (ucvt *uniquenessComponentVersionTemplate) createItem() {
	_, ucvt.lastErr = ucvt.db.CreateComponentVersion(&ucvt.testComponentVersion)
}

func (ucvt *uniquenessComponentVersionTemplate) deleteItem() {
	ucvt.lastErr = ucvt.db.DeleteComponentVersion(ucvt.testComponentVersion.Id, util.SystemUserId)
}

func (ucvt *uniquenessComponentVersionTemplate) expectItemCount(cnt int64) {
	ucvt.expectComponentVersionCountForVersionAndComponentId(ucvt.testComponentVersion.Version, ucvt.testComponentVersion.ComponentId, cnt)
}

func (ucvt *uniquenessComponentVersionTemplate) expectDeletedItemCount(cnt int64) {
	ucvt.expectDeletedComponentVersionCountForVersionAndComponentId(ucvt.testComponentVersion.Version, ucvt.testComponentVersion.ComponentId, cnt)
}

func (ucvt *uniquenessComponentVersionTemplate) expectDuplicationError() {
	Expect(ucvt.lastErr).To(HaveOccurred())
	Expect(ucvt.lastErr.Error()).To(ContainSubstring("Database entry already exist"))
}

// -- Component Version non-interface
func (ucvt *uniquenessComponentVersionTemplate) expectComponentVersionCountForVersionAndComponentId(version string, componentId int64, cnt int64) {
	cvf := entity.ComponentVersionFilter{
		Version:     []*string{&version},
		ComponentId: []*int64{&componentId},
		State:       []entity.StateFilterType{entity.Active, entity.Deleted},
	}
	ucvt.expectComponentVersionCount(&cvf, cnt)
}

func (ucvt *uniquenessComponentVersionTemplate) expectDeletedComponentVersionCountForVersionAndComponentId(version string, componentId int64, cnt int64) {
	cvf := entity.ComponentVersionFilter{
		Version:     []*string{&version},
		ComponentId: []*int64{&componentId},
		State:       []entity.StateFilterType{entity.Deleted},
	}
	ucvt.expectComponentVersionCount(&cvf, cnt)
}

func (ucvt *uniquenessComponentVersionTemplate) expectComponentVersionCount(cvf *entity.ComponentVersionFilter, cnt int64) {
	componentVersionCount, err := ucvt.db.CountComponentVersions(cvf)
	Expect(err).To(BeNil())
	Expect(componentVersionCount).To(BeEquivalentTo(cnt))
}

// -- Service
type uniquenessServiceTemplate struct {
	uniquenessTestItemTemplateBase
	testService entity.Service
}

func (ust *uniquenessServiceTemplate) setup(db *mariadb.SqlDatabase) {
	ust.uniquenessTestItemTemplateBase.setup(db)
	sr := test.NewFakeService()
	ust.testService = sr.AsService()
}

func (ust *uniquenessServiceTemplate) createItem() {
	_, ust.lastErr = ust.db.CreateService(&ust.testService)
}

func (ust *uniquenessServiceTemplate) deleteItem() {
	ust.lastErr = ust.db.DeleteService(ust.testService.Id, util.SystemUserId)
}

func (ust *uniquenessServiceTemplate) expectItemCount(cnt int64) {
	ust.expectServiceCountForCCRN(ust.testService.CCRN, cnt)
}

func (ust *uniquenessServiceTemplate) expectDeletedItemCount(cnt int64) {
	ust.expectDeletedServiceCountForCCRN(ust.testService.CCRN, cnt)
}

func (ust *uniquenessServiceTemplate) expectDuplicationError() {
	Expect(ust.lastErr).To(HaveOccurred())
	Expect(ust.lastErr.Error()).To(ContainSubstring("Duplicate entry"))
	Expect(ust.lastErr.Error()).To(ContainSubstring("service_unique_active_ccrn"))
}

// -- Service non-interface
func (ust *uniquenessServiceTemplate) expectServiceCountForCCRN(ccrn string, cnt int64) {
	sf := entity.ServiceFilter{CCRN: []*string{&ccrn}, State: []entity.StateFilterType{entity.Active, entity.Deleted}}
	ust.expectServiceCount(&sf, cnt)
}

func (ust *uniquenessServiceTemplate) expectDeletedServiceCountForCCRN(ccrn string, cnt int64) {
	sf := entity.ServiceFilter{CCRN: []*string{&ccrn}, State: []entity.StateFilterType{entity.Deleted}}
	ust.expectServiceCount(&sf, cnt)
}

func (ust *uniquenessServiceTemplate) expectServiceCount(sf *entity.ServiceFilter, cnt int64) {
	serviceCount, err := ust.db.CountServices(sf)
	Expect(err).To(BeNil())
	Expect(serviceCount).To(BeEquivalentTo(cnt))
}

// -- Component Instance
type uniquenessComponentInstanceTemplate struct {
	uniquenessTestItemTemplateBase
	testComponentInstance entity.ComponentInstance
}

func (ucit *uniquenessComponentInstanceTemplate) setup(db *mariadb.SqlDatabase) {
	ucit.uniquenessTestItemTemplateBase.setup(db)

	seeder, err := test.NewDatabaseSeeder(dbm.DbConfig())
	Expect(err).To(BeNil())
	services := seeder.SeedServices(1)

	cir := test.NewFakeComponentInstance()
	cir.ServiceId = services[0].Id

	ucit.testComponentInstance = cir.AsComponentInstance()
}

func (ucit *uniquenessComponentInstanceTemplate) createItem() {
	_, ucit.lastErr = ucit.db.CreateComponentInstance(&ucit.testComponentInstance)
}

func (ucit *uniquenessComponentInstanceTemplate) deleteItem() {
	ucit.lastErr = ucit.db.DeleteComponentInstance(ucit.testComponentInstance.Id, util.SystemUserId)
}

func (ucit *uniquenessComponentInstanceTemplate) expectItemCount(cnt int64) {
	ucit.expectComponentInstanceCountForCCRNAndServiceId(ucit.testComponentInstance.CCRN, ucit.testComponentInstance.ServiceId, cnt)
}

func (ucit *uniquenessComponentInstanceTemplate) expectDeletedItemCount(cnt int64) {
	ucit.expectDeletedComponentInstanceCountForCCRNAndServiceId(ucit.testComponentInstance.CCRN, ucit.testComponentInstance.ServiceId, cnt)
}

func (ucit *uniquenessComponentInstanceTemplate) expectDuplicationError() {
	Expect(ucit.lastErr).To(HaveOccurred())
	Expect(ucit.lastErr.Error()).To(ContainSubstring("Duplicate entry"))
	Expect(ucit.lastErr.Error()).To(ContainSubstring("componentinstance_unique_active_ccrn"))
}

// -- Component Instance non-interface
func (ucit *uniquenessComponentInstanceTemplate) expectComponentInstanceCountForCCRNAndServiceId(ccrn string, serviceId int64, cnt int64) {
	cif := entity.ComponentInstanceFilter{
		CCRN:      []*string{&ccrn},
		ServiceId: []*int64{&serviceId},
		State:     []entity.StateFilterType{entity.Active, entity.Deleted},
	}
	ucit.expectComponentInstanceCount(&cif, cnt)
}

func (ucit *uniquenessComponentInstanceTemplate) expectDeletedComponentInstanceCountForCCRNAndServiceId(ccrn string, serviceId int64, cnt int64) {
	cif := entity.ComponentInstanceFilter{
		CCRN:      []*string{&ccrn},
		ServiceId: []*int64{&serviceId},
		State:     []entity.StateFilterType{entity.Deleted},
	}
	ucit.expectComponentInstanceCount(&cif, cnt)
}

func (ucit *uniquenessComponentInstanceTemplate) expectComponentInstanceCount(cif *entity.ComponentInstanceFilter, cnt int64) {
	componentInstanceCount, err := ucit.db.CountComponentInstances(cif)
	Expect(err).To(BeNil())
	Expect(componentInstanceCount).To(BeEquivalentTo(cnt))
}

// -- IssueRepository
type uniquenessIssueRepositoryTemplate struct {
	uniquenessTestItemTemplateBase
	testIssueRepository entity.IssueRepository
}

func (uirt *uniquenessIssueRepositoryTemplate) setup(db *mariadb.SqlDatabase) {
	uirt.uniquenessTestItemTemplateBase.setup(db)
	irr := test.NewFakeIssueRepository()
	uirt.testIssueRepository = irr.AsIssueRepository()
}

func (uirt *uniquenessIssueRepositoryTemplate) createItem() {
	_, uirt.lastErr = uirt.db.CreateIssueRepository(&uirt.testIssueRepository)
}

func (uirt *uniquenessIssueRepositoryTemplate) deleteItem() {
	uirt.lastErr = uirt.db.DeleteIssueRepository(uirt.testIssueRepository.Id, util.SystemUserId)
}

func (uirt *uniquenessIssueRepositoryTemplate) expectItemCount(cnt int64) {
	uirt.expectIssueRepositoryCountForName(uirt.testIssueRepository.Name, cnt)
}

func (uirt *uniquenessIssueRepositoryTemplate) expectDeletedItemCount(cnt int64) {
	uirt.expectDeletedIssueRepositoryCountForName(uirt.testIssueRepository.Name, cnt)
}

func (uirt *uniquenessIssueRepositoryTemplate) expectDuplicationError() {
	Expect(uirt.lastErr).To(HaveOccurred())
	Expect(uirt.lastErr.Error()).To(ContainSubstring("Duplicate entry"))
	Expect(uirt.lastErr.Error()).To(ContainSubstring("issuerepository_unique_active_name"))
}

// -- Issue Repository non-interface
func (uirt *uniquenessIssueRepositoryTemplate) expectIssueRepositoryCountForName(name string, cnt int64) {
	irf := entity.IssueRepositoryFilter{Name: []*string{&name}, State: []entity.StateFilterType{entity.Active, entity.Deleted}}
	uirt.expectIssueRepositoryCount(&irf, cnt)
}

func (uirt *uniquenessIssueRepositoryTemplate) expectDeletedIssueRepositoryCountForName(name string, cnt int64) {
	irf := entity.IssueRepositoryFilter{Name: []*string{&name}, State: []entity.StateFilterType{entity.Deleted}}
	uirt.expectIssueRepositoryCount(&irf, cnt)
}

func (uirt *uniquenessIssueRepositoryTemplate) expectIssueRepositoryCount(irf *entity.IssueRepositoryFilter, cnt int64) {
	issueRepositoryCount, err := uirt.db.CountIssueRepositories(irf)
	Expect(err).To(BeNil())
	Expect(issueRepositoryCount).To(BeEquivalentTo(cnt))
}

// -- Issue
type uniquenessIssueTemplate struct {
	uniquenessTestItemTemplateBase
	testIssue entity.Issue
}

func (uit *uniquenessIssueTemplate) setup(db *mariadb.SqlDatabase) {
	uit.uniquenessTestItemTemplateBase.setup(db)
	ir := test.NewFakeIssue()
	uit.testIssue = ir.AsIssue()
}

func (uit *uniquenessIssueTemplate) createItem() {
	_, uit.lastErr = uit.db.CreateIssue(&uit.testIssue)
}

func (uit *uniquenessIssueTemplate) deleteItem() {
	uit.lastErr = uit.db.DeleteIssue(uit.testIssue.Id, util.SystemUserId)
}

func (uit *uniquenessIssueTemplate) expectItemCount(cnt int64) {
	uit.expectIssueCountForPrimaryName(uit.testIssue.PrimaryName, cnt)
}

func (uit *uniquenessIssueTemplate) expectDeletedItemCount(cnt int64) {
	uit.expectDeletedIssueCountForPrimaryName(uit.testIssue.PrimaryName, cnt)
}

func (uit *uniquenessIssueTemplate) expectDuplicationError() {
	Expect(uit.lastErr).To(HaveOccurred())
	Expect(uit.lastErr.Error()).To(ContainSubstring("Duplicate entry"))
	Expect(uit.lastErr.Error()).To(ContainSubstring("issue_unique_active_primary_name"))
}

// -- Issue non-interface
func (uit *uniquenessIssueTemplate) expectIssueCountForPrimaryName(primaryName string, cnt int64) {
	isf := entity.IssueFilter{PrimaryName: []*string{&primaryName}, State: []entity.StateFilterType{entity.Active, entity.Deleted}}
	uit.expectIssueCount(&isf, cnt)
}

func (uit *uniquenessIssueTemplate) expectDeletedIssueCountForPrimaryName(primaryName string, cnt int64) {
	isf := entity.IssueFilter{PrimaryName: []*string{&primaryName}, State: []entity.StateFilterType{entity.Deleted}}
	uit.expectIssueCount(&isf, cnt)
}

func (uit *uniquenessIssueTemplate) expectIssueCount(isf *entity.IssueFilter, cnt int64) {
	issueCount, err := uit.db.CountIssues(isf)
	Expect(err).To(BeNil())
	Expect(issueCount).To(BeEquivalentTo(cnt))
}

// -- Issue Variant
type uniquenessIssueVariantTemplate struct {
	uniquenessTestItemTemplateBase
	testIssueVariant entity.IssueVariant
}

func (uivt *uniquenessIssueVariantTemplate) setup(db *mariadb.SqlDatabase) {
	uivt.uniquenessTestItemTemplateBase.setup(db)

	seeder, err := test.NewDatabaseSeeder(dbm.DbConfig())
	Expect(err).To(BeNil())

	issueRepositories := seeder.SeedIssueRepositories()
	issues := seeder.SeedIssues(1)

	ivr := test.NewFakeIssueVariant(issueRepositories, issues)
	ir := issueRepositories[0].AsIssueRepository()
	uivt.testIssueVariant = ivr.AsIssueVariant(&ir)
}

func (uivt *uniquenessIssueVariantTemplate) createItem() {
	_, uivt.lastErr = uivt.db.CreateIssueVariant(&uivt.testIssueVariant)
}

func (uivt *uniquenessIssueVariantTemplate) deleteItem() {
	uivt.lastErr = uivt.db.DeleteIssueVariant(uivt.testIssueVariant.Id, util.SystemUserId)
}

func (uivt *uniquenessIssueVariantTemplate) expectItemCount(cnt int64) {
	uivt.expectIssueVariantCountForSecondaryName(uivt.testIssueVariant.SecondaryName, cnt)
}

func (uivt *uniquenessIssueVariantTemplate) expectDeletedItemCount(cnt int64) {
	uivt.expectDeletedIssueVariantCountForSecondaryName(uivt.testIssueVariant.SecondaryName, cnt)
}

func (uivt *uniquenessIssueVariantTemplate) expectDuplicationError() {
	Expect(uivt.lastErr).To(HaveOccurred())
	Expect(uivt.lastErr.Error()).To(ContainSubstring("Duplicate entry"))
	Expect(uivt.lastErr.Error()).To(ContainSubstring("issuevariant_unique_active_secondary_name"))
}

// -- Issue Variant non-interface
func (uivt *uniquenessIssueVariantTemplate) expectIssueVariantCountForSecondaryName(secondaryName string, cnt int64) {
	ivf := entity.IssueVariantFilter{SecondaryName: []*string{&secondaryName}, State: []entity.StateFilterType{entity.Active, entity.Deleted}}
	uivt.expectIssueVariantCount(&ivf, cnt)
}

func (uivt *uniquenessIssueVariantTemplate) expectDeletedIssueVariantCountForSecondaryName(secondaryName string, cnt int64) {
	ivf := entity.IssueVariantFilter{SecondaryName: []*string{&secondaryName}, State: []entity.StateFilterType{entity.Deleted}}
	uivt.expectIssueVariantCount(&ivf, cnt)
}

func (uivt *uniquenessIssueVariantTemplate) expectIssueVariantCount(ivf *entity.IssueVariantFilter, cnt int64) {
	issueVariantCount, err := uivt.db.CountIssueVariants(ivf)
	Expect(err).To(BeNil())
	Expect(issueVariantCount).To(BeEquivalentTo(cnt))
}
