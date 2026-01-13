// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package patch_test

import (
	"errors"
	"math"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	ph "github.com/cloudoperators/heureka/internal/app/patch"
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

var er event.EventRegistry
var authz openfga.Authorization

func TestPatchHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Patch Test Suite")
}

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)
})

var _ = Describe("When listing Patches", Label("app", "ListPatches"), func() {
	var (
		db             *mocks.MockDatabase
		patchHandler   ph.PatchHandler
		filter         *entity.PatchFilter
		options        *entity.ListOptions
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = &entity.PatchFilter{
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
			db.On("GetPatches", filter, []entity.Order{}).Return([]entity.PatchResult{}, nil)
			db.On("CountPatches", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			patchHandler = ph.NewPatchHandler(handlerContext)
			res, err := patchHandler.ListPatches(filter, options)
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
			patches := []entity.PatchResult{}
			for _, patch := range test.NNewFakePatches(resElements) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithPatch([]entity.Order{}, patch))
				patches = append(patches, entity.PatchResult{WithCursor: entity.WithCursor{Value: cursor}, Patch: lo.ToPtr(patch)})
			}

			var cursors = lo.Map(patches, func(m entity.PatchResult, _ int) string {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithPatch([]entity.Order{}, *m.Patch))
				return cursor
			})

			var i int64 = 0
			for len(cursors) < dbElements {
				i++
				patch := test.NewFakePatchEntity()
				c, _ := mariadb.EncodeCursor(mariadb.WithPatch([]entity.Order{}, patch))
				cursors = append(cursors, c)
			}
			db.On("GetPatches", filter, []entity.Order{}).Return(patches, nil)
			db.On("GetAllPatchCursors", filter, []entity.Order{}).Return(cursors, nil)
			patchHandler = ph.NewPatchHandler(handlerContext)
			res, err := patchHandler.ListPatches(filter, options)
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
	Context("when GetPatches fails", func() {
		It("should return Internal error", func() {
			// Mock database error
			dbError := errors.New("database connection failed")
			db.On("GetPatches", filter, []entity.Order{}).Return([]entity.PatchResult{}, dbError)

			patchHandler = ph.NewPatchHandler(handlerContext)
			result, err := patchHandler.ListPatches(filter, options)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			// Verify it's our structured error with correct code
			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("Patches"), "should reference Patches entity")
			Expect(appErr.ID).To(Equal(""), "should have empty ID for list operation")
			Expect(appErr.Op).To(Equal("patchHandler.ListPatches"), "should include operation")
			Expect(appErr.Err.Error()).To(ContainSubstring("database connection failed"), "should contain original error message")
		})
	})
	Context("when GetAllPatchCursors fails", func() {
		BeforeEach(func() {
			options.ShowPageInfo = true
			filter.First = lo.ToPtr(10)
		})

		It("should return Internal error", func() {
			patches := []entity.PatchResult{}
			for _, patch := range test.NNewFakePatches(5) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithPatch([]entity.Order{}, patch))
				patches = append(patches, entity.PatchResult{
					WithCursor: entity.WithCursor{Value: cursor},
					Patch:      lo.ToPtr(patch),
				})
			}

			db.On("GetPatches", filter, []entity.Order{}).Return(patches, nil)
			cursorsError := errors.New("cursor database error")
			db.On("GetAllPatchCursors", filter, []entity.Order{}).Return([]string{}, cursorsError)

			patchHandler = ph.NewPatchHandler(handlerContext)
			result, err := patchHandler.ListPatches(filter, options)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("PatchCursors"), "should reference PatchCursors entity")
			Expect(appErr.ID).To(Equal(""), "should have empty ID for list operation")
			Expect(appErr.Op).To(Equal("patchHandler.ListPatches"), "should include operation")
		})
	})

})
