// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package siem_alert_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/comment"
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

type mockCommentHandler struct {
	mock.Mock
}

func (m *mockCommentHandler) ListComments(ctx context.Context, filter *entity.CommentFilter, options *entity.ListOptions) (*entity.List[entity.CommentResult], error) {
	args := m.Called(ctx, filter, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entity.List[entity.CommentResult]), args.Error(1)
}

func (m *mockCommentHandler) CreateComment(ctx context.Context, c *entity.Comment) (*entity.Comment, error) {
	args := m.Called(ctx, c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entity.Comment), args.Error(1)
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
		ch               comment.CommentHandler
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
		ch = &mockCommentHandler{}
		imh, ivh, ih := newHandlers(db, er, authz)
		siemAlertHandler = sa.NewSIEMAlertHandler(hc, imh, ivh, ih, ch)
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

var _ = Describe("When acknowledging a SIEMAlert", Label("app", "AcknowledgeSIEMAlert"), func() {
	var (
		db               *mocks.MockDatabase
		er               event.EventRegistry
		authz            openfga.Authorization
		siemAlertHandler sa.SIEMAlertHandler
		hc               common.HandlerContext
		ch               comment.CommentHandler
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
		ch = &mockCommentHandler{}
		imh, ivh, ih := newHandlers(db, er, authz)
		siemAlertHandler = sa.NewSIEMAlertHandler(hc, imh, ivh, ih, ch)
	})

	isSecurityEventFilter := func(id int64) func(*entity.IssueMatchFilter) bool {
		return func(f *entity.IssueMatchFilter) bool {
			return len(f.Id) == 1 && *f.Id[0] == id &&
				len(f.IssueType) == 1 && *f.IssueType[0] == string(entity.IssueTypeSecurityEvent)
		}
	}

	Context("when the IssueMatch exists and is a SecurityEvent", func() {
		It("returns the acknowledged IssueMatch", func() {
			im := test.NewFakeIssueMatch()
			imResult := entity.IssueMatchResult{IssueMatch: &im}

			updatedIm := im
			updatedIm.Acknowledged = true
			updatedImResult := entity.IssueMatchResult{IssueMatch: &updatedIm}

			db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
			db.On(
				"GetIssueMatches", mock.Anything,
				mock.MatchedBy(isSecurityEventFilter(im.Id)),
				mock.Anything,
			).Return([]entity.IssueMatchResult{imResult}, nil).Once()

			db.On("UpdateIssueMatch", mock.Anything).Return(nil)

			db.On(
				"GetIssueMatches", mock.Anything,
				mock.MatchedBy(func(f *entity.IssueMatchFilter) bool {
					return len(f.Id) == 1 && *f.Id[0] == im.Id
				}),
				mock.Anything,
			).Return([]entity.IssueMatchResult{updatedImResult}, nil).Once()

			result, err := siemAlertHandler.AcknowledgeSIEMAlert(ctx, im.Id)
			Expect(err).To(BeNil())
			Expect(result.Acknowledged).To(BeTrue())
		})
	})

	Context("when the IssueMatch does not exist", func() {
		It("returns an error", func() {
			missingId := int64(9999)

			db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
			db.On(
				"GetIssueMatches", mock.Anything,
				mock.MatchedBy(isSecurityEventFilter(missingId)),
				mock.Anything,
			).Return([]entity.IssueMatchResult{}, nil)

			result, err := siemAlertHandler.AcknowledgeSIEMAlert(ctx, missingId)
			Expect(err).NotTo(BeNil())
			Expect(result).To(BeNil())
		})
	})

	Context("when the IssueMatch exists but is not a SecurityEvent", func() {
		It("returns an error", func() {
			im := test.NewFakeIssueMatch()

			db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
			// The IssueType-filtered query returns nothing because the record is not a SecurityEvent.
			// The handler treats this identically to a missing ID — both result in an empty filtered list.
			db.On(
				"GetIssueMatches", mock.Anything,
				mock.MatchedBy(isSecurityEventFilter(im.Id)),
				mock.Anything,
			).Return([]entity.IssueMatchResult{}, nil)

			result, err := siemAlertHandler.AcknowledgeSIEMAlert(ctx, im.Id)
			Expect(err).NotTo(BeNil())
			Expect(result).To(BeNil())
		})
	})

	Context("when UpdateIssueMatch fails", func() {
		It("returns an error", func() {
			im := test.NewFakeIssueMatch()
			imResult := entity.IssueMatchResult{IssueMatch: &im}

			db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
			db.On(
				"GetIssueMatches", mock.Anything,
				mock.MatchedBy(isSecurityEventFilter(im.Id)),
				mock.Anything,
			).Return([]entity.IssueMatchResult{imResult}, nil)

			db.On("UpdateIssueMatch", mock.Anything).Return(errors.New("db error"))

			result, err := siemAlertHandler.AcknowledgeSIEMAlert(ctx, im.Id)
			Expect(err).NotTo(BeNil())
			Expect(result).To(BeNil())
		})
	})
})

