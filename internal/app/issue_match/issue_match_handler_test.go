// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_match_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/cloudoperators/heureka/internal/app/event"
	im "github.com/cloudoperators/heureka/internal/app/issue_match"
	"github.com/cloudoperators/heureka/internal/app/issue_repository"
	"github.com/cloudoperators/heureka/internal/app/issue_variant"
	"github.com/cloudoperators/heureka/internal/app/severity"

	"github.com/samber/lo"

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

var er event.EventRegistry

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)
})

func getIssueMatchFilter() *entity.IssueMatchFilter {
	return &entity.IssueMatchFilter{
		Paginated: entity.Paginated{
			First: nil,
			After: nil,
		},
		Id:                  nil,
		AffectedServiceCCRN: nil,
		SeverityValue:       nil,
		Status:              nil,
		IssueId:             nil,
		EvidenceId:          nil,
		ComponentInstanceId: nil,
	}
}

var _ = Describe("When listing IssueMatches", Label("app", "ListIssueMatches"), func() {
	var (
		db                *mocks.MockDatabase
		issueMatchHandler im.IssueMatchHandler
		filter            *entity.IssueMatchFilter
		options           *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getIssueMatchFilter()
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetIssueMatches", filter).Return([]entity.IssueMatch{}, nil)
			db.On("CountIssueMatches", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			issueMatchHandler = im.NewIssueMatchHandler(db, er, nil)
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
			matches := test.NNewFakeIssueMatches(resElements)

			var ids = lo.Map(matches, func(m entity.IssueMatch, _ int) int64 { return m.Id })
			var i int64 = 0
			for len(ids) < dbElements {
				i++
				ids = append(ids, i)
			}
			db.On("GetIssueMatches", filter).Return(matches, nil)
			db.On("GetAllIssueMatchIds", filter).Return(ids, nil)
			issueMatchHandler = im.NewIssueMatchHandler(db, er, nil)
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
				db.On("GetIssueMatches", filter).Return([]entity.IssueMatch{}, nil)
			})
			It("should return an empty result", func() {

				issueMatchHandler = im.NewIssueMatchHandler(db, er, nil)
				res, err := issueMatchHandler.ListIssueMatches(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(0), "return no results")

			})
		})
		Context("and the filter does have results in the database", func() {
			BeforeEach(func() {
				db.On("GetIssueMatches", filter).Return(test.NNewFakeIssueMatches(15), nil)
			})
			It("should return the expected matches in the result", func() {
				issueMatchHandler = im.NewIssueMatchHandler(db, er, nil)
				res, err := issueMatchHandler.ListIssueMatches(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(15), "return 15 results")
			})
		})

		Context("and the database operations throw an error", func() {
			BeforeEach(func() {
				db.On("GetIssueMatches", filter).Return([]entity.IssueMatch{}, errors.New("some error"))
			})

			It("should return the expected matches in the result", func() {
				issueMatchHandler = im.NewIssueMatchHandler(db, er, nil)
				_, err := issueMatchHandler.ListIssueMatches(filter, options)
				Expect(err).Error()
				Expect(err.Error()).ToNot(BeEquivalentTo("some error"), "error gets not passed through")
			})
		})
	})
})

