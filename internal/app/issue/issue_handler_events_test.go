// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0
package issue_test

import (
	"github.com/cloudoperators/heureka/internal/app/issue"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/mock"
)

// matchIssueMatch ignores TargetRemediationDate when comparing IssueMatches
func matchIssueMatch(expected *entity.IssueMatch) interface{} {
	return mock.MatchedBy(func(actual *entity.IssueMatch) bool {
		return actual.Status == expected.Status &&
			actual.UserId == expected.UserId &&
			actual.ComponentInstanceId == expected.ComponentInstanceId &&
			actual.IssueId == expected.IssueId &&
			actual.Severity == expected.Severity
	})
}

var _ = Describe("OnComponentVersionAttachmentToIssue", Label("app", "ComponentVersionAttachment"), func() {
	var (
		db               *mocks.MockDatabase
		componentVersion entity.ComponentVersion
		issueEntity      entity.Issue
		issueVariant     entity.IssueVariant
		event            *issue.AddComponentVersionToIssueEvent
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())

		// Setup base test data
		componentVersion = test.NewFakeComponentVersionEntity()
		issueEntity = test.NewFakeIssueEntity()
		issueVariant = test.NewFakeIssueVariantEntity(nil)
		issueVariant.IssueId = issueEntity.Id

		event = &issue.AddComponentVersionToIssueEvent{
			IssueID:            issueEntity.Id,
			ComponentVersionID: componentVersion.Id,
		}
	})

	Context("when handling a single component instance", func() {
		var componentInstance entity.ComponentInstance

		BeforeEach(func() {
			componentInstance = test.NewFakeComponentInstanceEntity()
			componentInstance.ComponentVersionId = componentVersion.Id

			// Setup mock expectations for happy path
			db.On("GetComponentInstances", &entity.ComponentInstanceFilter{
				ComponentVersionId: []*int64{&componentVersion.Id},
			}).Return([]entity.ComponentInstance{componentInstance}, nil)

			db.On("GetIssueVariants", &entity.IssueVariantFilter{
				IssueId: []*int64{&issueEntity.Id},
			}).Return([]entity.IssueVariant{issueVariant}, nil)
		})

		It("creates an issue match for the component instance", func() {
			// Setup expectation for existing match check
			db.On("GetIssueMatches", &entity.IssueMatchFilter{
				ComponentInstanceId: []*int64{&componentInstance.Id},
				IssueId:             []*int64{&issueEntity.Id},
			}).Return([]entity.IssueMatch{}, nil)

			expectedMatch := &entity.IssueMatch{
				UserId:              1,
				Status:              entity.IssueMatchStatusValuesNew,
				Severity:            issueVariant.Severity,
				ComponentInstanceId: componentInstance.Id,
				IssueId:             issueEntity.Id,
			}

			db.On("CreateIssueMatch", matchIssueMatch(expectedMatch)).Return(&entity.IssueMatch{}, nil)

			issue.OnComponentVersionAttachmentToIssue(db, event)
			db.AssertExpectations(GinkgoT())
		})

		It("skips creation if match already exists", func() {
			existingMatch := test.NewFakeIssueMatch()

			// Setup expectation to return existing match
			db.On("GetIssueMatches", &entity.IssueMatchFilter{
				ComponentInstanceId: []*int64{&componentInstance.Id},
				IssueId:             []*int64{&issueEntity.Id},
			}).Return([]entity.IssueMatch{existingMatch}, nil)

			issue.OnComponentVersionAttachmentToIssue(db, event)
			db.AssertNotCalled(GinkgoT(), "CreateIssueMatch", mock.Anything)
		})
	})
})
