// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package remediation_test

import (
	"errors"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/stretchr/testify/mock"

	"math"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/event"
	rh "github.com/cloudoperators/heureka/internal/app/remediation"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/mocks"
	"github.com/cloudoperators/heureka/internal/openfga"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

func TestRemediationHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Remediation Test Suite")
}

var er event.EventRegistry
var authz openfga.Authorization

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)
})

var _ = Describe("When listing Remediations", Label("app", "ListRemediations"), func() {
	var (
		db                 *mocks.MockDatabase
		remediationHandler rh.RemediationHandler
		filter             *entity.RemediationFilter
		options            *entity.ListOptions
		handlerContext     common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = &entity.RemediationFilter{
			PaginatedX: entity.PaginatedX{
				First: nil,
				After: nil,
			},
		}
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetRemediations", filter, []entity.Order{}).Return([]entity.RemediationResult{}, nil)
			db.On("CountRemediations", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			remediationHandler = rh.NewRemediationHandler(handlerContext)
			res, err := remediationHandler.ListRemediations(filter, options)
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
			remediations := []entity.RemediationResult{}
			for _, remediation := range test.NNewFakeRemediations(resElements) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithRemediation([]entity.Order{}, remediation))
				remediations = append(remediations, entity.RemediationResult{WithCursor: entity.WithCursor{Value: cursor}, Remediation: lo.ToPtr(remediation)})
			}

			var cursors = lo.Map(remediations, func(m entity.RemediationResult, _ int) string {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithRemediation([]entity.Order{}, *m.Remediation))
				return cursor
			})

			var i int64 = 0
			for len(cursors) < dbElements {
				i++
				remediation := test.NewFakeRemediationEntity()
				c, _ := mariadb.EncodeCursor(mariadb.WithRemediation([]entity.Order{}, remediation))
				cursors = append(cursors, c)
			}
			db.On("GetRemediations", filter, []entity.Order{}).Return(remediations, nil)
			db.On("GetAllRemediationCursors", filter, []entity.Order{}).Return(cursors, nil)
			remediationHandler = rh.NewRemediationHandler(handlerContext)
			res, err := remediationHandler.ListRemediations(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(*res.PageInfo.HasNextPage).To(BeEquivalentTo(hasNextPage), "correct hasNextPage indicator")
			Expect(len(res.Elements)).To(BeEquivalentTo(resElements))
			Expect(len(res.PageInfo.Pages)).To(BeEquivalentTo(int(math.Ceil(float64(dbElements)/float64(pageSize)))), "correct  number of pages")
		},
			Entry("When  pageSize is 1 and the database was returning 2 elements", 1, 2, 1, true),
			Entry("When  pageSize is 10 and the database was returning 9 elements", 10, 9, 9, false),
			Entry("When  pageSize is 10 and the database was returning 11 elements", 10, 11, 10, true),
		)
	})

	Context("when GetRemediations fails", func() {
		It("should return Internal error", func() {
			// Mock database error
			dbError := errors.New("database connection failed")
			db.On("GetRemediations", filter, []entity.Order{}).Return([]entity.RemediationResult{}, dbError)

			remediationHandler = rh.NewRemediationHandler(handlerContext)
			result, err := remediationHandler.ListRemediations(filter, options)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			// Verify it's our structured error with correct code
			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("Remediations"), "should reference Remediations entity")
			Expect(appErr.ID).To(Equal(""), "should have empty ID for list operation")
			Expect(appErr.Op).To(Equal("remediationHandler.ListRemediations"), "should include operation")
			Expect(appErr.Err.Error()).To(ContainSubstring("database connection failed"), "should contain original error message")
		})
	})

	Context("when GetAllRemediationCursors fails", func() {
		BeforeEach(func() {
			options.ShowPageInfo = true
			filter.First = lo.ToPtr(10)
		})

		It("should return Internal error", func() {
			remediations := []entity.RemediationResult{}
			for _, remediation := range test.NNewFakeRemediations(5) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithRemediation([]entity.Order{}, remediation))
				remediations = append(remediations, entity.RemediationResult{
					WithCursor:  entity.WithCursor{Value: cursor},
					Remediation: lo.ToPtr(remediation),
				})
			}

			db.On("GetRemediations", filter, []entity.Order{}).Return(remediations, nil)
			cursorsError := errors.New("cursor database error")
			db.On("GetAllRemediationCursors", filter, []entity.Order{}).Return([]string{}, cursorsError)

			remediationHandler = rh.NewRemediationHandler(handlerContext)
			result, err := remediationHandler.ListRemediations(filter, options)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("RemediationCursors"), "should reference RemediationCursors entity")
			Expect(appErr.ID).To(Equal(""), "should have empty ID for list operation")
			Expect(appErr.Op).To(Equal("remediationHandler.ListRemediations"), "should include operation")
		})
	})

})

