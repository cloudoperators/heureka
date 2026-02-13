// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0
package issue_test

import (
	"errors"
	"math"
	"strconv"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/app/issue"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mock "github.com/stretchr/testify/mock"
)

func TestIssueHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Issue Service Test Suite")
}

var (
	er    event.EventRegistry
	authz openfga.Authorization
)

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)
})

var _ = Describe("When getting a single Issue", Label("app", "GetIssue", "errors"), func() {
	var (
		db             *mocks.MockDatabase
		issueHandler   issue.IssueHandler
		issueEntity    entity.Issue
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
		issueHandler = issue.NewIssueHandler(handlerContext)
		issueEntity = test.NewFakeIssueEntity()
	})

	Context("with valid input", func() {
		It("should return the issue when it exists", func() {
			// Setup mock to return one issue
			expectedResult := []entity.IssueResult{{
				Issue: &issueEntity,
			}}
			db.On("GetIssues", mock.MatchedBy(func(filter *entity.IssueFilter) bool {
				return len(filter.Id) == 1 && *filter.Id[0] == issueEntity.Id
			}), []entity.Order{}).Return(expectedResult, nil)

			result, err := issueHandler.GetIssue(issueEntity.Id)

			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(result).ToNot(BeNil(), "issue should be returned")
			Expect(result.Id).To(Equal(issueEntity.Id), "correct issue ID")
			Expect(result.PrimaryName).To(Equal(issueEntity.PrimaryName), "correct primary name")
		})
	})

	Context("with invalid input", func() {
		It("should return InvalidArgument error for negative ID", func() {
			result, err := issueHandler.GetIssue(-1)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			// Verify it's our structured error with correct code
			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.InvalidArgument), "should be InvalidArgument error")
			Expect(appErr.Entity).To(Equal("Issue"), "should reference Issue entity")
			Expect(appErr.Op).To(Equal("issueHandler.GetIssue"), "should include operation")
		})

		It("should return InvalidArgument error for zero ID", func() {
			result, err := issueHandler.GetIssue(0)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.InvalidArgument), "should be InvalidArgument error")
		})
	})

	Context("when issue does not exist", func() {
		It("should return NotFound error", func() {
			// Setup mock to return empty result (no issues found)
			db.On("GetIssues", mock.Anything, []entity.Order{}).Return([]entity.IssueResult{}, nil)

			result, err := issueHandler.GetIssue(999)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.NotFound), "should be NotFound error")
			Expect(appErr.Entity).To(Equal("Issue"), "should reference Issue entity")
			Expect(appErr.ID).To(Equal("999"), "should include issue ID")
		})
	})

	Context("when database error occurs", func() {
		It("should return Internal error wrapping the database error", func() {
			// Setup mock to return database error
			dbError := errors.New("database connection failed")
			db.On("GetIssues", mock.Anything, []entity.Order{}).Return([]entity.IssueResult{}, dbError)

			result, err := issueHandler.GetIssue(123)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("Issue"), "should reference Issue entity")
			Expect(appErr.ID).To(Equal("123"), "should include issue ID")
		})
	})
})

func getIssueFilter() *entity.IssueFilter {
	return &entity.IssueFilter{
		PaginatedX: entity.PaginatedX{
			First: nil,
			After: nil,
		},
	}
}

func getIssueListOptions() *entity.IssueListOptions {
	listOptions := entity.NewListOptions()
	return &entity.IssueListOptions{
		ListOptions:         *listOptions,
		ShowIssueTypeCounts: false,
	}
}

func getIssueTypeCounts() *entity.IssueTypeCounts {
	return &entity.IssueTypeCounts{
		VulnerabilityCount:   1000,
		PolicyViolationCount: 300,
		SecurityEventCount:   37,
	}
}

