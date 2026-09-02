// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package comment_test

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/comment"
	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/mocks"
	"github.com/cloudoperators/heureka/internal/openfga"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
)

var (
	er    event.EventRegistry
	authz openfga.Authorization
)

func TestCommentHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Comment Test Suite")
}

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db, authz)
})

var _ = Describe("When listing Comments", Label("app", "ListComments"), func() {
	var (
		db             *mocks.MockDatabase
		commentHandler comment.CommentHandler
		filter         *entity.CommentFilter
		options        *entity.ListOptions
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = &entity.CommentFilter{
			Paginated: entity.Paginated{
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

	When("the list option includes the totalCount", func() {
		BeforeEach(func() {
			options.ShowTotalCount = true

			db.On("GetComments", mock.Anything, filter, []entity.Order{}).Return([]entity.CommentResult{}, nil)
			db.On("CountComments", mock.Anything, filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			commentHandler = comment.NewCommentHandler(handlerContext)
			res, err := commentHandler.ListComments(context.Background(), filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(*res.TotalCount).Should(BeEquivalentTo(int64(1337)), "return correct TotalCount")
		})
	})

	When("the list option includes the PageInfo", func() {
		BeforeEach(func() {
			options.ShowPageInfo = true
		})

		DescribeTable(
			"pagination information is correct",
			func(pageSize int, dbElements int, resElements int, hasNextPage bool) {
				filter.First = &pageSize
				comments := []entity.CommentResult{}

				for _, c := range test.NNewFakeComments(resElements) {
					cursor, _ := mariadb.EncodeCursor(mariadb.WithComment(c))
					comments = append(comments, entity.CommentResult{
						WithCursor: entity.WithCursor{Value: cursor},
						Comment:    new(c),
					})
				}

				cursors := lo.Map(comments, func(m entity.CommentResult, _ int) string {
					cursor, _ := mariadb.EncodeCursor(mariadb.WithComment(*m.Comment))
					return cursor
				})

				for len(cursors) < dbElements {
					c := test.NewFakeCommentEntity()
					cur, _ := mariadb.EncodeCursor(mariadb.WithComment(c))
					cursors = append(cursors, cur)
				}

				db.On("GetComments", mock.Anything, filter, []entity.Order{}).Return(comments, nil)
				db.On("GetAllCommentCursors", mock.Anything, filter, []entity.Order{}).Return(cursors, nil)

				commentHandler = comment.NewCommentHandler(handlerContext)
				res, err := commentHandler.ListComments(context.Background(), filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(*res.PageInfo.HasNextPage).To(BeEquivalentTo(hasNextPage), "correct hasNextPage indicator")
				Expect(len(res.Elements)).To(BeEquivalentTo(resElements))
				Expect(len(res.PageInfo.Pages)).To(BeEquivalentTo(int(math.Ceil(float64(dbElements)/float64(pageSize)))), "correct number of pages")
			},
			Entry("When pageSize is 1 and the database returns 2 elements", 1, 2, 1, true),
			Entry("When pageSize is 10 and the database returns 9 elements", 10, 9, 9, false),
			Entry("When pageSize is 10 and the database returns 11 elements", 10, 11, 10, true),
		)
	})

	Context("when GetComments fails", func() {
		It("should return an Internal error", func() {
			dbError := errors.New("database connection failed")
			db.On("GetComments", mock.Anything, filter, []entity.Order{}).Return([]entity.CommentResult{}, dbError)

			commentHandler = comment.NewCommentHandler(handlerContext)
			result, err := commentHandler.ListComments(context.Background(), filter, options)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("Comments"), "should reference Comments entity")
			Expect(appErr.ID).To(Equal(""), "should have empty ID for list operation")
			Expect(appErr.Op).To(Equal("commentHandler.ListComments"), "should include operation")
			Expect(appErr.Err.Error()).To(ContainSubstring("database connection failed"), "should contain original error")
		})
	})

	Context("when GetAllCommentCursors fails", func() {
		BeforeEach(func() {
			options.ShowPageInfo = true
			filter.First = new(10)
		})

		It("should return an Internal error", func() {
			comments := []entity.CommentResult{}

			for _, c := range test.NNewFakeComments(5) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithComment(c))
				comments = append(comments, entity.CommentResult{
					WithCursor: entity.WithCursor{Value: cursor},
					Comment:    new(c),
				})
			}

			db.On("GetComments", mock.Anything, filter, []entity.Order{}).Return(comments, nil)

			cursorsError := errors.New("cursor database error")
			db.On("GetAllCommentCursors", mock.Anything, filter, []entity.Order{}).Return([]string{}, cursorsError)

			commentHandler = comment.NewCommentHandler(handlerContext)
			result, err := commentHandler.ListComments(context.Background(), filter, options)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("CommentCursors"), "should reference CommentCursors entity")
			Expect(appErr.ID).To(Equal(""), "should have empty ID for list operation")
			Expect(appErr.Op).To(Equal("commentHandler.ListComments"), "should include operation")
		})
	})
})

var _ = Describe("When creating a Comment", Label("app", "CreateComment"), func() {
	var (
		db             *mocks.MockDatabase
		commentHandler comment.CommentHandler
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	Context("when the database call succeeds", func() {
		It("returns the created comment", func() {
			newComment := test.NewFakeCommentEntity()
			newComment.CreatedBy = 0
			newComment.UpdatedBy = 0

			db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
			db.On("CreateComment", mock.MatchedBy(func(c *entity.Comment) bool {
				return c.IssueMatchId == newComment.IssueMatchId && c.Text == newComment.Text
			})).Return(&newComment, nil)

			commentHandler = comment.NewCommentHandler(handlerContext)
			result, err := commentHandler.CreateComment(context.Background(), &newComment)

			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(result).ToNot(BeNil(), "result should be returned")
			Expect(result.Text).To(Equal(newComment.Text), "text should match")
			Expect(result.IssueMatchId).To(Equal(newComment.IssueMatchId), "issueMatchId should match")
		})
	})

	Context("when the database call fails", func() {
		It("should return an Internal error", func() {
			newComment := test.NewFakeCommentEntity()

			db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
			db.On("CreateComment", mock.Anything).Return((*entity.Comment)(nil), errors.New("database connection failed"))

			commentHandler = comment.NewCommentHandler(handlerContext)
			result, err := commentHandler.CreateComment(context.Background(), &newComment)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")
			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("Comment"), "should reference Comment entity")
		})
	})
})
