// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_match_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	im "github.com/cloudoperators/heureka/internal/app/issue_match"
	"github.com/cloudoperators/heureka/internal/app/issue_repository"
	"github.com/cloudoperators/heureka/internal/app/issue_variant"
	"github.com/cloudoperators/heureka/internal/app/severity"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/cloudoperators/heureka/internal/util"

	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

func TestIssueMatchHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IssueMatch Service Test Suite")
}

var handlerContext common.HandlerContext
var cfg *util.Config

var _ = BeforeSuite(func() {
	cfg = common.GetTestConfig()
	enableLogs := false
	authz := openfga.NewAuthorizationHandler(cfg, enableLogs)
	handlerContext = common.HandlerContext{
		Cache: cache.NewNoCache(),
		Authz: authz,
	}
})

func getIssueMatchFilter() *entity.IssueMatchFilter {
	return &entity.IssueMatchFilter{
		PaginatedX: entity.PaginatedX{
			First: nil,
			After: nil,
		},
		Id:                  nil,
		ServiceCCRN:         nil,
		SeverityValue:       nil,
		Status:              nil,
		IssueId:             nil,
		EvidenceId:          nil,
		ComponentInstanceId: nil,
	}
}

var _ = Describe("When listing IssueMatches", Label("app", "ListIssueMatches"), func() {
	var (
		er                event.EventRegistry
		db                *mocks.MockDatabase
		issueMatchHandler im.IssueMatchHandler
		filter            *entity.IssueMatchFilter
		options           *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		options = entity.NewListOptions()
		filter = getIssueMatchFilter()
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetIssueMatches", filter, []entity.Order{}).Return([]entity.IssueMatchResult{}, nil)
			db.On("CountIssueMatches", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			issueMatchHandler = im.NewIssueMatchHandler(handlerContext, nil)
			res, err := issueMatchHandler.ListIssueMatches(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(*res.TotalCount).Should(BeEquivalentTo(int64(1337)), "return correct Totalcount")
		})
	})

	When("the list option does include the PageInfo", func() {
		BeforeEach(func() {
			options.ShowPageInfo = true
		})
		DescribeTable("pagination information is correct", func(pageSize int, dbElements int, resElements int, hasNextPage bool) {
			filter.First = &pageSize
			matches := []entity.IssueMatchResult{}
			for _, im := range test.NNewFakeIssueMatches(resElements) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithIssueMatch([]entity.Order{}, im))
				matches = append(matches, entity.IssueMatchResult{WithCursor: entity.WithCursor{Value: cursor}, IssueMatch: lo.ToPtr(im)})
			}

			var cursors = lo.Map(matches, func(m entity.IssueMatchResult, _ int) string {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithIssueMatch([]entity.Order{}, *m.IssueMatch))
				return cursor
			})

			var i int64 = 0
			for len(cursors) < dbElements {
				i++
				im := test.NewFakeIssueMatch()
				c, _ := mariadb.EncodeCursor(mariadb.WithIssueMatch([]entity.Order{}, im))
				cursors = append(cursors, c)
			}
			db.On("GetIssueMatches", filter, []entity.Order{}).Return(matches, nil)
			db.On("GetAllIssueMatchCursors", filter, []entity.Order{}).Return(cursors, nil)
			issueMatchHandler = im.NewIssueMatchHandler(handlerContext, nil)
			res, err := issueMatchHandler.ListIssueMatches(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(*res.PageInfo.HasNextPage).To(BeEquivalentTo(hasNextPage), "correct hasNextPage indicator")
			Expect(len(res.Elements)).To(BeEquivalentTo(resElements))
			Expect(len(res.PageInfo.Pages)).To(BeEquivalentTo(int(math.Ceil(float64(dbElements)/float64(pageSize)))), "correct  number of pages")
		},
			Entry("When pageSize is 1 and the database was returning 2 elements", 1, 2, 1, true),
			Entry("When pageSize is 10 and the database was returning 9 elements", 10, 9, 9, false),
			Entry("When pageSize is 10 and the database was returning 11 elements", 10, 11, 10, true),
		)
	})

	When("the list options does NOT include aggregations", func() {

		BeforeEach(func() {
			options.IncludeAggregations = false
		})

		Context("and the given filter does not have any matches in the database", func() {

			BeforeEach(func() {
				db.On("GetIssueMatches", filter, []entity.Order{}).Return([]entity.IssueMatchResult{}, nil)
			})
			It("should return an empty result", func() {

				issueMatchHandler = im.NewIssueMatchHandler(handlerContext, nil)
				res, err := issueMatchHandler.ListIssueMatches(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(0), "return no results")

			})
		})
		Context("and the filter does have results in the database", func() {
			BeforeEach(func() {
				issueMatches := []entity.IssueMatchResult{}
				for _, im := range test.NNewFakeIssueMatches(15) {
					issueMatches = append(issueMatches, entity.IssueMatchResult{IssueMatch: lo.ToPtr(im)})
				}
				db.On("GetIssueMatches", filter, []entity.Order{}).Return(issueMatches, nil)
			})
			It("should return the expected matches in the result", func() {
				issueMatchHandler = im.NewIssueMatchHandler(handlerContext, nil)
				res, err := issueMatchHandler.ListIssueMatches(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(15), "return 15 results")
			})
		})

		Context("and the database operations throw an error", func() {
			BeforeEach(func() {
				db.On("GetIssueMatches", filter, []entity.Order{}).Return([]entity.IssueMatchResult{}, errors.New("some error"))
			})

			It("should return the expected matches in the result", func() {
				issueMatchHandler = im.NewIssueMatchHandler(handlerContext, nil)
				_, err := issueMatchHandler.ListIssueMatches(filter, options)
				Expect(err).Error()
				Expect(err.Error()).ToNot(BeEquivalentTo("some error"), "error gets not passed through")
			})
		})
	})
})