var _ = Describe("When listing Issues", Label("app", "ListIssues"), func() {
	var (
		db              *mocks.MockDatabase
		issueHandler    issue.IssueHandler
		filter          *entity.IssueFilter
		options         *entity.IssueListOptions
		issueTypeCounts *entity.IssueTypeCounts
		handlerContext  common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = getIssueListOptions()
		filter = getIssueFilter()
		issueTypeCounts = getIssueTypeCounts()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	When("the list option does include the totalCount", func() {
		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetIssues", filter, []entity.Order{}).Return([]entity.IssueResult{}, nil)
			db.On("CountIssueTypes", filter).Return(issueTypeCounts, nil)
		})

		It("shows the total count in the results", func() {
			issueHandler = issue.NewIssueHandler(handlerContext)
			res, err := issueHandler.ListIssues(filter, options)
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
			issues := []entity.IssueResult{}
			for _, i := range test.NNewFakeIssueEntities(resElements) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithIssue([]entity.Order{}, i, 0))
				issues = append(issues, entity.IssueResult{WithCursor: entity.WithCursor{Value: cursor}, Issue: lo.ToPtr(i)})
			}

			cursors := lo.Map(issues, func(ir entity.IssueResult, _ int) string {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithIssue([]entity.Order{}, *ir.Issue, 0))
				return cursor
			})

			var i int64 = 0
			for len(cursors) < dbElements {
				i++
				issue := test.NewFakeIssueEntity()
				c, _ := mariadb.EncodeCursor(mariadb.WithIssue([]entity.Order{}, issue, 0))
				cursors = append(cursors, c)
			}
			db.On("GetIssues", filter, []entity.Order{}).Return(issues, nil)
			db.On("GetAllIssueCursors", filter, []entity.Order{}).Return(cursors, nil)
			db.On("CountIssueTypes", filter).Return(issueTypeCounts, nil)
			issueHandler = issue.NewIssueHandler(handlerContext)
			res, err := issueHandler.ListIssues(filter, options)
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

	When("the list options does include aggregations", func() {
		BeforeEach(func() {
			options.IncludeAggregations = true
		})
		Context("and the given filter does not have any matches in the database", func() {
			BeforeEach(func() {
				db.On("GetIssuesWithAggregations", filter, []entity.Order{}).Return([]entity.IssueResult{}, nil)
			})

			It("should return an empty result", func() {
				issueHandler = issue.NewIssueHandler(handlerContext)
				res, err := issueHandler.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(0), "return no results")
			})
		})
		Context("and the filter does have results in the database", func() {
			BeforeEach(func() {
				db.On("GetIssuesWithAggregations", filter, []entity.Order{}).Return(test.NNewFakeIssueResultsWithAggregations(10), nil)
			})
			It("should return the expected issues in the result", func() {
				issueHandler = issue.NewIssueHandler(handlerContext)
				res, err := issueHandler.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(10), "return 10 results")
			})
		})
		Context("and the database operations throw an error", func() {
			BeforeEach(func() {
				db.On("GetIssuesWithAggregations", filter, []entity.Order{}).Return([]entity.IssueResult{}, errors.New("database error"))
			})

			It("should return the expected issues in the result", func() {
				issueHandler = issue.NewIssueHandler(handlerContext)
				_, err := issueHandler.ListIssues(filter, options)

				Expect(err).ToNot(BeNil(), "error should be returned")

				var appErr *appErrors.Error
				Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
				Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
				Expect(appErr.Entity).To(Equal("Issues"), "should reference Issues entity")
				Expect(appErr.ID).To(Equal(""), "should have empty ID for list operation")
				Expect(appErr.Op).To(Equal("issueHandler.ListIssues"), "should include operation")
			})
		})
	})

	When("the list options does NOT include aggregations", func() {
		BeforeEach(func() {
			options.IncludeAggregations = false
		})

		Context("and the given filter does not have any matches in the database", func() {
			BeforeEach(func() {
				db.On("GetIssues", filter, []entity.Order{}).Return([]entity.IssueResult{}, nil)
			})
			It("should return an empty result", func() {
				issueHandler = issue.NewIssueHandler(handlerContext)
				res, err := issueHandler.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(0), "return no results")
			})
		})
		Context("and the filter does have results in the database", func() {
			BeforeEach(func() {
				db.On("GetIssues", filter, []entity.Order{}).Return(test.NNewFakeIssueResults(15), nil)
			})
			It("should return the expected issues in the result", func() {
				issueHandler = issue.NewIssueHandler(handlerContext)
				res, err := issueHandler.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(15), "return 15 results")
			})
		})

		Context("and the database operations throw an error", func() {
			BeforeEach(func() {
				db.On("GetIssues", filter, []entity.Order{}).Return([]entity.IssueResult{}, errors.New("database error"))
			})

			It("should return the expected issues in the result", func() {
				issueHandler = issue.NewIssueHandler(handlerContext)
				_, err := issueHandler.ListIssues(filter, options)

				Expect(err).ToNot(BeNil(), "error should be returned")

				var appErr *appErrors.Error
				Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
				Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
				Expect(appErr.Entity).To(Equal("Issues"), "should reference Issues entity")
				Expect(appErr.ID).To(Equal(""), "should have empty ID for list operation")
				Expect(appErr.Op).To(Equal("issueHandler.ListIssues"), "should include operation")
			})
		})
	})

	Context("when GetAllIssueCursors fails", func() {
		BeforeEach(func() {
			options.ShowPageInfo = true
			filter.First = lo.ToPtr(10)
		})

		It("should return Internal error", func() {
			// Mock successful GetIssues but failing GetAllIssueCursors
			db.On("GetIssues", filter, []entity.Order{}).Return(test.NNewFakeIssueResults(5), nil)
			cursorsError := errors.New("cursor database error")
			db.On("GetAllIssueCursors", filter, []entity.Order{}).Return([]string{}, cursorsError)

			issueHandler = issue.NewIssueHandler(handlerContext)
			_, err := issueHandler.ListIssues(filter, options)

			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("IssueCursors"), "should reference IssueCursors entity")
			Expect(appErr.ID).To(Equal(""), "should have empty ID for list operation")
			Expect(appErr.Op).To(Equal("issueHandler.ListIssues"), "should include operation")
		})
	})

	Context("when CountIssueTypes fails", func() {
		BeforeEach(func() {
			options.ShowTotalCount = true
		})

		It("should return Internal error", func() {
			// Mock successful GetIssues but failing CountIssueTypes
			db.On("GetIssues", filter, []entity.Order{}).Return([]entity.IssueResult{}, nil)
			countError := errors.New("count database error")
			db.On("CountIssueTypes", filter).Return((*entity.IssueTypeCounts)(nil), countError)

			issueHandler = issue.NewIssueHandler(handlerContext)
			_, err := issueHandler.ListIssues(filter, options)

			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("IssueTypeCounts"), "should reference IssueTypeCounts entity")
			Expect(appErr.ID).To(Equal(""), "should have empty ID for list operation")
			Expect(appErr.Op).To(Equal("issueHandler.ListIssues"), "should include operation")
		})
	})
})

