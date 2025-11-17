//// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
//// SPDX-License-Identifier: Apache-2.0

package issue_variant_test

import (
	"math"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/app/issue_repository"
	iv "github.com/cloudoperators/heureka/internal/app/issue_variant"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	"github.com/cloudoperators/heureka/internal/openfga"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	mock "github.com/stretchr/testify/mock"
)

func TestIssueVariantHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IssueVariant Service Test Suite")
}

var er event.EventRegistry

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)

})

func issueVariantFilter() *entity.IssueVariantFilter {
	return &entity.IssueVariantFilter{
		Paginated: entity.Paginated{
			First: nil,
			After: nil,
		},
		IssueId:           nil,
		IssueRepositoryId: nil,
		ServiceId:         nil,
		IssueMatchId:      nil,
	}
}

func issueRepositoryFilter() *entity.IssueRepositoryFilter {
	return &entity.IssueRepositoryFilter{
		Paginated: entity.Paginated{
			First: nil,
			After: nil,
		},
		Id:          nil,
		ServiceId:   nil,
		Name:        nil,
		ServiceCCRN: nil,
	}
}

func issueVariantListOptions() *entity.ListOptions {
	return &entity.ListOptions{
		ShowTotalCount:      false,
		ShowPageInfo:        false,
		IncludeAggregations: false,
	}
}

var _ = Describe("When listing IssueVariants", Label("app", "ListIssueVariants"), func() {
	var (
		db                  *mocks.MockDatabase
		issueVariantHandler iv.IssueVariantHandler
		rs                  issue_repository.IssueRepositoryHandler
		filter              *entity.IssueVariantFilter
		options             *entity.ListOptions
		authz               openfga.Authorization
		handlerContext      common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = issueVariantListOptions()
		filter = issueVariantFilter()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
		rs = issue_repository.NewIssueRepositoryHandler(handlerContext)
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetIssueVariants", filter).Return([]entity.IssueVariant{}, nil)
			db.On("CountIssueVariants", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			issueVariantHandler = iv.NewIssueVariantHandler(handlerContext, rs)
			res, err := issueVariantHandler.ListIssueVariants(filter, options)
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
			advisories := test.NNewFakeIssueVariants(resElements)

			var ids = lo.Map(advisories, func(iv entity.IssueVariant, _ int) int64 { return iv.Id })
			var i int64 = 0
			for len(ids) < dbElements {
				i++
				ids = append(ids, i)
			}
			db.On("GetIssueVariants", filter).Return(advisories, nil)
			db.On("GetAllIssueVariantIds", filter).Return(ids, nil)
			issueVariantHandler = iv.NewIssueVariantHandler(handlerContext, rs)
			res, err := issueVariantHandler.ListIssueVariants(filter, options)
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
})

var _ = Describe("When listing EffectiveIssueVariants", Label("app", "ListEffectiveIssueVariants"), func() {
	var (
		db                  *mocks.MockDatabase
		issueVariantHandler iv.IssueVariantHandler
		ivFilter            *entity.IssueVariantFilter
		irFilter            *entity.IssueRepositoryFilter
		options             *entity.ListOptions
		issueVariants       []entity.IssueVariant
		repositories        []entity.IssueRepository
		rs                  issue_repository.IssueRepositoryHandler
		authz               openfga.Authorization
		handlerContext      common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = issueVariantListOptions()
		ivFilter = issueVariantFilter()
		irFilter = issueRepositoryFilter()
		first := 10
		ivFilter.First = &first
		var after int64 = 0
		irFilter.First = &first
		irFilter.After = &after
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
		rs = issue_repository.NewIssueRepositoryHandler(handlerContext)
	})

	When("having different priority", func() {
		BeforeEach(func() {
			issueVariants = test.NNewFakeIssueVariants(25)
			repositories = test.NNewFakeIssueRepositories(2)
			repositories[0].Priority = 1
			repositories[1].Priority = 2
			for i := range issueVariants {
				issueVariants[i].IssueRepositoryId = repositories[i%2].Id
			}
			irFilter.Id = lo.Map(issueVariants, func(item entity.IssueVariant, _ int) *int64 {
				return &item.IssueRepositoryId
			})
			db.On("GetIssueVariants", ivFilter).Return(issueVariants, nil)
			db.On("GetIssueRepositories", irFilter).Return(repositories, nil)
		})
		It("can list advisories", func() {
			issueVariantHandler = iv.NewIssueVariantHandler(handlerContext, rs)
			res, err := issueVariantHandler.ListEffectiveIssueVariants(ivFilter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			for _, item := range res.Elements {
				Expect(item.IssueRepositoryId).To(BeEquivalentTo(repositories[1].Id))
			}
		})
	})
	When("having same priority", func() {
		BeforeEach(func() {
			issueVariants = test.NNewFakeIssueVariants(25)
			repositories = test.NNewFakeIssueRepositories(2)
			repositories[0].Priority = 1
			repositories[1].Priority = 1
			for i := range issueVariants {
				issueVariants[i].IssueRepositoryId = repositories[i%2].Id
			}
			irFilter.Id = lo.Map(issueVariants, func(item entity.IssueVariant, _ int) *int64 {
				return &item.IssueRepositoryId
			})
			db.On("GetIssueVariants", ivFilter).Return(issueVariants, nil)
			db.On("GetIssueRepositories", irFilter).Return(repositories, nil)
		})
		It("can list issueVariants", func() {
			issueVariantHandler = iv.NewIssueVariantHandler(handlerContext, rs)
			res, err := issueVariantHandler.ListEffectiveIssueVariants(ivFilter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			ir_ids := lo.Map(res.Elements, func(item entity.IssueVariantResult, _ int) int64 {
				return item.IssueRepositoryId
			})
			Expect(lo.Contains(ir_ids, repositories[0].Id)).To(BeTrue())
			Expect(lo.Contains(ir_ids, repositories[1].Id)).To(BeTrue())
		})
	})
})

var _ = Describe("When creating IssueVariant", Label("app", "CreateIssueVariant"), func() {
	var (
		db                  *mocks.MockDatabase
		issueVariantHandler iv.IssueVariantHandler
		issueVariant        entity.IssueVariant
		filter              *entity.IssueVariantFilter
		rs                  issue_repository.IssueRepositoryHandler
		authz               openfga.Authorization
		handlerContext      common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueVariant = test.NewFakeIssueVariantEntity(nil)
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueVariantFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
		rs = issue_repository.NewIssueRepositoryHandler(handlerContext)
	})

	It("creates issueVariant", func() {
		filter.SecondaryName = []*string{&issueVariant.SecondaryName}
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("CreateIssueVariant", &issueVariant).Return(&issueVariant, nil)
		db.On("GetIssueVariants", filter).Return([]entity.IssueVariant{}, nil)
		issueVariantHandler = iv.NewIssueVariantHandler(handlerContext, rs)
		newIssueVariant, err := issueVariantHandler.CreateIssueVariant(common.NewAdminContext(), &issueVariant)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newIssueVariant.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newIssueVariant.SecondaryName).To(BeEquivalentTo(issueVariant.SecondaryName))
			Expect(newIssueVariant.Description).To(BeEquivalentTo(issueVariant.Description))
			Expect(newIssueVariant.IssueRepositoryId).To(BeEquivalentTo(issueVariant.IssueRepositoryId))
			Expect(newIssueVariant.IssueId).To(BeEquivalentTo(issueVariant.IssueId))
			Expect(newIssueVariant.Severity.Cvss.Vector).To(BeEquivalentTo(issueVariant.Severity.Cvss.Vector))
			Expect(newIssueVariant.Severity.Score).To(BeEquivalentTo(issueVariant.Severity.Score))
			Expect(newIssueVariant.Severity.Value).To(BeEquivalentTo(issueVariant.Severity.Value))
		})
	})
})

