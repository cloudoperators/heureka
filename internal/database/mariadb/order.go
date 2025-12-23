// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
)

func ColumnName(f entity.OrderByField) string {
	switch f {
	case entity.ComponentId:
		return "component_id"
	case entity.ComponentCcrn:
		return "component_ccrn"
	case entity.ComponentInstanceCcrn:
		return "componentinstance_ccrn"
	case entity.ComponentInstanceId:
		return "componentinstance_id"
	case entity.ComponentInstanceRegion:
		return "componentinstance_region"
	case entity.ComponentInstanceCluster:
		return "componentinstance_cluster"
	case entity.ComponentInstanceNamespace:
		return "componentinstance_namespace"
	case entity.ComponentInstanceDomain:
		return "componentinstance_domain"
	case entity.ComponentInstanceProject:
		return "componentinstance_project"
	case entity.ComponentInstancePod:
		return "componentinstance_pod"
	case entity.ComponentInstanceContainer:
		return "componentinstance_container"
	case entity.ComponentInstanceTypeOrder:
		return "componentinstance_type"
	case entity.ComponentVersionId:
		return "componentversion_id"
	case entity.ComponentVersionRepository:
		return "componentversion_repository"
	case entity.IssueId:
		return "issue_id"
	case entity.IssueVariantRating:
		return "issuevariant_rating_num"
	case entity.IssuePrimaryName:
		return "issue_primary_name"
	case entity.IssueMatchId:
		return "issuematch_id"
	case entity.IssueMatchRating:
		return "issuematch_rating"
	case entity.IssueMatchTargetRemediationDate:
		return "issuematch_target_remediation_date"
	case entity.SupportGroupId:
		return "supportgroup_id"
	case entity.SupportGroupCcrn:
		return "supportgroup_ccrn"
	case entity.ServiceId:
		return "service_id"
	case entity.ServiceCcrn:
		return "service_ccrn"
	case entity.CriticalCount:
		return "critical_count"
	case entity.HighCount:
		return "high_count"
	case entity.MediumCount:
		return "medium_count"
	case entity.LowCount:
		return "low_count"
	case entity.NoneCount:
		return "none_count"
	case entity.RemediationId:
		return "remediation_id"
	case entity.PatchId:
		return "patch_id"
	default:
		return ""
	}
}

func OrderDirectionStr(dir entity.OrderDirection) string {
	switch dir {
	case entity.OrderDirectionAsc:
		return "ASC"
	case entity.OrderDirectionDesc:
		return "DESC"
	default:
		return ""
	}
}

func CreateOrderString(order []entity.Order) string {
	orderStr := ""
	for i, o := range order {
		if i > 0 {
			orderStr = fmt.Sprintf("%s, %s %s", orderStr, ColumnName(o.By), OrderDirectionStr(o.Direction))
		} else {
			orderStr = fmt.Sprintf("%s %s %s", orderStr, ColumnName(o.By), OrderDirectionStr(o.Direction))
		}
	}
	return orderStr
}