var _ = Describe("When creating IssueMatch", Label("app", "CreateIssueMatch"), func() {
	var (
		er                event.EventRegistry
		db                *mocks.MockDatabase
		issueMatchHandler im.IssueMatchHandler
		issueMatch        entity.IssueMatch
		ivFilter          *entity.IssueVariantFilter
		irFilter          *entity.IssueRepositoryFilter
		issueVariants     []entity.IssueVariant
		repositories      []entity.IssueRepository
		ss                severity.SeverityHandler
		ivs               issue_variant.IssueVariantHandler
		rs                issue_repository.IssueRepositoryHandler
		p                 openfga.PermissionInput
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		issueMatch = test.NewFakeIssueMatch()
		ivFilter = entity.NewIssueVariantFilter()
		irFilter = entity.NewIssueRepositoryFilter()
		first := 10
		ivFilter.First = &first
		var after int64 = 0
		ivFilter.After = &after
		irFilter.First = &first
		irFilter.After = &after

		p = openfga.PermissionInput{
			UserType:   openfga.TypeRole,
			UserId:     "0",
			ObjectId:   "test_issue_match",
			ObjectType: openfga.TypeIssueMatch,
			Relation:   openfga.TypeRole,
		}

		handlerContext.DB = db
		handlerContext.EventReg = er

		rs = issue_repository.NewIssueRepositoryHandler(handlerContext)
		ivs = issue_variant.NewIssueVariantHandler(handlerContext, rs)
		ss = severity.NewSeverityHandler(handlerContext, ivs)
	})

	It("creates issueMatch", func() {
		issueVariants = test.NNewFakeIssueVariants(1)
		repositories = test.NNewFakeIssueRepositories(1)
		issueVariants[0].IssueRepositoryId = repositories[0].Id
		irFilter.Id = []*int64{&repositories[0].Id}
		ivFilter.IssueId = []*int64{&issueMatch.IssueId}
		issueMatch.Severity = issueVariants[0].Severity
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("CreateIssueMatch", &issueMatch).Return(&issueMatch, nil)
		db.On("GetIssueVariants", ivFilter).Return(issueVariants, nil)
		db.On("GetIssueRepositories", irFilter).Return(repositories, nil)
		issueMatchHandler = im.NewIssueMatchHandler(handlerContext, ss)
		newIssueMatch, err := issueMatchHandler.CreateIssueMatch(&issueMatch)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newIssueMatch.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newIssueMatch.TargetRemediationDate.Format(time.RFC3339)).To(BeEquivalentTo(issueMatch.TargetRemediationDate.Format(time.RFC3339)))
			Expect(newIssueMatch.RemediationDate.Format(time.RFC3339)).To(BeEquivalentTo(issueMatch.RemediationDate.Format(time.RFC3339)))
			Expect(newIssueMatch.Status).To(BeEquivalentTo(issueMatch.Status))
			Expect(newIssueMatch.UserId).To(BeEquivalentTo(issueMatch.UserId))
			Expect(newIssueMatch.ComponentInstanceId).To(BeEquivalentTo(issueMatch.ComponentInstanceId))
			Expect(newIssueMatch.IssueId).To(BeEquivalentTo(issueMatch.IssueId))
			Expect(newIssueMatch.Severity.Cvss.Vector).To(BeEquivalentTo(issueMatch.Severity.Cvss.Vector))
			Expect(newIssueMatch.Severity.Score).To(BeEquivalentTo(issueMatch.Severity.Score))
			Expect(newIssueMatch.Severity.Value).To(BeEquivalentTo(issueMatch.Severity.Value))
		})
	})

	Context("when handling a CreateIssueMatchEvent", func() {
		Context("when new issue match is created", func() {
			It("should add user resource relationship tuple in openfga", func() {
				imFake := test.NewFakeIssueMatch()
				createEvent := &im.CreateIssueMatchEvent{
					IssueMatch: &imFake,
				}

				// Use type assertion to convert a CreateServiceEvent into an Event
				var event event.Event = createEvent
				p.ObjectId = openfga.ObjectIdFromInt(createEvent.IssueMatch.Id)
				// Simulate event
				im.OnIssueMatchCreateAuthz(db, event, handlerContext.Authz)

				ok, err := handlerContext.Authz.CheckPermission(p)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(ok).To(BeTrue(), "permission should be granted")
			})
		})
	})
})

