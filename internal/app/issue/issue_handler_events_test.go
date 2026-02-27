// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0
package issue_test

import (
	"github.com/cloudoperators/heureka/internal/app/issue"
	"github.com/cloudoperators/heureka/internal/openfga"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/mock"
)

// matchIssueMatch creates a more precise matcher for IssueMatch objects
func matchIssueMatch(expected *entity.IssueMatch) interface{} {
	return mock.MatchedBy(func(actual *entity.IssueMatch) bool {
		// Basic field matching
		basicMatch := actual.Status == expected.Status &&
			actual.UserId == expected.UserId &&
			actual.ComponentInstanceId == expected.ComponentInstanceId &&
			actual.IssueId == expected.IssueId

		// Severity matching
		severityMatch := actual.Severity.Value == expected.Severity.Value &&
			actual.Severity.Score == expected.Severity.Score

		return basicMatch && severityMatch
	})
}

var _ = Describe("OnComponentVersionAttachmentToIssue", Label("app", "ComponentVersionAttachment"), func() {
	var (
		db                  *mocks.MockDatabase
		componentVersion    entity.ComponentVersion
		issueEntity         entity.Issue
		issueVariant        entity.IssueVariant
		serviceIssueVariant entity.ServiceIssueVariant
		event               *issue.AddComponentVersionToIssueEvent
		authz               openfga.Authorization
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())

		// Setup base test data
		componentVersion = test.NewFakeComponentVersionEntity()

		// Setup issue
		issueEntity = test.NewFakeIssueEntity()
		issueEntity.Id = 625000

		// Setup issueVariant
		issueVariant = test.NewFakeIssueVariantEntity(nil)
		issueVariant.IssueId = issueEntity.Id

		// Setup serviceIssueVariant
		serviceIssueVariant = test.NewFakeServiceIssueVariantEntity(10, &issueEntity.Id)
		serviceIssueVariant.Severity = entity.Severity{
			Value: "Medium",
			Score: 4.5,
			Cvss: entity.Cvss{
				Vector: "CVSS:3.1/AV:N/AC:H/PR:L/UI:N/S:U/C:N/I:H/A:N/E:F/RL:O/RC:U",
			},
		}

		event = &issue.AddComponentVersionToIssueEvent{
			IssueID:            issueEntity.Id,
			ComponentVersionID: componentVersion.Id,
		}
	})

	Context("when handling a single component instance", func() {
		var componentInstance entity.ComponentInstanceResult

		BeforeEach(func() {
			componentInstance = test.NewFakeComponentInstanceResult()
			componentInstance.Id = 58708
			componentInstance.ComponentVersionId = componentVersion.Id

			// // Setup mock expectations for happy path
			db.On("GetComponentInstances", &entity.ComponentInstanceFilter{
				ComponentVersionId: []*int64{&componentVersion.Id},
			}, []entity.Order{}).Return([]entity.ComponentInstanceResult{componentInstance}, nil)
		})

		It("creates an issue match for the component instance", func() {
			// Setup expectation for existing match check
			db.On("GetIssueMatches", &entity.IssueMatchFilter{
				ComponentInstanceId: []*int64{&componentInstance.Id},
				IssueId:             []*int64{&issueEntity.Id},
			}, []entity.Order{}).Return([]entity.IssueMatchResult{}, nil)

			db.On("GetServiceIssueVariants", &entity.ServiceIssueVariantFilter{
				ComponentInstanceId: []*int64{&componentInstance.Id},
				IssueId:             []*int64{&issueEntity.Id},
			}).Return([]entity.ServiceIssueVariant{serviceIssueVariant}, nil)

			expectedMatch := &entity.IssueMatch{
				UserId:              1,
				Status:              entity.IssueMatchStatusValuesNew,
				Severity:            serviceIssueVariant.Severity,
				ComponentInstanceId: componentInstance.Id,
				IssueId:             issueEntity.Id,
			}
			db.On("GetAllUserIds", mock.Anything).Return([]int64{1}, nil)
			db.On("CreateIssueMatch", matchIssueMatch(expectedMatch)).Return(expectedMatch, nil)

			// Emit event
			issue.OnComponentVersionAttachmentToIssue(db, event, authz)

			// Assert expectations
			db.AssertExpectations(GinkgoT())
		})

		It("skips creation if match already exists", func() {
			existingMatch := test.NewFakeIssueMatchResult()
			db.On("GetServiceIssueVariants", &entity.ServiceIssueVariantFilter{
				ComponentInstanceId: []*int64{&componentInstance.Id},
				IssueId:             []*int64{&issueEntity.Id},
			}).Return([]entity.ServiceIssueVariant{serviceIssueVariant}, nil)

			// Setup expectation to return existing match
			db.On("GetIssueMatches", &entity.IssueMatchFilter{
				ComponentInstanceId: []*int64{&componentInstance.Id},
				IssueId:             []*int64{&issueEntity.Id},
			}, []entity.Order{}).Return([]entity.IssueMatchResult{existingMatch}, nil)

			issue.OnComponentVersionAttachmentToIssue(db, event, authz)
			db.AssertNotCalled(GinkgoT(), "CreateIssueMatch", mock.Anything)
		})
	})
})