var _ = Describe("When listing Issue Names", Label("app", "ListIssueNames"), func() {
	var (
		db             *mocks.MockDatabase
		issueHandler   issue.IssueHandler
		filter         *entity.IssueFilter
		options        *entity.ListOptions
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		filter = getIssueFilter()
		options = entity.NewListOptions()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	Context("with valid input", func() {
		It("returns issue names successfully", func() {
			expectedNames := []string{"CVE-2023-1234", "CVE-2023-5678", "POLICY-001"}
			db.On("GetIssueNames", filter).Return(expectedNames, nil)

			issueHandler = issue.NewIssueHandler(handlerContext)
			result, err := issueHandler.ListIssueNames(filter, options)

			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(result).ToNot(BeNil(), "result should be returned")
			Expect(result).To(Equal(expectedNames), "should return expected issue names")
			Expect(len(result)).To(Equal(3), "should return correct number of names")
		})

		It("returns empty list when no issues found", func() {
			expectedNames := []string{}
			db.On("GetIssueNames", filter).Return(expectedNames, nil)

			issueHandler = issue.NewIssueHandler(handlerContext)
			result, err := issueHandler.ListIssueNames(filter, options)

			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(result).ToNot(BeNil(), "result should be returned")
			Expect(result).To(BeEmpty(), "should return empty list")
		})
	})

	Context("when database operation fails", func() {
		It("should return Internal error", func() {
			// Mock database error
			dbError := errors.New("database connection failed")
			db.On("GetIssueNames", filter).Return([]string{}, dbError)

			issueHandler = issue.NewIssueHandler(handlerContext)
			result, err := issueHandler.ListIssueNames(filter, options)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("IssueNames"), "should reference IssueNames entity")
			Expect(appErr.ID).To(Equal(""), "should have empty ID for list operation")
			Expect(appErr.Op).To(Equal("issueHandler.ListIssueNames"), "should include operation")
			Expect(appErr.Err.Error()).To(ContainSubstring("database connection failed"), "should contain original error message")
		})
	})
})