var _ = Describe("When updating IssueMatch", Label("app", "UpdateIssueMatch"), func() {
	var (
		er                event.EventRegistry
		db                *mocks.MockDatabase
		issueMatchHandler im.IssueMatchHandler
		issueMatch        entity.IssueMatchResult
		filter            *entity.IssueMatchFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)

		issueMatch = test.NewFakeIssueMatchResult()
		first := 10
		after := ""
		filter = &entity.IssueMatchFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("updates issueMatch", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateIssueMatch", issueMatch.IssueMatch).Return(nil)
		issueMatchHandler = im.NewIssueMatchHandler(handlerContext, nil)
		if issueMatch.Status == entity.NewIssueMatchStatusValue("new") {
			issueMatch.Status = entity.NewIssueMatchStatusValue("risk_accepted")
		} else {
			issueMatch.Status = entity.NewIssueMatchStatusValue("new")
		}
		filter.Id = []*int64{&issueMatch.Id}
		db.On("GetIssueMatches", filter, []entity.Order{}).Return([]entity.IssueMatchResult{issueMatch}, nil)
		updatedIssueMatch, err := issueMatchHandler.UpdateIssueMatch(issueMatch.IssueMatch)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedIssueMatch.TargetRemediationDate).To(BeEquivalentTo(issueMatch.TargetRemediationDate))
			Expect(updatedIssueMatch.RemediationDate).To(BeEquivalentTo(issueMatch.RemediationDate))
			Expect(updatedIssueMatch.Status).To(BeEquivalentTo(issueMatch.Status))
			Expect(updatedIssueMatch.UserId).To(BeEquivalentTo(issueMatch.UserId))
			Expect(updatedIssueMatch.ComponentInstanceId).To(BeEquivalentTo(issueMatch.ComponentInstanceId))
			Expect(updatedIssueMatch.IssueId).To(BeEquivalentTo(issueMatch.IssueId))
			Expect(updatedIssueMatch.Severity.Cvss.Vector).To(BeEquivalentTo(issueMatch.Severity.Cvss.Vector))
			Expect(updatedIssueMatch.Severity.Score).To(BeEquivalentTo(issueMatch.Severity.Score))
			Expect(updatedIssueMatch.Severity.Value).To(BeEquivalentTo(issueMatch.Severity.Value))
		})
	})

	Context("when handling an UpdateIssueMatchEvent", func() {
		It("should update the component_instance relation tuple in openfga", func() {
			imFake := test.NewFakeIssueMatch()
			oldComponentInstanceId := int64(12345)
			newComponentInstanceId := int64(67890)

			// Add an initial relation: issue_match -> old component_instance
			initialRelation := openfga.RelationInput{
				UserType:   openfga.TypeComponentInstance,
				UserId:     openfga.UserIdFromInt(oldComponentInstanceId),
				Relation:   openfga.RelComponentInstance,
				ObjectType: openfga.TypeIssueMatch,
				ObjectId:   openfga.ObjectIdFromInt(imFake.Id),
			}

			handlerContext.Authz.AddRelationBulk([]openfga.RelationInput{initialRelation})

			// Prepare the update event with the new component_instance id
			imFake.ComponentInstanceId = newComponentInstanceId
			updateEvent := &im.UpdateIssueMatchEvent{
				IssueMatch: &imFake,
			}
			var event event.Event = updateEvent

			// Simulate event
			im.OnIssueMatchUpdateAuthz(db, event, handlerContext.Authz)

			// Check that the old relation is gone
			remainingOld, err := handlerContext.Authz.ListRelations(initialRelation)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(remainingOld).To(BeEmpty(), "old relation should be removed")

			// Check that the new relation exists
			newRelation := openfga.RelationInput{
				UserType:   openfga.TypeComponentInstance,
				UserId:     openfga.UserIdFromInt(newComponentInstanceId),
				Relation:   openfga.RelComponentInstance,
				ObjectType: openfga.TypeIssueMatch,
				ObjectId:   openfga.ObjectIdFromInt(imFake.Id),
			}
			remainingNew, err := handlerContext.Authz.ListRelations(newRelation)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(remainingNew).NotTo(BeEmpty(), "new relation should exist")
		})
	})
})