var _ = Describe("When creating IssueMatch", Label("app", "CreateIssueMatch"), func() {
	var (
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
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueMatch = test.NewFakeIssueMatch()
		ivFilter = entity.NewIssueVariantFilter()
		irFilter = entity.NewIssueRepositoryFilter()
		first := 10
		ivFilter.First = &first
		var after int64 = 0
		ivFilter.After = &after
		irFilter.First = &first
		irFilter.After = &after
		rs = issue_repository.NewIssueRepositoryHandler(db, er)
		ivs = issue_variant.NewIssueVariantHandler(db, er, rs)
		ss = severity.NewSeverityHandler(db, er, ivs)
	})

	It("creates issueMatch", func() {
		issueVariants = test.NNewFakeIssueVariants(1)
		repositories = test.NNewFakeIssueRepositories(1)
		issueVariants[0].IssueRepositoryId = repositories[0].Id
		irFilter.Id = []*int64{&repositories[0].Id}
		ivFilter.IssueId = []*int64{&issueMatch.IssueId}
		issueMatch.Severity = issueVariants[0].Severity
		db.On("CreateIssueMatch", &issueMatch).Return(&issueMatch, nil)
		db.On("GetIssueVariants", ivFilter).Return(issueVariants, nil)
		db.On("GetIssueRepositories", irFilter).Return(repositories, nil)
		issueMatchHandler = im.NewIssueMatchHandler(db, er, ss)
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
})

var _ = Describe("When updating IssueMatch", Label("app", "UpdateIssueMatch"), func() {
	var (
		db                *mocks.MockDatabase
		issueMatchHandler im.IssueMatchHandler
		issueMatch        entity.IssueMatch
		filter            *entity.IssueMatchFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueMatch = test.NewFakeIssueMatch()
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueMatchFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("updates issueMatch", func() {
		db.On("UpdateIssueMatch", &issueMatch).Return(nil)
		issueMatchHandler = im.NewIssueMatchHandler(db, er, nil)
		if issueMatch.Status == entity.NewIssueMatchStatusValue("new") {
			issueMatch.Status = entity.NewIssueMatchStatusValue("risk_accepted")
		} else {
			issueMatch.Status = entity.NewIssueMatchStatusValue("new")
		}
		filter.Id = []*int64{&issueMatch.Id}
		db.On("GetIssueMatches", filter).Return([]entity.IssueMatch{issueMatch}, nil)
		updatedIssueMatch, err := issueMatchHandler.UpdateIssueMatch(&issueMatch)
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
})

var _ = Describe("When deleting IssueMatch", Label("app", "DeleteIssueMatch"), func() {
	var (
		db                *mocks.MockDatabase
		issueMatchHandler im.IssueMatchHandler
		id                int64
		filter            *entity.IssueMatchFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueMatchFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("deletes issueMatch", func() {
		db.On("DeleteIssueMatch", id).Return(nil)
		issueMatchHandler = im.NewIssueMatchHandler(db, er, nil)
		db.On("GetIssueMatches", filter).Return([]entity.IssueMatch{}, nil)
		err := issueMatchHandler.DeleteIssueMatch(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		issueMatches, err := issueMatchHandler.ListIssueMatches(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issueMatches.Elements).To(BeEmpty(), "no error should be thrown")
	})
})

var _ = Describe("When modifying relationship of evidence and issueMatch", Label("app", "EvidenceIssueMatchRelationship"), func() {
	var (
		db                *mocks.MockDatabase
		issueMatchHandler im.IssueMatchHandler
		evidence          entity.Evidence
		issueMatch        entity.IssueMatch
		filter            *entity.IssueMatchFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueMatch = test.NewFakeIssueMatch()
		evidence = test.NewFakeEvidenceEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueMatchFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
			Id: []*int64{&issueMatch.Id},
		}
	})

	It("adds evidence to issueMatch", func() {
		db.On("AddEvidenceToIssueMatch", issueMatch.Id, evidence.Id).Return(nil)
		db.On("GetIssueMatches", filter).Return([]entity.IssueMatch{issueMatch}, nil)
		issueMatchHandler = im.NewIssueMatchHandler(db, er, nil)
		issueMatch, err := issueMatchHandler.AddEvidenceToIssueMatch(issueMatch.Id, evidence.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issueMatch).NotTo(BeNil(), "issueMatch should be returned")
	})

	It("removes evidence from issueMatch", func() {
		db.On("RemoveEvidenceFromIssueMatch", issueMatch.Id, evidence.Id).Return(nil)
		db.On("GetIssueMatches", filter).Return([]entity.IssueMatch{issueMatch}, nil)
		issueMatchHandler = im.NewIssueMatchHandler(db, er, nil)
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
				db.On("GetIssueMatches", mock.Anything).Return([]entity.IssueMatch{}, nil)
				// Mock CreateIssueMatch
				db.On("CreateIssueMatch", mock.AnythingOfType("*entity.IssueMatch")).Return(&entity.IssueMatch{}, nil).Twice()
				im.OnComponentVersionAssignmentToComponentInstance(db, componentInstanceID, componentVersionID)

				// Verify that CreateIssueMatch was called twice (once for each issue)
				db.AssertNumberOfCalls(GinkgoT(), "CreateIssueMatch", 2)
			})

			Context("when issue matches already exist", func() {
				BeforeEach(func() {
					// Fake issues
					issueMatch := test.NewFakeIssueMatch()
					issueMatch.IssueId = 2 // issue2.Id
					//when issueid is 2 return a fake issue match
					db.On("GetIssueMatches", mock.Anything).Return([]entity.IssueMatch{issueMatch}, nil).Once()
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