var _ = Describe("When creating Issue", Label("app", "CreateIssue"), func() {
	var (
		db             *mocks.MockDatabase
		issueHandler   issue.IssueHandler
		issueEntity    entity.Issue
		filter         *entity.IssueFilter
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueEntity = test.NewFakeIssueEntity()
		first := 10
		after := ""
		filter = &entity.IssueFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	Context("with valid input and no conflicts", func() {
		It("creates issue successfully", func() {
			filter.PrimaryName = []*string{&issueEntity.PrimaryName}
			// Mock successful user ID retrieval
			db.On("GetAllUserIds", mock.Anything).Return([]int64{123}, nil)
			// Mock no existing issues with same primary name
			db.On("GetIssues", filter, []entity.Order{}).Return([]entity.IssueResult{}, nil)
			// Mock successful database creation
			db.On("CreateIssue", mock.AnythingOfType("*entity.Issue")).Return(&issueEntity, nil)

			issueHandler = issue.NewIssueHandler(handlerContext)
			newIssue, err := issueHandler.CreateIssue(common.NewAdminContext(), &issueEntity)

			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(newIssue).ToNot(BeNil(), "issue should be returned")
			Expect(newIssue.Id).NotTo(BeEquivalentTo(0))
			Expect(newIssue.CreatedBy).To(Equal(int64(123)), "should set CreatedBy from user ID")
			Expect(newIssue.UpdatedBy).To(Equal(int64(123)), "should set UpdatedBy from user ID")
			By("setting fields", func() {
				Expect(newIssue.PrimaryName).To(BeEquivalentTo(issueEntity.PrimaryName))
				Expect(newIssue.Description).To(BeEquivalentTo(issueEntity.Description))
				Expect(newIssue.Type.String()).To(BeEquivalentTo(issueEntity.Type.String()))
			})
		})
	})

	Context("when GetCurrentUserId fails", func() {
		It("should return Internal error", func() {
			// Mock GetCurrentUserId failure
			dbError := errors.New("user database connection failed")
			db.On("GetAllUserIds", mock.Anything).Return([]int64{}, dbError)

			issueHandler = issue.NewIssueHandler(handlerContext)
			result, err := issueHandler.CreateIssue(common.NewAdminContext(), &issueEntity)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			// Verify it's our structured error with correct code
			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("Issue"), "should reference Issue entity")
			Expect(appErr.Op).To(Equal("issueHandler.CreateIssue"), "should include operation")
			Expect(appErr.Message).To(BeEmpty(), "Internal errors from InternalError helper don't set custom messages")
			// The GetCurrentUserId function wraps the original error, so we need to check the wrapped error message
			Expect(appErr.Err.Error()).To(ContainSubstring("user database connection failed"), "should wrap original error")
		})
	})

	Context("when checking for existing issues fails", func() {
		It("should return Internal error", func() {
			// Mock successful user ID retrieval
			db.On("GetAllUserIds", mock.Anything).Return([]int64{123}, nil)
			// Mock ListIssues failure
			listError := errors.New("database query failed")
			db.On("GetIssues", mock.Anything, []entity.Order{}).Return([]entity.IssueResult{}, listError)

			issueHandler = issue.NewIssueHandler(handlerContext)
			result, err := issueHandler.CreateIssue(common.NewAdminContext(), &issueEntity)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("Issue"), "should reference Issue entity")
			Expect(appErr.Op).To(Equal("issueHandler.CreateIssue"), "should include operation")
			Expect(appErr.Message).To(BeEmpty(), "Internal errors from InternalError helper don't set custom messages")
		})
	})

	Context("when issue with same primary name already exists", func() {
		It("should return AlreadyExists error", func() {
			// Mock successful user ID retrieval
			db.On("GetAllUserIds", mock.Anything).Return([]int64{123}, nil)
			// Mock existing issue with same primary name
			existingIssue := test.NewFakeIssueEntity()
			existingIssue.Id = 999
			existingIssue.PrimaryName = issueEntity.PrimaryName
			db.On("GetIssues", mock.Anything, []entity.Order{}).Return([]entity.IssueResult{{
				Issue: &existingIssue,
			}}, nil)

			issueHandler = issue.NewIssueHandler(handlerContext)
			result, err := issueHandler.CreateIssue(common.NewAdminContext(), &issueEntity)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.AlreadyExists), "should be AlreadyExists error")
			Expect(appErr.Entity).To(Equal("Issue"), "should reference Issue entity")
			Expect(appErr.ID).To(Equal(issueEntity.PrimaryName), "should include primary name as ID")
			Expect(appErr.Op).To(Equal("issueHandler.CreateIssue"), "should include operation")
			Expect(appErr.Message).To(Equal("already exists"), "should have standard AlreadyExists message")
		})
	})

	Context("when database creation fails", func() {
		It("should return Internal error", func() {
			// Mock successful user ID retrieval
			db.On("GetAllUserIds", mock.Anything).Return([]int64{123}, nil)
			// Mock no existing issues
			db.On("GetIssues", mock.Anything, []entity.Order{}).Return([]entity.IssueResult{}, nil)
			// Mock database creation failure
			dbError := errors.New("constraint violation")
			db.On("CreateIssue", mock.AnythingOfType("*entity.Issue")).Return((*entity.Issue)(nil), dbError)

			issueHandler = issue.NewIssueHandler(handlerContext)
			result, err := issueHandler.CreateIssue(common.NewAdminContext(), &issueEntity)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("Issue"), "should reference Issue entity")
			Expect(appErr.Op).To(Equal("issueHandler.CreateIssue"), "should include operation")
			Expect(appErr.Message).To(BeEmpty(), "Internal errors from InternalError helper don't set custom messages")
			Expect(appErr.Err).To(Equal(dbError), "should wrap original error")
		})
	})
})

