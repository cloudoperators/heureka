// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package dataloader

import (
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/vikstrous/dataloadgen"
)

type Loaders struct {
	SeverityByIssueID            *dataloadgen.Loader[int64, *entity.IssueMatch]
	EarliestRemediationByIssueID *dataloadgen.Loader[int64, *entity.IssueMatch]
	SourceURLByIssueID           *dataloadgen.Loader[int64, *entity.IssueVariant]
	VulnCountsByComponentID      *dataloadgen.Loader[int64, *entity.IssueSeverityCounts]
	ComponentVersionByID         *dataloadgen.Loader[int64, *entity.ComponentVersion]
	ServiceByID                  *dataloadgen.Loader[int64, *entity.Service]
	ComponentInstanceByID        *dataloadgen.Loader[int64, *entity.ComponentInstance]
	IssueByID                    *dataloadgen.Loader[int64, *entity.Issue]
	ComponentByID                *dataloadgen.Loader[int64, *entity.Component]
	IssueRepositoryByID          *dataloadgen.Loader[int64, *entity.IssueRepository]
}

func NewLoaders(a app.Heureka) *Loaders {
	return &Loaders{
		SeverityByIssueID:            dataloadgen.NewLoader(newSeverityBatchFn(a)),
		EarliestRemediationByIssueID: dataloadgen.NewLoader(newEarliestRemediationBatchFn(a)),
		SourceURLByIssueID:           dataloadgen.NewLoader(newSourceURLBatchFn(a)),
		VulnCountsByComponentID:      dataloadgen.NewLoader(newVulnCountsBatchFn(a)),
		ComponentVersionByID:         dataloadgen.NewLoader(newComponentVersionByIDBatchFn(a)),
		ServiceByID:                  dataloadgen.NewLoader(newServiceByIDBatchFn(a)),
		ComponentInstanceByID:        dataloadgen.NewLoader(newComponentInstanceByIDBatchFn(a)),
		IssueByID:                    dataloadgen.NewLoader(newIssueByIDBatchFn(a)),
		ComponentByID:                dataloadgen.NewLoader(newComponentByIDBatchFn(a)),
		IssueRepositoryByID:          dataloadgen.NewLoader(newIssueRepositoryByIDBatchFn(a)),
	}
}