var _ = Describe("When creating Remediation", Label("app", "CreateRemediation"), func() {
	var (
		db                 *mocks.MockDatabase
		remediationHandler rh.RemediationHandler
		remediation        entity.Remediation
		handlerContext     common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		remediation = test.NewFakeRemediationEntity()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	Context("with valid input", func() {
		It("creates remediation", func() {
			db.On("GetAllUserIds", mock.Anything).Return([]int64{123}, nil)
			db.On("CreateRemediation", mock.AnythingOfType("*entity.Remediation")).Return(&remediation, nil)

			remediationHandler = rh.NewRemediationHandler(handlerContext)
			newRemediation, err := remediationHandler.CreateRemediation(common.NewAdminContext(), &remediation)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(newRemediation.Id).NotTo(BeEquivalentTo(0))
			By("setting fields", func() {
				Expect(newRemediation.Description).To(BeEquivalentTo(remediation.Description))
				Expect(newRemediation.Type).To(BeEquivalentTo(remediation.Type))
				Expect(newRemediation.ExpirationDate).To(BeEquivalentTo(remediation.ExpirationDate))
				Expect(newRemediation.RemediationDate).To(BeEquivalentTo(remediation.RemediationDate))
				Expect(newRemediation.Service).To(BeEquivalentTo(remediation.Service))
				Expect(newRemediation.ServiceId).To(BeEquivalentTo(remediation.ServiceId))
				Expect(newRemediation.Component).To(BeEquivalentTo(remediation.Component))
				Expect(newRemediation.ComponentId).To(BeEquivalentTo(remediation.ComponentId))
				Expect(newRemediation.Issue).To(BeEquivalentTo(remediation.Issue))
				Expect(newRemediation.IssueId).To(BeEquivalentTo(remediation.IssueId))
				Expect(newRemediation.RemediatedBy).To(BeEquivalentTo(remediation.RemediatedBy))
				Expect(newRemediation.RemediatedById).To(BeEquivalentTo(remediation.RemediatedById))
			})
		})
	})
})

var _ = Describe("When updating Remediation", Label("app", "UpdateRemediation"), func() {
	var (
		db                 *mocks.MockDatabase
		remediationHandler rh.RemediationHandler
		remediation        entity.RemediationResult
		filter             *entity.RemediationFilter
		handlerContext     common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		remediation = test.NewFakeRemediationResult()
		first := 10
		after := ""
		filter = &entity.RemediationFilter{
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
		It("updates remediation", func() {
			db.On("GetAllUserIds", mock.Anything).Return([]int64{123}, nil)
			db.On("UpdateRemediation", remediation.Remediation).Return(nil)
			remediationHandler = rh.NewRemediationHandler(handlerContext)
			remediation.Description = "Updated description"
			remediation.Service = "Updated Service"
			remediation.Component = "Updated Component"
			remediation.Issue = "Updated Issue"
			filter.Id = []*int64{&remediation.Id}
			db.On("GetRemediations", filter, []entity.Order{}).Return([]entity.RemediationResult{remediation}, nil)
			updatedRemediation, err := remediationHandler.UpdateRemediation(common.NewAdminContext(), remediation.Remediation)
			Expect(err).To(BeNil(), "no error should be thrown")
			By("setting fields", func() {
				Expect(updatedRemediation.Description).To(BeEquivalentTo(remediation.Description))
				Expect(updatedRemediation.Service).To(BeEquivalentTo(remediation.Service))
				Expect(updatedRemediation.Component).To(BeEquivalentTo(remediation.Component))
				Expect(updatedRemediation.Issue).To(BeEquivalentTo(remediation.Issue))
			})
		})
	})
})

var _ = Describe("When deleting Remediation", Label("app", "DeleteRemediation"), func() {
	var (
		db                 *mocks.MockDatabase
		remediationHandler rh.RemediationHandler
		id                 int64
		filter             *entity.RemediationFilter
		handlerContext     common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		after := ""
		filter = &entity.RemediationFilter{
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
		It("deletes remediation", func() {
			db.On("GetAllUserIds", mock.Anything).Return([]int64{123}, nil)
			db.On("DeleteRemediation", id, int64(123)).Return(nil)
			remediationHandler = rh.NewRemediationHandler(handlerContext)
			db.On("GetRemediations", filter, []entity.Order{}).Return([]entity.RemediationResult{}, nil)
			err := remediationHandler.DeleteRemediation(common.NewAdminContext(), id)
			Expect(err).To(BeNil(), "no error should be thrown")

			filter.Id = []*int64{&id}
			lo := entity.NewListOptions()
			remediations, err := remediationHandler.ListRemediations(filter, lo)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(remediations.Elements).To(BeEmpty(), "remediation should be deleted")
		})
	})
})
