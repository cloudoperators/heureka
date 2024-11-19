// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"math"
	"time"

	"github.com/cloudoperators/heureka/pkg/util"

	"github.com/onsi/ginkgo/v2/dsl/core"

	"github.com/goark/go-cvss/v3/metric"
	"github.com/sirupsen/logrus"
)

type HeurekaEntity interface {
	Activity |
		ActivityAggregations |
		ActivityHasIssue |
		IssueVariant |
		IssueVariantAggregations |
		BaseIssueRepository |
		IssueRepository |
		IssueRepositoryAggregations |
		ResultList |
		ListOptions |
		PageInfo |
		Paginated |
		Severity |
		Cvss |
		Component |
		ComponentAggregations |
		ComponentInstance |
		ComponentInstanceAggregations |
		ComponentVersion |
		ComponentVersionAggregations |
		Evidence |
		EvidenceAggregations |
		BaseService |
		Service |
		ServiceAggregations |
		ServiceWithAggregations |
		SupportGroup |
		SupportGroupAggregations |
		SupportGroupService |
		SupportGroupUser |
		User |
		UserAggregations |
		IssueWithAggregations |
		IssueAggregations |
		Issue |
		IssueMatch |
		IssueMatchChange |
		HeurekaFilter |
		IssueCount |
		IssueTypeCounts |
		ServiceIssueVariant
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
		IssueRepositoryFilter |
		SeverityFilter
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

func NewListOptions() *ListOptions {
	return &ListOptions{
		ShowTotalCount:      false,
		ShowPageInfo:        false,
		IncludeAggregations: false,
	}
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

func MaxPaginated() Paginated {
	return Paginated{
		First: util.Ptr(math.MaxInt),
	}
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

func NewSeverityFromRating(rating SeverityValues) Severity {
	// These values are based on the CVSS v3.1 specification
	// https://www.first.org/cvss/v3.1/specification-document#Qualitative-Severity-Rating-Scale
	// https://nvd.nist.gov/vuln-metrics/cvss
	// They are the lower bounds of the CVSS Score ranges that correlate to each given Rating
	score := 0.0
	switch rating {
	case SeverityValuesLow:
		score = 0.1
	case SeverityValuesMedium:
		score = 4.0
	case SeverityValuesHigh:
		score = 7.0
	case SeverityValuesCritical:
		score = 9.0
	}

	return Severity{
		Value: string(rating),
		Score: score,
		Cvss:  Cvss{},
	}
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

type Metadata struct {
	CreatedAt time.Time `json:"created_at"`
	CreatedBy int64     `json:"created_by"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy int64     `json:"updated_by"`
	DeletedAt time.Time `json:"deleted_at,omitempty"`
}