var _ = Describe("When updating Issue", Label("app", "UpdateIssue"), func() {
	var (
		db             *mocks.MockDatabase
		issueHandler   issue.IssueHandler
		issueResult    entity.IssueResult
		filter         *entity.IssueFilter
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueResult = test.NewFakeIssueResult()
		first := 10
		after := ""
		filter = &entity.IssueFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})
	Context("with valid input", func() {
		It("updates issueEntity successfully", func() {
			// Setup mocks for successful path
			db.On("GetAllUserIds", mock.Anything).Return([]int64{123}, nil)
			db.On("UpdateIssue", issueResult.Issue).Return(nil)
			filter.Id = []*int64{&issueResult.Issue.Id}
			db.On("GetIssues", filter, []entity.Order{}).Return([]entity.IssueResult{issueResult}, nil)

			issueHandler = issue.NewIssueHandler(handlerContext)
			issueResult.Issue.Description = "New Description"

			updatedIssue, err := issueHandler.UpdateIssue(common.NewAdminContext(), issueResult.Issue)

			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(updatedIssue).ToNot(BeNil(), "updated issue should be returned")
			Expect(updatedIssue.Description).To(Equal("New Description"))
		})
	})

	Context("when GetCurrentUserId fails", func() {
		It("should return Internal error", func() {
			// Mock GetCurrentUserId failure
			dbError := errors.New("user database connection failed")
			db.On("GetAllUserIds", mock.Anything).Return([]int64{}, dbError)

			issueHandler = issue.NewIssueHandler(handlerContext)
			result, err := issueHandler.UpdateIssue(common.NewAdminContext(), issueResult.Issue)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			// Verify structured error
			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("Issue"), "should reference Issue entity")
			Expect(appErr.ID).To(Equal(strconv.FormatInt(issueResult.Issue.Id, 10)), "should include issue ID")
			Expect(appErr.Op).To(Equal("issueHandler.UpdateIssue"), "should include operation")
		})
	})

	Context("when database update fails", func() {
		It("should return Internal error", func() {
			// Mock successful user ID retrieval
			db.On("GetAllUserIds", mock.Anything).Return([]int64{123}, nil)
			// Mock database update failure
			dbError := errors.New("constraint violation")
			db.On("UpdateIssue", issueResult.Issue).Return(dbError)

			issueHandler = issue.NewIssueHandler(handlerContext)
			result, err := issueHandler.UpdateIssue(common.NewAdminContext(), issueResult.Issue)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("Issue"), "should reference Issue entity")
			Expect(appErr.ID).To(Equal(strconv.FormatInt(issueResult.Issue.Id, 10)), "should include issue ID")
			Expect(appErr.Op).To(Equal("issueHandler.UpdateIssue"), "should include operation")
			Expect(appErr.Err).To(Equal(dbError), "should wrap original error")
		})
	})

	Context("when retrieving updated issue fails", func() {
		It("should return Internal error", func() {
			// Mock successful user ID and update
			db.On("GetAllUserIds", mock.Anything).Return([]int64{123}, nil)
			db.On("UpdateIssue", issueResult.Issue).Return(nil)
			// Mock ListIssues failure
			listError := errors.New("database query failed")
			db.On("GetIssues", mock.Anything, []entity.Order{}).Return([]entity.IssueResult{}, listError)

			issueHandler = issue.NewIssueHandler(handlerContext)
			result, err := issueHandler.UpdateIssue(common.NewAdminContext(), issueResult.Issue)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("Issue"), "should reference Issue entity")
			Expect(appErr.ID).To(Equal(strconv.FormatInt(issueResult.Issue.Id, 10)), "should include issue ID")
			Expect(appErr.Op).To(Equal("issueHandler.UpdateIssue"), "should include operation")
		})
	})

	It("updates issueEntity", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateIssue", issueResult.Issue).Return(nil)
		issueHandler = issue.NewIssueHandler(handlerContext)
		issueResult.Issue.Description = "New Description"
		filter.Id = []*int64{&issueResult.Issue.Id}
		db.On("GetIssues", filter, []entity.Order{}).Return([]entity.IssueResult{issueResult}, nil)
		updatedIssue, err := issueHandler.UpdateIssue(common.NewAdminContext(), issueResult.Issue)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedIssue.PrimaryName).To(BeEquivalentTo(issueResult.PrimaryName))
			Expect(updatedIssue.Description).To(BeEquivalentTo(issueResult.Issue.Description))
			Expect(updatedIssue.Type.String()).To(BeEquivalentTo(issueResult.Type.String()))
		})
	})
})

