// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"time"

	"github.com/onsi/ginkgo/v2/dsl/core"

	"github.com/goark/go-cvss/v3/metric"
	"github.com/sirupsen/logrus"
)

type HeurekaEntity interface {
	Activity |
		ActivityHasIssue |
		IssueVariant |
		BaseIssueRepository |
		IssueRepository |
		ResultList |
		ListOptions |
		PageInfo |
		Paginated |
		Severity |
		Cvss |
		Component |
		ComponentInstanceAggregations |
		ComponentInstance |
		ComponentVersion |
		Evidence |
		BaseService |
		Service |
		ServiceAggregations |
		SupportGroup |
		SupportGroupService |
		SupportGroupUser |
		User |
		IssueWithAggregations |
		IssueAggregations |
		Issue |
		IssueMatch |
		IssueMatchChange |
		HeurekaFilter
}

type HeurekaFilter interface {
	IssueMatchFilter |
		IssueMatchChangeFilter |
		IssueFilter |
		UserFilter |
		SupportGroupFilter |
		ServiceFilter |
		ComponentInstanceFilter |
		TimeFilter |
		IssueVariantFilter |
		ActivityFilter |
		EvidenceFilter |
		ComponentFilter |
		ComponentVersionFilter |
		IssueRepositoryFilter
}

type HasCursor interface {
	Cursor() *string
}

type WithCursor struct {
	Value string
}

func (c WithCursor) Cursor() *string {
	return &c.Value
}

type ResultList struct {
	TotalCount *int64
	PageInfo   *PageInfo
}

type ListOptions struct {
	ShowTotalCount      bool `json:"show_total_count"`
	ShowPageInfo        bool `json:"show_page_info"`
	IncludeAggregations bool `json:"include_aggregations"`
}

type PageInfo struct {
	HasNextPage     *bool   `json:"has_next_page,omitempty"`
	HasPreviousPage *bool   `json:"has_previous_page,omitempty"`
	IsValidPage     *bool   `json:"is_valid_page,omitempty"`
	PageNumber      *int    `json:"page_number,omitempty"`
	NextPageAfter   *string `json:"next_page_after,omitempty"`
	StartCursor     *string `json:"deprecated,omitempty"` //@todo remove as deprecated
	EndCursor       *string `json:"end_cursor,omitempty"` //@todo remove as deprecated
	Pages           []Page  `json:"pages,omitempty"`
}

type Page struct {
	After      *string `json:"after,omitempty"`
	PageNumber *int    `json:"page_number,omitempty"`
	IsCurrent  bool    `json:"is_current,omitempty"`
	PageCount  *int    `json:"page_count,omitempty"`
}

type List[T interface{}] struct {
	TotalCount *int64
	PageInfo   *PageInfo
	Elements   []T
}

type TimeFilter struct {
	After  time.Time `json:"after"`
	Before time.Time `json:"before"`
}

type Paginated struct {
	First *int   `json:"first"`
	After *int64 `json:"from"`
}

type Severity struct {
	Value string
	Score float64
	Cvss  Cvss
}

type Cursor struct {
	Statement string
	Value     int64
	Limit     int
}

func NewSeverity(url string) Severity {
	ev, err := metric.NewEnvironmental().Decode(url)

	if err != nil {
		core.GinkgoLogr.WithValues("cvssUrl", url, "error", err).Info("Error while parsing CVSS")
		logrus.WithField("cvssUrl", url).WithError(err).Warning("Error while parsing CVSS Url.")
	}

	severity := "unkown"
	score := 0.0
	cvss := Cvss{}
	if ev != nil {
		severity = ev.Severity().String()
		score = ev.Score()
		cvss = Cvss{
			Vector:        url,
			Base:          ev.Base,
			Temporal:      ev.Temporal,
			Environmental: ev,
		}
	}

	return Severity{
		Value: severity,
		Score: score,
		Cvss:  cvss,
	}
}

type Cvss struct {
	Vector        string
	Base          *metric.Base
	Temporal      *metric.Temporal
	Environmental *metric.Environmental
}
