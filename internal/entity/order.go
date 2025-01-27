// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "fmt"

type DbColumnName int

const (
	ComponentInstanceCcrn DbColumnName = iota

	IssuePrimaryName

	IssueMatchId
	IssueMatchRating
	IssueMatchTargetRemediationDate

	SupportGroupName
)

var DbColumnNameMap = map[DbColumnName]string{
	ComponentInstanceCcrn:           "componentinstance_ccrn",
	IssuePrimaryName:                "issue_primary_name",
	IssueMatchId:                    "issuematch_id",
	IssueMatchRating:                "issuematch_rating",
	IssueMatchTargetRemediationDate: "issuematch_target_remediation_date",
	SupportGroupName:                "supportgroup_name",
}

func (d DbColumnName) String() string {
	return DbColumnNameMap[d]
}

type OrderDirection int

const (
	OrderDirectionAsc OrderDirection = iota
	OrderDirectionDesc
)

var OrderDirectionMap = map[OrderDirection]string{
	OrderDirectionAsc:  "ASC",
	OrderDirectionDesc: "DESC",
}

func (o OrderDirection) String() string {
	return OrderDirectionMap[o]
}

type Order struct {
	By        DbColumnName
	Direction OrderDirection
}

func CreateOrderMap(order []Order) map[DbColumnName]OrderDirection {
	m := map[DbColumnName]OrderDirection{}
	for _, o := range order {
		m[o.By] = o.Direction
	}
	return m
}

func CreateOrderString(order []Order) string {
	orderStr := ""
	for i, o := range order {
		if i > 0 {
			orderStr = fmt.Sprintf("%s, %s %s", orderStr, o.By, o.Direction)
		} else {
			orderStr = fmt.Sprintf("%s %s %s", orderStr, o.By, o.Direction)
		}
	}
	return orderStr
}