var _ = Describe("When deleting Issue", Label("app", "DeleteIssue"), func() {
	var (
		db             *mocks.MockDatabase
		issueHandler   issue.IssueHandler
		id             int64
		filter         *entity.IssueFilter
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		after := ""
		filter = &entity.IssueFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})
	Context("with valid input", func() {
		It("deletes issue successfully", func() {
			db.On("GetAllUserIds", mock.Anything).Return([]int64{123}, nil)
			db.On("DeleteIssue", id, int64(123)).Return(nil)
			db.On("GetIssues", mock.Anything, []entity.Order{}).Return([]entity.IssueResult{}, nil)

			issueHandler = issue.NewIssueHandler(handlerContext)
			err := issueHandler.DeleteIssue(common.NewAdminContext(), id)

			Expect(err).To(BeNil(), "no error should be thrown")

			filter.Id = []*int64{&id}
			lo := entity.IssueListOptions{
				ListOptions: *entity.NewListOptions(),
			}
			issues, err := issueHandler.ListIssues(filter, &lo)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(issues.Elements).To(BeEmpty(), "issue should be deleted")
		})
	})

	Context("when GetCurrentUserId fails", func() {
		It("should return Internal error", func() {
			// Mock GetCurrentUserId failure
			dbError := errors.New("user database connection failed")
			db.On("GetAllUserIds", mock.Anything).Return([]int64{}, dbError)

			issueHandler = issue.NewIssueHandler(handlerContext)
			err := issueHandler.DeleteIssue(common.NewAdminContext(), id)

			Expect(err).ToNot(BeNil(), "error should be returned")

			// Verify structured error
			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("Issue"), "should reference Issue entity")
			Expect(appErr.ID).To(Equal(strconv.FormatInt(id, 10)), "should include issue ID")
			Expect(appErr.Op).To(Equal("issueHandler.DeleteIssue"), "should include operation")
		})
	})

	Context("when database delete fails", func() {
		It("should return Internal error", func() {
			// Mock successful user ID retrieval
			db.On("GetAllUserIds", mock.Anything).Return([]int64{123}, nil)
			// Mock database delete failure
			dbError := errors.New("foreign key constraint violation")
			db.On("DeleteIssue", id, int64(123)).Return(dbError)

			issueHandler = issue.NewIssueHandler(handlerContext)
			err := issueHandler.DeleteIssue(common.NewAdminContext(), id)

			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("Issue"), "should reference Issue entity")
			Expect(appErr.ID).To(Equal(strconv.FormatInt(id, 10)), "should include issue ID")
			Expect(appErr.Op).To(Equal("issueHandler.DeleteIssue"), "should include operation")
			Expect(appErr.Err).To(Equal(dbError), "should wrap original error")
		})
	})
})

