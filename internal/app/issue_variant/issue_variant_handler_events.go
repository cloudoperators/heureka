// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_variant

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/entity"
)

const (
	ListIssueVariantsEventName          event.EventName = "ListIssueVariants"
	ListEffectiveIssueVariantsEventName event.EventName = "ListEffectiveIssueVariants"
	CreateIssueVariantEventName         event.EventName = "CreateIssueVariant"
	UpdateIssueVariantEventName         event.EventName = "UpdateIssueVariant"
	DeleteIssueVariantEventName         event.EventName = "DeleteIssueVariant"
)

type ListIssueVariantsEvent struct {
	Filter  *entity.IssueVariantFilter
	Options *entity.ListOptions
	Results *entity.List[entity.IssueVariantResult]
}

func (e *ListIssueVariantsEvent) Name() event.EventName {
	return ListIssueVariantsEventName
}

type ListEffectiveIssueVariantsEvent struct {
	Filter  *entity.IssueVariantFilter
	Options *entity.ListOptions
	Results *entity.List[entity.IssueVariantResult]
}

func (e *ListEffectiveIssueVariantsEvent) Name() event.EventName {
	return ListEffectiveIssueVariantsEventName
}

type CreateIssueVariantEvent struct {
	IssueVariant *entity.IssueVariant
}

func (e *CreateIssueVariantEvent) Name() event.EventName {
	return CreateIssueVariantEventName
}

type UpdateIssueVariantEvent struct {
	IssueVariant *entity.IssueVariant
}

func (e *UpdateIssueVariantEvent) Name() event.EventName {
	return UpdateIssueVariantEventName
}

type DeleteIssueVariantEvent struct {
	IssueVariantID int64
}

func (e *DeleteIssueVariantEvent) Name() event.EventName {
	return DeleteIssueVariantEventName
}
