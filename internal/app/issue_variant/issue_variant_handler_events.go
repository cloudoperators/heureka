// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_variant

import (
	"errors"

	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/sirupsen/logrus"
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

// OnIssueVariantCreateAuthz is a handler for the CreateIssueVariantEvent
// It creates an OpenFGA relation tuple for the issue variant and the current user,
// and links the issue variant to its parent issue repository
func OnIssueVariantCreateAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.CallerOp()

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnIssueVariantCreateAuthz",
		"payload": e,
	})

	if createEvent, ok := e.(*CreateIssueVariantEvent); ok {
		userId := openfga.UserIdFromInt(createEvent.IssueVariant.CreatedBy)

		relations := []openfga.RelationInput{
			{
				UserType:   openfga.TypeRole,
				UserId:     userId,
				Relation:   openfga.RelRole,
				ObjectType: openfga.TypeIssueVariant,
				ObjectId:   openfga.ObjectIdFromInt(createEvent.IssueVariant.Id),
			},
			{
				UserType:   openfga.TypeIssueRepository,
				UserId:     openfga.UserIdFromInt(createEvent.IssueVariant.IssueRepositoryId),
				Relation:   openfga.RelIssueRepository,
				ObjectType: openfga.TypeIssueVariant,
				ObjectId:   openfga.ObjectIdFromInt(createEvent.IssueVariant.Id),
			},
		}

		err := authz.AddRelationBulk(relations)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "IssueVariant", "", err)
			l.Error(wrappedErr)
		}
	} else {
		err := errors.New("OnIssueVariantCreateAuthz: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "IssueVariant", "", err)
		l.Error(wrappedErr)
	}
}

// OnIssueVariantDeleteAuthz is a handler for the DeleteIssueVariantEvent
// It deletes all OpenFGA relation tuples containing that issue variant
func OnIssueVariantDeleteAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.CallerOp()

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnIssueVariantDeleteAuthz",
		"payload": e,
	})

	if deleteEvent, ok := e.(*DeleteIssueVariantEvent); ok {
		deleteInput := []openfga.RelationInput{
			{
				ObjectType: openfga.TypeIssueVariant,
				ObjectId:   openfga.ObjectIdFromInt(deleteEvent.IssueVariantID),
			},
		}

		err := authz.RemoveRelationBulk(deleteInput)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "IssueVariant", "", err)
			l.Error(wrappedErr)
		}
	} else {
		err := errors.New("OnIssueVariantDeleteAuthz: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "IssueVariant", "", err)
		l.Error(wrappedErr)
	}
}
