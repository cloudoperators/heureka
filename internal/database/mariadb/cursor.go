// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
)

type Field struct {
	Name  entity.OrderByField
	Value any
	Order entity.OrderDirection
}

type cursors struct {
	fields []Field
}

type NewCursor func(cursors *cursors) error

func EncodeCursor(opts ...NewCursor) (string, error) {
	var cursors cursors
	for _, opt := range opts {
		err := opt(&cursors)
		if err != nil {
			return "", err
		}
	}

	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	err := json.NewEncoder(encoder).Encode(cursors.fields)
	if err != nil {
		return "", err
	}
	encoder.Close()
	return buf.String(), nil
}

func DecodeCursor(cursor *string) ([]Field, error) {
	var fields []Field
	if cursor == nil || *cursor == "" {
		return fields, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(*cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 string: %w", err)
	}

	if err := json.Unmarshal(decoded, &fields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return fields, nil
}

func CreateCursorQuery(query string, fields []Field) string {
	if len(fields) == 0 {
		return query
	}

	subQuery := ""
	for i, f := range fields {
		dir := ">"
		switch f.Order {
		case entity.OrderDirectionAsc:
			dir = ">"
		case entity.OrderDirectionDesc:
			dir = "<"
		}
		if i >= len(fields)-1 {
			subQuery = fmt.Sprintf("%s %s %s ? ", subQuery, ColumnName(f.Name), dir)
		} else {
			subQuery = fmt.Sprintf("%s %s = ? AND ", subQuery, ColumnName(f.Name))
		}
	}

	subQuery = fmt.Sprintf("( %s )", subQuery)
	if query != "" {
		subQuery = fmt.Sprintf("%s OR %s", query, subQuery)
	}

	return CreateCursorQuery(subQuery, fields[:len(fields)-1])
}

func CreateCursorParameters(params []any, fields []Field) []any {
	if len(fields) == 0 {
		return params
	}

	for i := 0; i < len(fields); i++ {
		params = append(params, fields[i].Value)
	}

	return CreateCursorParameters(params, fields[:len(fields)-1])
}

func WithIssueMatch(order []entity.Order, im entity.IssueMatch) NewCursor {

	return func(cursors *cursors) error {
		order = GetDefaultOrder(order, entity.IssueMatchId, entity.OrderDirectionAsc)
		for _, o := range order {
			switch o.By {
			case entity.IssueMatchId:
				cursors.fields = append(cursors.fields, Field{Name: entity.IssueMatchId, Value: im.Id, Order: o.Direction})
			case entity.IssueMatchTargetRemediationDate:
				cursors.fields = append(cursors.fields, Field{Name: entity.IssueMatchTargetRemediationDate, Value: im.TargetRemediationDate, Order: o.Direction})
			case entity.IssueMatchRating:
				cursors.fields = append(cursors.fields, Field{Name: entity.IssueMatchRating, Value: im.Severity.Value, Order: o.Direction})
			case entity.ComponentInstanceCcrn:
				if im.ComponentInstance != nil {
					cursors.fields = append(cursors.fields, Field{Name: entity.ComponentInstanceCcrn, Value: im.ComponentInstance.CCRN, Order: o.Direction})
				}
			case entity.IssuePrimaryName:
				if im.Issue != nil {
					cursors.fields = append(cursors.fields, Field{Name: entity.IssuePrimaryName, Value: im.Issue.PrimaryName, Order: o.Direction})
				}
			default:
				continue
			}
		}

		return nil
	}
}

func WithService(order []entity.Order, s entity.Service, isc entity.IssueSeverityCounts) NewCursor {

	return func(cursors *cursors) error {
		order = GetDefaultOrder(order, entity.ServiceId, entity.OrderDirectionAsc)
		for _, o := range order {
			switch o.By {
			case entity.ServiceId:
				cursors.fields = append(cursors.fields, Field{Name: entity.ServiceId, Value: s.Id, Order: o.Direction})
			case entity.ServiceCcrn:
				cursors.fields = append(cursors.fields, Field{Name: entity.ServiceCcrn, Value: s.CCRN, Order: o.Direction})
			case entity.CriticalCount:
				cursors.fields = append(cursors.fields, Field{Name: entity.CriticalCount, Value: isc.Critical, Order: o.Direction})
			case entity.HighCount:
				cursors.fields = append(cursors.fields, Field{Name: entity.HighCount, Value: isc.High, Order: o.Direction})
			case entity.MediumCount:
				cursors.fields = append(cursors.fields, Field{Name: entity.MediumCount, Value: isc.Medium, Order: o.Direction})
			case entity.LowCount:
				cursors.fields = append(cursors.fields, Field{Name: entity.LowCount, Value: isc.Low, Order: o.Direction})
			case entity.NoneCount:
				cursors.fields = append(cursors.fields, Field{Name: entity.NoneCount, Value: isc.None, Order: o.Direction})
			default:
				continue
			}
		}

		return nil
	}
}

func WithComponentInstance(order []entity.Order, ci entity.ComponentInstance) NewCursor {
	return func(cursors *cursors) error {
		order = GetDefaultOrder(order, entity.ComponentInstanceId, entity.OrderDirectionAsc)
		for _, o := range order {
			switch o.By {
			case entity.ComponentInstanceId:
				cursors.fields = append(cursors.fields, Field{Name: entity.ComponentInstanceId, Value: ci.Id, Order: o.Direction})
			case entity.ComponentInstanceCcrn:
				cursors.fields = append(cursors.fields, Field{Name: entity.ComponentInstanceCcrn, Value: ci.CCRN, Order: o.Direction})
			case entity.ComponentInstanceRegion:
				cursors.fields = append(cursors.fields, Field{Name: entity.ComponentInstanceRegion, Value: ci.Region, Order: o.Direction})
			case entity.ComponentInstanceCluster:
				cursors.fields = append(cursors.fields, Field{Name: entity.ComponentInstanceCluster, Value: ci.Cluster, Order: o.Direction})
			case entity.ComponentInstanceNamespace:
				cursors.fields = append(cursors.fields, Field{Name: entity.ComponentInstanceNamespace, Value: ci.Namespace, Order: o.Direction})
			case entity.ComponentInstanceDomain:
				cursors.fields = append(cursors.fields, Field{Name: entity.ComponentInstanceDomain, Value: ci.Domain, Order: o.Direction})
			case entity.ComponentInstanceProject:
				cursors.fields = append(cursors.fields, Field{Name: entity.ComponentInstanceProject, Value: ci.Project, Order: o.Direction})
			case entity.ComponentInstancePod:
				cursors.fields = append(cursors.fields, Field{Name: entity.ComponentInstancePod, Value: ci.Pod, Order: o.Direction})
			case entity.ComponentInstanceContainer:
				cursors.fields = append(cursors.fields, Field{Name: entity.ComponentInstanceContainer, Value: ci.Container, Order: o.Direction})
			case entity.ComponentInstanceTypeOrder:
				cursors.fields = append(cursors.fields, Field{Name: entity.ComponentInstanceTypeOrder, Value: ci.Type, Order: o.Direction})
			default:
				continue
			}
		}
		return nil
	}
}

func WithComponentVersion(order []entity.Order, cv entity.ComponentVersion, isc entity.IssueSeverityCounts) NewCursor {

	return func(cursors *cursors) error {
		order = GetDefaultOrder(order, entity.ComponentVersionId, entity.OrderDirectionAsc)
		for _, o := range order {
			switch o.By {
			case entity.ComponentVersionId:
				cursors.fields = append(cursors.fields, Field{Name: entity.ComponentVersionId, Value: cv.Id, Order: o.Direction})
			case entity.ComponentVersionRepository:
				cursors.fields = append(cursors.fields, Field{Name: entity.ComponentVersionRepository, Value: cv.Repository, Order: o.Direction})
			case entity.CriticalCount:
				cursors.fields = append(cursors.fields, Field{Name: entity.CriticalCount, Value: isc.Critical, Order: o.Direction})
			case entity.HighCount:
				cursors.fields = append(cursors.fields, Field{Name: entity.HighCount, Value: isc.High, Order: o.Direction})
			case entity.MediumCount:
				cursors.fields = append(cursors.fields, Field{Name: entity.MediumCount, Value: isc.Medium, Order: o.Direction})
			case entity.LowCount:
				cursors.fields = append(cursors.fields, Field{Name: entity.LowCount, Value: isc.Low, Order: o.Direction})
			case entity.NoneCount:
				cursors.fields = append(cursors.fields, Field{Name: entity.NoneCount, Value: isc.None, Order: o.Direction})
			default:
				continue
			}
		}
		return nil
	}
}

func WithIssue(order []entity.Order, issue entity.Issue, ivRating int64) NewCursor {

	return func(cursors *cursors) error {
		order = GetDefaultOrder(order, entity.IssueId, entity.OrderDirectionAsc)
		for _, o := range order {
			switch o.By {
			case entity.IssueId:
				cursors.fields = append(cursors.fields, Field{Name: entity.IssueId, Value: issue.Id, Order: o.Direction})
			case entity.IssuePrimaryName:
				cursors.fields = append(cursors.fields, Field{Name: entity.IssuePrimaryName, Value: issue.PrimaryName, Order: o.Direction})
			case entity.IssueVariantRating:
				cursors.fields = append(cursors.fields, Field{Name: entity.IssueVariantRating, Value: ivRating, Order: o.Direction})
			default:
				continue
			}
		}
		return nil
	}
}

func WithSupportGroup(order []entity.Order, sg entity.SupportGroup) NewCursor {
	return func(cursors *cursors) error {
		order = GetDefaultOrder(order, entity.SupportGroupId, entity.OrderDirectionAsc)
		for _, o := range order {
			switch o.By {
			case entity.SupportGroupId:
				cursors.fields = append(cursors.fields, Field{Name: entity.SupportGroupId, Value: sg.Id, Order: o.Direction})
			case entity.SupportGroupCcrn:
				cursors.fields = append(cursors.fields, Field{Name: entity.SupportGroupCcrn, Value: sg.CCRN, Order: o.Direction})
			default:
				continue
			}
		}
		return nil
	}
}