var _ = Describe("When updating IssueVariant", Label("app", "UpdateIssueVariant"), func() {
	var (
		db                  *mocks.MockDatabase
		issueVariantHandler iv.IssueVariantHandler
		issueVariant        entity.IssueVariant
		filter              *entity.IssueVariantFilter
		rs                  issue_repository.IssueRepositoryHandler
		authz               openfga.Authorization
		handlerContext      common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueVariant = test.NewFakeIssueVariantEntity(nil)
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueVariantFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
		rs = issue_repository.NewIssueRepositoryHandler(handlerContext)
	})

	It("updates issueVariant", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateIssueVariant", &issueVariant).Return(nil)
		issueVariantHandler = iv.NewIssueVariantHandler(handlerContext, rs)
		issueVariant.SecondaryName = "SecretAdvisory"
		filter.Id = []*int64{&issueVariant.Id}
		db.On("GetIssueVariants", filter).Return([]entity.IssueVariant{issueVariant}, nil)
		updatedIssueVariant, err := issueVariantHandler.UpdateIssueVariant(common.NewAdminContext(), &issueVariant)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedIssueVariant.SecondaryName).To(BeEquivalentTo(issueVariant.SecondaryName))
			Expect(updatedIssueVariant.Description).To(BeEquivalentTo(issueVariant.Description))
			Expect(updatedIssueVariant.IssueRepositoryId).To(BeEquivalentTo(issueVariant.IssueRepositoryId))
			Expect(updatedIssueVariant.IssueId).To(BeEquivalentTo(issueVariant.IssueId))
			Expect(updatedIssueVariant.Severity.Cvss.Vector).To(BeEquivalentTo(issueVariant.Severity.Cvss.Vector))
			Expect(updatedIssueVariant.Severity.Score).To(BeEquivalentTo(issueVariant.Severity.Score))
			Expect(updatedIssueVariant.Severity.Value).To(BeEquivalentTo(issueVariant.Severity.Value))
		})
	})
})

var _ = Describe("When deleting IssueVariant", Label("app", "DeleteIssueVariant"), func() {
	var (
		db                  *mocks.MockDatabase
		issueVariantHandler iv.IssueVariantHandler
		id                  int64
		filter              *entity.IssueVariantFilter
		rs                  issue_repository.IssueRepositoryHandler
		authz               openfga.Authorization
		handlerContext      common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueVariantFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
		rs = issue_repository.NewIssueRepositoryHandler(handlerContext)
	})

	It("deletes issueVariant", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("DeleteIssueVariant", id, mock.Anything).Return(nil)
		issueVariantHandler = iv.NewIssueVariantHandler(handlerContext, rs)
		db.On("GetIssueVariants", filter).Return([]entity.IssueVariant{}, nil)
		err := issueVariantHandler.DeleteIssueVariant(common.NewAdminContext(), id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		issueVariants, err := issueVariantHandler.ListIssueVariants(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issueVariants.Elements).To(BeEmpty(), "no error should be thrown")
	})
})
