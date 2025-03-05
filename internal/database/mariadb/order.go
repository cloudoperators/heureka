// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
)

func ColumnName(f entity.OrderByField) string {
	switch f {
	case entity.ComponentInstanceCcrn:
		return "componentinstance_ccrn"
	case entity.IssuePrimaryName:
		return "issue_primary_name"
	case entity.IssueMatchId:
		return "issuematch_id"
	case entity.IssueMatchRating:
		return "issuematch_rating"
	case entity.IssueMatchTargetRemediationDate:
		return "issuematch_target_remediation_date"
	case entity.SupportGroupName:
		return "supportgroup_name"
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
