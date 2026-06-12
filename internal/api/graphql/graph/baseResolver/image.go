// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// ImageBaseResolver retrieves a list of images based for a specific service.
// It's designed for the Image List View in the UI
// - Default ordering is by Vulnerability Count (descending) and Repository (ascending)
func ImageBaseResolver(
	app app.Heureka,
	ctx context.Context,
	filter *model.ImageFilter,
	first *int,
	after *string,
) (*model.ImageConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called ImageBaseResolver")

	if filter == nil {
		filter = &model.ImageFilter{}
	}

	f := &entity.ComponentFilter{
		Paginated:   entity.Paginated{First: first, After: after},
		ServiceCCRN: filter.Service,
		Repository:  filter.Repository,
	}

	opt := GetListOptions(requestedFields)
	// Set default ordering by vulnerability counts
	// Secondary ordering by repository name
	if len(f.ServiceCCRN) > 0 {
		opt.Order = append(opt.Order, entity.Order{By: entity.CriticalCount, Direction: entity.OrderDirectionDesc})
		opt.Order = append(opt.Order, entity.Order{By: entity.HighCount, Direction: entity.OrderDirectionDesc})
		opt.Order = append(opt.Order, entity.Order{By: entity.MediumCount, Direction: entity.OrderDirectionDesc})
		opt.Order = append(opt.Order, entity.Order{By: entity.LowCount, Direction: entity.OrderDirectionDesc})
		opt.Order = append(opt.Order, entity.Order{By: entity.NoneCount, Direction: entity.OrderDirectionDesc})
	}

	opt.Order = append(opt.Order, entity.Order{
		By:        entity.ComponentRepository,
		Direction: entity.OrderDirectionAsc,
	})

	components, err := app.ListComponents(ctx, f, opt)
	if err != nil {
		return nil, NewResolverError("ImageBaseResolver", err.Error())
	}

	edges := []*model.ImageEdge{}

	for _, result := range components.Elements {
		image := model.NewImage(result.Component)
		edge := model.ImageEdge{
			Node:   &image,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	// Batch pre-load for nested fields
	needVersions := lo.Contains(requestedFields, "edges.node.versions")
	needVulnCounts := lo.Contains(requestedFields, "edges.node.vulnerabilityCounts")
	needVulnerabilities := containsField(requestedFields, "edges.node.vulnerabilities")

	if (needVersions || needVulnCounts || needVulnerabilities) && len(edges) > 0 {
		componentIDs := make([]int64, 0, len(edges))
		for _, edge := range edges {
			id, err := strconv.ParseInt(edge.Node.ID, 10, 64)
			if err != nil {
				logrus.WithField("id", edge.Node.ID).Warn("ImageBaseResolver: failed to parse component ID for batch preload")
				continue
			}

			componentIDs = append(componentIDs, id)
		}

		var (
			versionsMap    map[int64][]entity.ComponentVersionResult
			issueCountsMap map[int64]entity.IssueSeverityCounts
			vulnsMap       map[int64][]entity.VulnerabilityResult
			vulnCountsMap  map[int64]int
		)

		g, gCtx := errgroup.WithContext(ctx)

		if needVersions {
			g.Go(func() error {
				var err error

				versionsMap, err = app.GetVersionsByComponentIDs(gCtx, componentIDs, filter.Service)
				if err != nil {
					logrus.WithField("error", err).Warn("ImageBaseResolver: batch preload versions failed")
				}

				return nil // don't fail the whole request
			})
		}

		if needVulnCounts {
			g.Go(func() error {
				var err error

				issueCountsMap, err = app.GetIssueCountsByComponentIDs(gCtx, componentIDs, filter.Service)
				if err != nil {
					logrus.WithField("error", err).Warn("ImageBaseResolver: batch preload issue counts failed")
				}

				return nil
			})
		}

		if needVulnerabilities {
			g.Go(func() error {
				var err error

				vulnsMap, err = app.GetVulnerabilitiesByComponentIDs(gCtx, componentIDs)
				if err != nil {
					logrus.WithField("error", err).Warn("ImageBaseResolver: batch preload vulnerabilities failed")
				}

				return nil
			})
			g.Go(func() error {
				var err error

				vulnCountsMap, err = app.GetVulnerabilityCountsByComponentIDs(gCtx, componentIDs)
				if err != nil {
					logrus.WithField("error", err).Warn("ImageBaseResolver: batch preload vulnerability counts failed")
				}

				return nil
			})
		}

		_ = g.Wait()

		// Attach pre-loaded data to image nodes
		for _, edge := range edges {
			id, err := strconv.ParseInt(edge.Node.ID, 10, 64)
			if err != nil {
				continue
			}

			if needVersions && versionsMap != nil {
				versions := versionsMap[id]
				cvEdges := make([]*model.ComponentVersionEdge, 0, len(versions))

				for i := range versions {
					cv := model.NewComponentVersion(versions[i].ComponentVersion)
					cursor := fmt.Sprintf("%d", versions[i].Id)
					cvEdges = append(cvEdges, &model.ComponentVersionEdge{
						Node:   &cv,
						Cursor: &cursor,
					})
				}

				edge.Node.Versions = &model.ComponentVersionConnection{
					TotalCount: len(cvEdges),
					Edges:      cvEdges,
				}
			}

			if needVulnCounts && issueCountsMap != nil {
				if counts, ok := issueCountsMap[id]; ok {
					sc := model.NewSeverityCounts(&counts)
					edge.Node.VulnerabilityCounts = &sc
				} else {
					sc := model.NewSeverityCounts(&entity.IssueSeverityCounts{})
					edge.Node.VulnerabilityCounts = &sc
				}
			}

			if needVulnerabilities && vulnsMap != nil {
				vulns := vulnsMap[id]
				vulnEdges := make([]*model.VulnerabilityEdge, 0, len(vulns))

				for i := range vulns {
					vr := &vulns[i]
					vuln := model.Vulnerability{
						ID:          fmt.Sprintf("%d", vr.IssueID),
						Name:        &vr.PrimaryName,
						Description: &vr.Description,
					}

					if vr.MaxSeverity != "" {
						sevVal, err := model.SeverityValue(vr.MaxSeverity)
						if err == nil {
							vuln.Severity = &sevVal
						}
					}

					if vr.SourceURL != "" {
						vuln.SourceURL = &vr.SourceURL
					}

					if vr.EarliestRemediationDate != nil {
						dateStr := vr.EarliestRemediationDate.Format(time.RFC3339)
						vuln.EarliestTargetRemediationDate = &dateStr
					}

					vulnEdges = append(vulnEdges, &model.VulnerabilityEdge{
						Node:   &vuln,
						Cursor: lo.ToPtr(fmt.Sprintf("%d", vr.IssueID)),
					})
				}

				totalVulnCount := len(vulns)

				if vulnCountsMap != nil {
					if cnt, ok := vulnCountsMap[id]; ok {
						totalVulnCount = cnt
					}
				}

				edge.Node.Vulnerabilities = &model.VulnerabilityConnection{
					TotalCount: totalVulnCount,
					Edges:      vulnEdges,
				}
			}
		}
	}

	totalCount := 0
	if components.TotalCount != nil {
		totalCount = int(*components.TotalCount)
	}

	connection := model.ImageConnection{
		TotalCount: totalCount,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(components.PageInfo),
	}

	if lo.Contains(requestedFields, "counts") {
		icFilter := &entity.ComponentFilter{
			ServiceCCRN: filter.Service,
			Repository:  filter.Repository,
		}

		counts, err := app.GetComponentVulnerabilityCounts(ctx, icFilter)
		if err != nil {
			return nil, NewResolverError("ImageBaseResolver", err.Error())
		}

		var severityCounts model.SeverityCounts
		for _, c := range counts {
			severityCounts.Critical += int(c.Critical)
			severityCounts.High += int(c.High)
			severityCounts.Medium += int(c.Medium)
			severityCounts.Low += int(c.Low)
			severityCounts.None += int(c.None)
			severityCounts.Total += int(c.Total)
		}

		connection.Counts = &severityCounts
	}

	return &connection, nil
}