var _ = Describe("When modifying relationship of ComponentVersion and Issue", Label("app", "ComponentVersionIssueRelationship"), func() {
	var (
		db               *mocks.MockDatabase
		issueHandler     issue.IssueHandler
		issueResult      entity.IssueResult
		componentVersion entity.ComponentVersion
		handlerContext   common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueResult = test.NewFakeIssueResult()
		componentVersion = test.NewFakeComponentVersionEntity()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	It("adds componentVersion to issueEntity", func() {
		db.On("AddComponentVersionToIssue", issueResult.Issue.Id, componentVersion.Id).Return(nil)
		db.On("GetIssues", mock.Anything, mock.Anything).Return([]entity.IssueResult{issueResult}, nil)
		issueHandler = issue.NewIssueHandler(handlerContext)
		issue, err := issueHandler.AddComponentVersionToIssue(issueResult.Issue.Id, componentVersion.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issue).NotTo(BeNil(), "issueEntity should be returned")
	})

	It("removes componentVersion from issueEntity", func() {
		db.On("RemoveComponentVersionFromIssue", issueResult.Issue.Id, componentVersion.Id).Return(nil)
		db.On("GetIssues", mock.Anything, mock.Anything).Return([]entity.IssueResult{issueResult}, nil)
		issueHandler = issue.NewIssueHandler(handlerContext)
		issue, err := issueHandler.RemoveComponentVersionFromIssue(issueResult.Issue.Id, componentVersion.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issue).NotTo(BeNil(), "issueEntity should be returned")
	})
})

var _ = Describe("When getting Issue Severity Counts", Label("app", "GetIssueSeverityCounts"), func() {
	var (
		db             *mocks.MockDatabase
		issueHandler   issue.IssueHandler
		filter         *entity.IssueFilter
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		filter = getIssueFilter()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	Context("with valid input", func() {
		It("returns issue severity counts successfully", func() {
			expectedCounts := &entity.IssueSeverityCounts{
				Critical: 10,
				High:     25,
				Medium:   50,
				Low:      15,
			}
			db.On("CountIssueRatings", filter).Return(expectedCounts, nil)

			issueHandler = issue.NewIssueHandler(handlerContext)
			result, err := issueHandler.GetIssueSeverityCounts(filter)

			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(result).ToNot(BeNil(), "result should be returned")
			Expect(result).To(Equal(expectedCounts), "should return expected severity counts")
			Expect(result.Critical).To(Equal(int64(10)), "should return correct critical count")
			Expect(result.High).To(Equal(int64(25)), "should return correct high count")
			Expect(result.Medium).To(Equal(int64(50)), "should return correct medium count")
			Expect(result.Low).To(Equal(int64(15)), "should return correct low count")
		})

		It("returns zero counts when no issues found", func() {
			expectedCounts := &entity.IssueSeverityCounts{
				Critical: 0,
				High:     0,
				Medium:   0,
				Low:      0,
			}
			db.On("CountIssueRatings", filter).Return(expectedCounts, nil)

			issueHandler = issue.NewIssueHandler(handlerContext)
			result, err := issueHandler.GetIssueSeverityCounts(filter)

			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(result).ToNot(BeNil(), "result should be returned")
			Expect(result.Critical).To(Equal(int64(0)), "should return zero critical count")
			Expect(result.High).To(Equal(int64(0)), "should return zero high count")
			Expect(result.Medium).To(Equal(int64(0)), "should return zero medium count")
			Expect(result.Low).To(Equal(int64(0)), "should return zero low count")
		})
	})

	Context("when database operation fails", func() {
		It("should return Internal error", func() {
			// Mock database error
			dbError := errors.New("database aggregation failed")
			db.On("CountIssueRatings", filter).Return((*entity.IssueSeverityCounts)(nil), dbError)

			issueHandler = issue.NewIssueHandler(handlerContext)
			result, err := issueHandler.GetIssueSeverityCounts(filter)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("IssueSeverityCounts"), "should reference IssueSeverityCounts entity")
			Expect(appErr.ID).To(Equal(""), "should have empty ID for aggregation operation")
			Expect(appErr.Op).To(Equal("issueHandler.GetIssueSeverityCounts"), "should include operation")
			Expect(appErr.Err.Error()).To(ContainSubstring(dbError.Error()), "should contain original error message")
		})
	})
})