var _ = Describe("When updating a SIEMAlert", Label("app", "UpdateSIEMAlert"), func() {
	var (
		db               *mocks.MockDatabase
		er               event.EventRegistry
		authz            openfga.Authorization
		siemAlertHandler sa.SIEMAlertHandler
		hc               common.HandlerContext
		ch               *mockCommentHandler
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
		ch = &mockCommentHandler{}
		imh, ivh, ih := newHandlers(db, er, authz)
		siemAlertHandler = sa.NewSIEMAlertHandler(hc, imh, ivh, ih, ch)
	})

	Context("when updating the status with a comment", func() {
		It("updates the IssueMatch status and creates a comment", func() {
			im := test.NewFakeIssueMatch()
			imResult := entity.IssueMatchResult{IssueMatch: &im}
			newStatus := entity.IssueMatchStatusValuesMitigated

			imUpdated := im
			imUpdated.Status = newStatus
			imUpdatedResult := entity.IssueMatchResult{IssueMatch: &imUpdated}

			db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
			db.On(
				"GetIssueMatches", mock.Anything,
				mock.MatchedBy(func(f *entity.IssueMatchFilter) bool {
					return len(f.Id) == 1 && *f.Id[0] == im.Id
				}),
				mock.Anything,
			).Return([]entity.IssueMatchResult{imResult}, nil).Once()

			db.On("UpdateIssueMatch", mock.MatchedBy(func(updated *entity.IssueMatch) bool {
				return updated.Id == im.Id && updated.Status == newStatus
			})).Return(nil)

			db.On(
				"GetIssueMatches", mock.Anything,
				mock.MatchedBy(func(f *entity.IssueMatchFilter) bool {
					return len(f.Id) == 1 && *f.Id[0] == im.Id
				}),
				mock.Anything,
			).Return([]entity.IssueMatchResult{imUpdatedResult}, nil).Once()

			expectedComment := &entity.Comment{IssueMatchId: im.Id, Text: "triaged by analyst"}
			ch.On("CreateComment", mock.Anything, expectedComment).Return(expectedComment, nil)

			input := entity.UpdateIssueMatchInput{
				Status:  &newStatus,
				Comment: "triaged by analyst",
			}

			updated, err := siemAlertHandler.UpdateSIEMAlert(ctx, im.Id, input)
			Expect(err).To(BeNil())
			Expect(updated).NotTo(BeNil())
			Expect(updated.Status).To(Equal(newStatus))
			ch.AssertCalled(GinkgoT(), "CreateComment", mock.Anything, expectedComment)
		})
	})

	Context("when updating the assignee with a comment", func() {
		It("updates the IssueMatch userId and creates a comment", func() {
			im := test.NewFakeIssueMatch()
			imResult := entity.IssueMatchResult{IssueMatch: &im}
			newUserId := int64(77)

			imUpdated := im
			imUpdated.UserId = newUserId
			imUpdatedResult := entity.IssueMatchResult{IssueMatch: &imUpdated}

			db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
			db.On(
				"GetIssueMatches", mock.Anything,
				mock.MatchedBy(func(f *entity.IssueMatchFilter) bool {
					return len(f.Id) == 1 && *f.Id[0] == im.Id
				}),
				mock.Anything,
			).Return([]entity.IssueMatchResult{imResult}, nil).Once()

			db.On("UpdateIssueMatch", mock.MatchedBy(func(updated *entity.IssueMatch) bool {
				return updated.Id == im.Id && updated.UserId == newUserId
			})).Return(nil)

			db.On(
				"GetIssueMatches", mock.Anything,
				mock.MatchedBy(func(f *entity.IssueMatchFilter) bool {
					return len(f.Id) == 1 && *f.Id[0] == im.Id
				}),
				mock.Anything,
			).Return([]entity.IssueMatchResult{imUpdatedResult}, nil).Once()

			expectedComment := &entity.Comment{IssueMatchId: im.Id, Text: "assigned to user 77"}
			ch.On("CreateComment", mock.Anything, expectedComment).Return(expectedComment, nil)

			input := entity.UpdateIssueMatchInput{
				UserId:  &newUserId,
				Comment: "assigned to user 77",
			}

			updated, err := siemAlertHandler.UpdateSIEMAlert(ctx, im.Id, input)
			Expect(err).To(BeNil())
			Expect(updated).NotTo(BeNil())
			Expect(updated.UserId).To(Equal(newUserId))
			ch.AssertCalled(GinkgoT(), "CreateComment", mock.Anything, expectedComment)
		})
	})

	Context("when the comment is empty", func() {
		It("returns an error without touching the database", func() {
			input := entity.UpdateIssueMatchInput{
				Comment: "",
			}

			updated, err := siemAlertHandler.UpdateSIEMAlert(ctx, int64(1), input)
			Expect(err).NotTo(BeNil())
			Expect(updated).To(BeNil())
			db.AssertNotCalled(GinkgoT(), "GetIssueMatches", mock.Anything, mock.Anything, mock.Anything)
			db.AssertNotCalled(GinkgoT(), "UpdateIssueMatch", mock.Anything)
			ch.AssertNotCalled(GinkgoT(), "CreateComment", mock.Anything, mock.Anything)
		})
	})

	Context("when the SIEM alert does not exist", func() {
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

			input := entity.UpdateIssueMatchInput{
				Comment: "some comment",
			}

			updated, err := siemAlertHandler.UpdateSIEMAlert(ctx, missingId, input)
			Expect(err).NotTo(BeNil())
			Expect(updated).To(BeNil())
			ch.AssertNotCalled(GinkgoT(), "CreateComment", mock.Anything, mock.Anything)
		})
	})
})
