// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

// BuildIssueVariantMap builds a map of issue id to issue variant for the given filter.
// it does take the first issue_variant with the highest priority for the respective component instance.
// This is archived by utilizing database.GetServiceIssueVariants that does return ALL issue variants for a given
// component instance id together with the priorty and afterwards identifying for each issue the variant with the highest
// priority
//
// Returns a map of issue id to issue variant
func BuildIssueVariantMap(db database.Database, filter *entity.ServiceIssueVariantFilter, componentVersionId int64) (map[int64]entity.ServiceIssueVariant, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":               "BuildIssueVariantMap",
		"componentInstanceID": filter.ComponentInstanceId,
		"componentVersionID":  componentVersionId,
		"issueId":             filter.IssueId,
	})

	// Get Issue Variants based on filter
	issueVariants, err := db.GetServiceIssueVariants(filter, []entity.Order{})
	if err != nil {
		l.WithField("event-step", "FetchIssueVariants").WithError(err).Error("Error while fetching issue variants")
		return nil, fmt.Errorf("Error while fetching issue variants: %w", err)
	}

	// No issue variants found,
	if len(issueVariants) < 1 {
		l.WithField("event-step", "FetchIssueVariants").Error("No issue variants found that are related to the issue repository")
		return nil, fmt.Errorf("No issue variants found that are related to the issue repository")
	}

	// create a map of issue id to variants for easy access
	issueVariantMap := make(map[int64]entity.ServiceIssueVariant)

	for _, variant := range issueVariants {
		if _, ok := issueVariantMap[variant.IssueId]; ok {
			// if there are multiple variants with the same priority on their repositories we take the highest severity one
			// if serverity and score are the same the first occuring issue variant is taken
			if issueVariantMap[variant.IssueId].Priority < variant.Priority {
				issueVariantMap[variant.IssueId] = *variant.ServiceIssueVariant
			} else if issueVariantMap[variant.IssueId].Severity.Score < variant.Severity.Score {
				issueVariantMap[variant.IssueId] = *variant.ServiceIssueVariant
			}
		} else {
			issueVariantMap[variant.IssueId] = *variant.ServiceIssueVariant
		}
	}

	return issueVariantMap, nil
}