var _ = Describe("When deleting IssueMatch", Label("app", "DeleteIssueMatch"), func() {
	var (
		er                event.EventRegistry
		db                *mocks.MockDatabase
		issueMatchHandler im.IssueMatchHandler
		id                int64
		filter            *entity.IssueMatchFilter
		options           *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)

		id = 1
		first := 10
		after := ""
		filter = &entity.IssueMatchFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		options = entity.NewListOptions()
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("deletes issueMatch", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("DeleteIssueMatch", id, mock.Anything).Return(nil)
		issueMatchHandler = im.NewIssueMatchHandler(handlerContext, nil)
		db.On("GetIssueMatches", filter, []entity.Order{}).Return([]entity.IssueMatchResult{}, nil)
		err := issueMatchHandler.DeleteIssueMatch(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		issueMatches, err := issueMatchHandler.ListIssueMatches(filter, options)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issueMatches.Elements).To(BeEmpty(), "no error should be thrown")
	})

	Context("when handling a DeleteIssueMatchEvent", func() {
		Context("when new issue match is deleted", func() {
			It("should delete tuples related to that issuematch in openfga", func() {
				// Test OnIssueMatchDeleteAuthz against all possible relations
				imFake := test.NewFakeIssueMatch()
				deleteEvent := &im.DeleteIssueMatchEvent{
					IssueMatchID: imFake.Id,
				}
				objectId := openfga.ObjectIdFromInt(deleteEvent.IssueMatchID)
				relations := []openfga.RelationInput{
					{ // user - issue_match: a user can view the issue match
						UserType:   openfga.TypeUser,
						UserId:     openfga.IDUser,
						ObjectId:   objectId,
						ObjectType: openfga.TypeIssueMatch,
						Relation:   openfga.RelCanView,
					},
					{ // component_instance - issue_match: a component instance is related to the issue match
						UserType:   openfga.TypeComponentInstance,
						UserId:     openfga.IDComponentInstance,
						ObjectId:   objectId,
						ObjectType: openfga.TypeIssueMatch,
						Relation:   openfga.RelComponentInstance,
					},
					{ // role - issue_match: a role is assigned to the issue match
						UserType:   openfga.TypeRole,
						UserId:     openfga.IDRole,
						ObjectId:   objectId,
						ObjectType: openfga.TypeIssueMatch,
						Relation:   openfga.RelRole,
					},
				}

				handlerContext.Authz.AddRelationBulk(relations)

				// get the number of relations before deletion
				relCountBefore := 0
				for _, r := range relations {
					relations, err := handlerContext.Authz.ListRelations(r)
					if err != nil {
						Expect(err).To(BeNil(), "no error should be thrown")
					}
					relCountBefore += len(relations)
				}
				Expect(relCountBefore).To(Equal(len(relations)), "all relations should exist before deletion")

				// check that relations were created
				for _, r := range relations {
					ok, err := handlerContext.Authz.CheckPermission(openfga.PermissionInput{
						UserType:   r.UserType,
						UserId:     r.UserId,
						ObjectType: r.ObjectType,
						ObjectId:   r.ObjectId,
						Relation:   r.Relation,
					})
					Expect(err).To(BeNil(), "no error should be thrown")
					Expect(ok).To(BeTrue(), "permission should be granted")
				}

				var event event.Event = deleteEvent
				// Simulate event
				im.OnIssueMatchDeleteAuthz(db, event, handlerContext.Authz)

				// get the number of relations after deletion
				relCountAfter := 0
				for _, r := range relations {
					relations, err := handlerContext.Authz.ListRelations(r)
					if err != nil {
						Expect(err).To(BeNil(), "no error should be thrown")
					}
					relCountAfter += len(relations)
				}
				Expect(relCountAfter < relCountBefore).To(BeTrue(), "less relations after deletion")
				Expect(relCountAfter).To(BeEquivalentTo(0), "no relations should exist after deletion")

				// verify that relations were deleted
				for _, r := range relations {
					ok, err := handlerContext.Authz.CheckPermission(openfga.PermissionInput{
						UserType:   r.UserType,
						UserId:     r.UserId,
						ObjectType: r.ObjectType,
						ObjectId:   r.ObjectId,
						Relation:   r.Relation,
					})
					Expect(err).To(BeNil(), "no error should be thrown")
					Expect(ok).To(BeFalse(), "permission should NOT be granted")
				}
			})
		})
	})
})

