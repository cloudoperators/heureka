// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"fmt"
	"math"
	"time"

	interutil "github.com/cloudoperators/heureka/internal/util"
	"github.com/cloudoperators/heureka/pkg/util"

	"github.com/goark/go-cvss/v3/metric"
	"github.com/sirupsen/logrus"
)

type HeurekaEntity interface {
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
		ComponentResult |
		ComponentAggregations |
		ComponentInstance |
		ComponentInstanceResult |
		ComponentInstanceAggregations |
		ComponentVersion |
		ComponentVersionResult |
		ComponentVersionAggregations |
		BaseService |
		Service |
		ServiceAggregations |
		ServiceWithAggregations |
		ServiceResult |
		SupportGroup |
		SupportGroupResult |
		SupportGroupAggregations |
		SupportGroupService |
		SupportGroupUser |
		User |
		UserAggregations |
		IssueWithAggregations |
		IssueAggregations |
		Issue |
		IssueResult |
		IssueMatch |
		IssueMatchResult |
		HeurekaFilter |
		IssueCount |
		IssueTypeCounts |
		IssueSeverityCounts |
		ServiceIssueVariant |
		Remediation |
		RemediationResult |
		Patch |
		PatchResult
}

type HeurekaFilter interface {
	IssueMatchFilter |
		IssueFilter |
		UserFilter |
		SupportGroupFilter |
		ServiceFilter |
		ComponentInstanceFilter |
		TimeFilter |
		IssueVariantFilter |
		ComponentFilter |
		ComponentVersionFilter |
		IssueRepositoryFilter |
		SeverityFilter |
		RemediationFilter |
		PatchFilter
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
	Order               []Order
}

func NewListOptions() *ListOptions {
	return &ListOptions{
		ShowTotalCount:      false,
		ShowPageInfo:        false,
		IncludeAggregations: false,
		Order:               []Order{},
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

type PaginatedX struct {
	First *int    `json:"first"`
	After *string `json:"from"`
}

type HasPagination interface {
	GetPaginatedX() *PaginatedX
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
		logrus.WithField("cvssUrl", url).WithError(err).Warning("Error while parsing CVSS Url.")
	}

	severity := "unknown"
	score := 0.0
	cvss := Cvss{}
	if ev != nil {
		severity = ev.Severity().String()
		score = ev.Score()
		var externalUrl string
		switch ev.Ver {
		case metric.V3_0:
			externalUrl = fmt.Sprintf("https://nvd.nist.gov/vuln-metrics/cvss/v3-calculator?vector=%s&version=3.0", url)
		case metric.V3_1:
			externalUrl = fmt.Sprintf("https://nvd.nist.gov/vuln-metrics/cvss/v3-calculator?vector=%s&version=3.1", url)
		case metric.VUnknown:
			externalUrl = ""
		}
		cvss = Cvss{
			Vector:        url,
			ExternalUrl:   externalUrl,
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
	ExternalUrl   string
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

type StateFilterType int

const (
	Active StateFilterType = iota
	Deleted
)

var StateFilterTypeMap = map[StateFilterType]string{
	Active:  "active",
	Deleted: "deleted",
}

func (sft StateFilterType) String() string {
	return StateFilterTypeMap[sft]
}

type Json map[string]interface{}

func (e Json) String() string {
	return interutil.ConvertJsonToStrNoError((*map[string]interface{})(&e))
}
