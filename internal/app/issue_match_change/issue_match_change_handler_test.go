// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_match_change_test

import (
	"math"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/event"
	imc "github.com/cloudoperators/heureka/internal/app/issue_match_change"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	mock "github.com/stretchr/testify/mock"
)

func TestIssueHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IssueMatchChange Service Test Suite")
}

var er event.EventRegistry

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)
})

func getIssueMatchChangeFilter() *entity.IssueMatchChangeFilter {
	return &entity.IssueMatchChangeFilter{
		Paginated: entity.Paginated{
			First: nil,
			After: nil,
		},
		Id:           nil,
		Action:       nil,
		ActivityId:   nil,
		IssueMatchId: nil,
	}
}

var _ = Describe("When listing IssueMatchChanges", Label("app", "ListIssueMatchChanges"), func() {
	var (
		db                      *mocks.MockDatabase
		issueMatchChangeHandler imc.IssueMatchChangeHandler
		filter                  *entity.IssueMatchChangeFilter
		options                 *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getIssueMatchChangeFilter()
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetIssueMatchChanges", filter).Return([]entity.IssueMatchChange{}, nil)
			db.On("CountIssueMatchChanges", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			issueMatchChangeHandler = imc.NewIssueMatchChangeHandler(db, er)
			res, err := issueMatchChangeHandler.ListIssueMatchChanges(filter, options)
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
			imcs := test.NNewFakeIssueMatchChanges(resElements)

			var ids = lo.Map(imcs, func(imc entity.IssueMatchChange, _ int) int64 { return imc.Id })
			var i int64 = 0
			for len(ids) < dbElements {
				i++
				ids = append(ids, i)
			}
			db.On("GetIssueMatchChanges", filter).Return(imcs, nil)
			db.On("GetAllIssueMatchChangeIds", filter).Return(ids, nil)
			issueMatchChangeHandler = imc.NewIssueMatchChangeHandler(db, er)
			res, err := issueMatchChangeHandler.ListIssueMatchChanges(filter, options)
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

var _ = Describe("When creating IssueMatchChange", Label("app", "CreateIssueMatchChange"), func() {
	var (
		db                      *mocks.MockDatabase
		issueMatchChangeHandler imc.IssueMatchChangeHandler
		issueMatchChange        entity.IssueMatchChange
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueMatchChange = test.NewFakeIssueMatchChange()
	})

	It("creates issueMatchChange", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("CreateIssueMatchChange", &issueMatchChange).Return(&issueMatchChange, nil)
		issueMatchChangeHandler = imc.NewIssueMatchChangeHandler(db, er)
		newIssueMatchChange, err := issueMatchChangeHandler.CreateIssueMatchChange(&issueMatchChange)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newIssueMatchChange.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newIssueMatchChange.Action).To(BeEquivalentTo(issueMatchChange.Action))
		})
	})
})

var _ = Describe("When updating IssueMatchChange", Label("app", "UpdateIssueMatchChange"), func() {
	var (
		db                      *mocks.MockDatabase
		issueMatchChangeHandler imc.IssueMatchChangeHandler
		issueMatchChange        entity.IssueMatchChange
		filter                  *entity.IssueMatchChangeFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueMatchChange = test.NewFakeIssueMatchChange()
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueMatchChangeFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("updates issueMatchChange", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateIssueMatchChange", &issueMatchChange).Return(nil)
		issueMatchChangeHandler = imc.NewIssueMatchChangeHandler(db, er)
		if issueMatchChange.Action == entity.IssueMatchChangeActionAdd.String() {
			issueMatchChange.Action = entity.IssueMatchChangeActionRemove.String()
		} else {
			issueMatchChange.Action = entity.IssueMatchChangeActionAdd.String()
		}
		filter.Id = []*int64{&issueMatchChange.Id}
		db.On("GetIssueMatchChanges", filter).Return([]entity.IssueMatchChange{issueMatchChange}, nil)
		updatedIssueMatchChange, err := issueMatchChangeHandler.UpdateIssueMatchChange(&issueMatchChange)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedIssueMatchChange.Action).To(BeEquivalentTo(issueMatchChange.Action))
		})
	})
})

var _ = Describe("When deleting IssueMatchChange", Label("app", "DeleteIssueMatchChange"), func() {
	var (
		db                      *mocks.MockDatabase
		issueMatchChangeHandler imc.IssueMatchChangeHandler
		id                      int64
		filter                  *entity.IssueMatchChangeFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueMatchChangeFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("deletes issueMatchChange", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("DeleteIssueMatchChange", id, mock.Anything).Return(nil)
		issueMatchChangeHandler = imc.NewIssueMatchChangeHandler(db, er)
		db.On("GetIssueMatchChanges", filter).Return([]entity.IssueMatchChange{}, nil)
		err := issueMatchChangeHandler.DeleteIssueMatchChange(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		issueMatchChanges, err := issueMatchChangeHandler.ListIssueMatchChanges(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issueMatchChanges.Elements).To(BeEmpty(), "no error should be thrown")
	})
})