var _ = Describe("When modifying relationship of evidence and issueMatch", Label("app", "EvidenceIssueMatchRelationship"), func() {
	var (
		er                event.EventRegistry
		db                *mocks.MockDatabase
		issueMatchHandler im.IssueMatchHandler
		evidence          entity.Evidence
		issueMatch        entity.IssueMatchResult
		filter            *entity.IssueMatchFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)

		issueMatch = test.NewFakeIssueMatchResult()
		evidence = test.NewFakeEvidenceEntity()
		first := 10
		after := ""
		filter = &entity.IssueMatchFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
			Id: []*int64{&issueMatch.Id},
		}
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("adds evidence to issueMatch", func() {
		db.On("AddEvidenceToIssueMatch", issueMatch.Id, evidence.Id).Return(nil)
		db.On("GetIssueMatches", filter, []entity.Order{}).Return([]entity.IssueMatchResult{issueMatch}, nil)
		issueMatchHandler = im.NewIssueMatchHandler(handlerContext, nil)
		issueMatch, err := issueMatchHandler.AddEvidenceToIssueMatch(issueMatch.Id, evidence.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issueMatch).NotTo(BeNil(), "issueMatch should be returned")
	})

	It("removes evidence from issueMatch", func() {
		db.On("RemoveEvidenceFromIssueMatch", issueMatch.Id, evidence.Id).Return(nil)
		db.On("GetIssueMatches", filter, []entity.Order{}).Return([]entity.IssueMatchResult{issueMatch}, nil)
		issueMatchHandler = im.NewIssueMatchHandler(handlerContext, nil)
		issueMatch, err := issueMatchHandler.RemoveEvidenceFromIssueMatch(issueMatch.Id, evidence.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issueMatch).NotTo(BeNil(), "issueMatch should be returned")
	})
})

