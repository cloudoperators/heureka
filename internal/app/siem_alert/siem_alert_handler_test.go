// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package siem_alert_test

import (
	"testing"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/app/issue"
	"github.com/cloudoperators/heureka/internal/app/issue_match"
	"github.com/cloudoperators/heureka/internal/app/issue_repository"
	"github.com/cloudoperators/heureka/internal/app/issue_variant"
	"github.com/cloudoperators/heureka/internal/app/severity"
	sa "github.com/cloudoperators/heureka/internal/app/siem_alert"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	"github.com/cloudoperators/heureka/internal/openfga"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

func TestSIEMAlertHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SIEMAlert Handler Test Suite")
}

func newHandlers(db *mocks.MockDatabase, er event.EventRegistry, authz openfga.Authorization) (
	issue_match.IssueMatchHandler,
	issue_variant.IssueVariantHandler,
	issue.IssueHandler,
) {
	hc := common.HandlerContext{DB: db, EventReg: er, Authz: authz}
	rh := issue_repository.NewIssueRepositoryHandler(hc)
	ivh := issue_variant.NewIssueVariantHandler(hc, rh)
	sh := severity.NewSeverityHandler(hc, ivh)
	imh := issue_match.NewIssueMatchHandler(hc, sh)
	ih := issue.NewIssueHandler(hc)

	return imh, ivh, ih
}

var _ = Describe("When deleting a SIEMAlert", Label("app", "DeleteSIEMAlert"), func() {
	var (
		db               *mocks.MockDatabase
		er               event.EventRegistry
		authz            openfga.Authorization
		siemAlertHandler sa.SIEMAlertHandler
		hc               common.HandlerContext
		ctx              = common.NewAdminContext()
	)

	BeforeEach(func() {
		authEnabled := false
		cfg := common.GetTestConfig(authEnabled)
		enableLogs := false
		authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, authz)
		hc = common.HandlerContext{DB: db, EventReg: er, Authz: authz}
		imh, ivh, ih := newHandlers(db, er, authz)
		siemAlertHandler = sa.NewSIEMAlertHandler(hc, imh, ivh, ih)
	})

	Context("when the IssueMatch has no other IssueMatches for the same Issue", func() {
		It("deletes the IssueMatch, IssueVariants, and Issue", func() {
			im := test.NewFakeIssueMatch()
			im.IssueId = int64(42)
			imResult := entity.IssueMatchResult{IssueMatch: &im}

			iv := test.NewFakeIssueVariantEntity(&im.IssueId)
			ivResult := entity.IssueVariantResult{IssueVariant: &iv}

			db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
			db.On(
				"GetIssueMatches", mock.Anything,
				mock.MatchedBy(func(f *entity.IssueMatchFilter) bool {
					return len(f.Id) == 1 && *f.Id[0] == im.Id
				}),
				mock.Anything,
			).Return([]entity.IssueMatchResult{imResult}, nil).Once()

			db.On("DeleteIssueMatch", im.Id, mock.Anything).Return(nil)

			db.On(
				"GetIssueMatches", mock.Anything,
				mock.MatchedBy(func(f *entity.IssueMatchFilter) bool {
					return len(f.IssueId) == 1 && *f.IssueId[0] == im.IssueId
				}),
				mock.Anything,
			).Return([]entity.IssueMatchResult{}, nil).Once()

			db.On(
				"GetIssueVariants", mock.Anything,
				mock.MatchedBy(func(f *entity.IssueVariantFilter) bool {
					return len(f.IssueId) == 1 && *f.IssueId[0] == im.IssueId
				}),
				mock.Anything,
			).Return([]entity.IssueVariantResult{ivResult}, nil)

			db.On("DeleteIssueVariant", iv.Id, mock.Anything).Return(nil)

			db.On("DeleteIssue", im.IssueId, mock.Anything).Return(nil)

			err := siemAlertHandler.DeleteSIEMAlert(ctx, im.Id)
			Expect(err).To(BeNil())
		})
	})

	Context("when the Issue still has other IssueMatches", func() {
		It("deletes only the IssueMatch and leaves the Issue intact", func() {
			im := test.NewFakeIssueMatch()
			im.IssueId = int64(99)
			imResult := entity.IssueMatchResult{IssueMatch: &im}

			other := test.NewFakeIssueMatch()
			other.IssueId = im.IssueId
			otherResult := entity.IssueMatchResult{IssueMatch: &other}

			db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
			db.On(
				"GetIssueMatches", mock.Anything,
				mock.MatchedBy(func(f *entity.IssueMatchFilter) bool {
					return len(f.Id) == 1 && *f.Id[0] == im.Id
				}),
				mock.Anything,
			).Return([]entity.IssueMatchResult{imResult}, nil).Once()

			db.On("DeleteIssueMatch", im.Id, mock.Anything).Return(nil)

			db.On(
				"GetIssueMatches", mock.Anything,
				mock.MatchedBy(func(f *entity.IssueMatchFilter) bool {
					return len(f.IssueId) == 1 && *f.IssueId[0] == im.IssueId
				}),
				mock.Anything,
			).Return([]entity.IssueMatchResult{otherResult}, nil).Once()

			err := siemAlertHandler.DeleteSIEMAlert(ctx, im.Id)
			Expect(err).To(BeNil())

			db.AssertNotCalled(GinkgoT(), "DeleteIssue", mock.Anything, mock.Anything)
			db.AssertNotCalled(GinkgoT(), "DeleteIssueVariant", mock.Anything, mock.Anything)
		})
	})

	Context("when the IssueMatch does not exist", func() {
		It("returns an error", func() {
			missingId := int64(9999)

			db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
			db.On(
				"GetIssueMatches", mock.Anything,
				mock.MatchedBy(func(f *entity.IssueMatchFilter) bool {
					return len(f.Id) == 1 && *f.Id[0] == missingId
				}),
				mock.Anything,
			).Return([]entity.IssueMatchResult{}, nil)

			err := siemAlertHandler.DeleteSIEMAlert(ctx, missingId)
			Expect(err).NotTo(BeNil())
		})
	})
})