var _ = Describe("OnComponentInstanceCreate", Label("app", "OnComponentInstanceCreate"), func() {
	var (
		db                  *mocks.MockDatabase
		componentInstanceID int64
		componentVersionID  int64
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		componentInstanceID = 1
		componentVersionID = 1
	})

	// Tests for BuildIssueVariantMap
	Context("BuildIssueVariantMap", func() {
		When("no issues are found", func() {
			BeforeEach(func() {
				// Fake IssueRepository
				ir := test.NewFakeIssueRepositoryEntity()
				ir.Id = 1
				ir.Priority = 1

				// Fake Service
				service := test.NewFakeServiceEntity()
				service.Id = 1

				// Mocks
				db.On("GetServiceIssueVariants", &entity.ServiceIssueVariantFilter{
					ComponentInstanceId: []*int64{lo.ToPtr(int64(1))},
				}).Return([]entity.ServiceIssueVariant{}, nil)
			})

			It("should return an empty map", func() {
				result, err := im.BuildIssueVariantMap(db, componentInstanceID, componentVersionID)

				Expect(err).NotTo(BeNil())
				Expect(result).To(BeEmpty())
			})
		})

		When("all data is retrieved successfully", func() {
			BeforeEach(func() {
				variants := test.NNewFakeServiceIssueVariantEntity(2, 10, nil)
				// Mocks
				db.On("GetServiceIssueVariants", mock.MatchedBy(func(filter *entity.ServiceIssueVariantFilter) bool {
					// Check that IssueId and IssueRepositoryId are not nil, but don't care about their contents
					return filter.ComponentInstanceId != nil
				})).Return(variants, nil)
			})

			It("should return the correct issue variant map", func() {
				result, err := im.BuildIssueVariantMap(db, componentInstanceID, componentVersionID)

				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(2))
			})
		})

		When("multiple issue repository with different priority", func() {
			var v2 entity.ServiceIssueVariant
			BeforeEach(func() {
				v1 := test.NewFakeServiceIssueVariantEntity(100, lo.ToPtr(int64(1)))
				v2 = test.NewFakeServiceIssueVariantEntity(200, lo.ToPtr(int64(1)))
				variants := []entity.ServiceIssueVariant{v1, v2}
				// Mocks
				db.On("GetServiceIssueVariants", mock.MatchedBy(func(filter *entity.ServiceIssueVariantFilter) bool {
					// Check that IssueId and IssueRepositoryId are not nil, but don't care about their contents
					return filter.ComponentInstanceId != nil
				})).Return(variants, nil)
			})
			It("it should chose the issue repository with the highest priority", func() {
				result, err := im.BuildIssueVariantMap(db, componentInstanceID, componentVersionID)

				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))
				Expect(result).To(HaveKey(int64(1)))
				Expect(result[1].Id).To(BeEquivalentTo(v2.Id))
			})
		})

		When("multiple issue repository with same priority", func() {

			BeforeEach(func() {
				variants := test.NNewFakeServiceIssueVariantEntity(2, 10, lo.ToPtr(int64(1)))
				// Mocks
				db.On("GetServiceIssueVariants", mock.MatchedBy(func(filter *entity.ServiceIssueVariantFilter) bool {
					// Check that IssueId and IssueRepositoryId are not nil, but don't care about their contents
					return filter.ComponentInstanceId != nil
				})).Return(variants, nil)
			})
			It("it should randomly chose one issue repository", func() {
				result, err := im.BuildIssueVariantMap(db, componentInstanceID, componentVersionID)

				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))
				Expect(result).To(HaveKey(int64(1)))

				iv, ok := result[1]
				Expect(ok).To(BeTrue())
				Expect(iv).To(BeAssignableToTypeOf(entity.ServiceIssueVariant{}))

			})
		})
	})

	// Tests for OnComponentVersionAssignmentToComponentInstance
	Context("OnComponentVersionAssignmentToComponentInstance", func() {
		Context("when BuildIssueVariantMap succeeds", func() {
			BeforeEach(func() {
				v1 := test.NewFakeServiceIssueVariantEntity(100, lo.ToPtr(int64(1)))
				v2 := test.NewFakeServiceIssueVariantEntity(200, lo.ToPtr(int64(2)))
				variants := []entity.ServiceIssueVariant{v1, v2}
				// Mocks
				db.On("GetServiceIssueVariants", mock.MatchedBy(func(filter *entity.ServiceIssueVariantFilter) bool {
					// Check that IssueId and IssueRepositoryId are not nil, but don't care about their contents
					return filter.ComponentInstanceId != nil
				})).Return(variants, nil)
			})

			It("should create issue matches for each issue", func() {
				db.On("GetIssueMatches", mock.Anything, mock.Anything).Return([]entity.IssueMatchResult{}, nil)
				// Mock CreateIssueMatch
				db.On("CreateIssueMatch", mock.AnythingOfType("*entity.IssueMatch")).Return(&entity.IssueMatch{}, nil).Twice()
				im.OnComponentVersionAssignmentToComponentInstance(db, componentInstanceID, componentVersionID)

				// Verify that CreateIssueMatch was called twice (once for each issue)
				db.AssertNumberOfCalls(GinkgoT(), "CreateIssueMatch", 2)
			})

			Context("when issue matches already exist", func() {
				BeforeEach(func() {
					// Fake issues
					issueMatch := test.NewFakeIssueMatchResult()
					issueMatch.IssueId = 2 // issue2.Id
					//when issueid is 2 return a fake issue match
					db.On("GetIssueMatches", mock.Anything, mock.Anything).Return([]entity.IssueMatchResult{issueMatch}, nil).Once()
				})

				It("should should not create new issues", func() {
					// Mock CreateIssueMatch
					im.OnComponentVersionAssignmentToComponentInstance(db, componentInstanceID, componentVersionID)

					// Verify that CreateIssueMatch was called only once (for the new issue)
					db.AssertNumberOfCalls(GinkgoT(), "CreateIssueMatch", 0)
				})
			})

		})
	})
})
