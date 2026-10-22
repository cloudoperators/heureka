// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"encoding/json"
	"errors"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/baseResolver"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app/remediation"
	"github.com/cloudoperators/heureka/internal/app/siem_alert"
)

// remediationResolverError maps a domain error from the remediation
// orchestration back to the exact client-facing resolver error. The transport
// layer owns the message wording; the domain only reports which reference
// failed. verb is "creating" or "updating".
func remediationResolverError(op, verb string, err error) error {
	msg := "Internal Error - when " + verb + " remediation"

	var refErr *remediation.RemediationReferenceError
	if errors.As(err, &refErr) {
		msg += " - " + remediationReferenceDetail(refErr)
	}

	return baseResolver.NewResolverError(op, msg)
}

func remediationReferenceDetail(e *remediation.RemediationReferenceError) string {
	switch e.Reference {
	case remediation.RemediationReferenceService:
		return "service id not found"
	case remediation.RemediationReferenceIssue:
		return "issue id not found"
	case remediation.RemediationReferenceUser:
		return "user id not found"
	case remediation.RemediationReferenceComponent:
		if e.Ambiguous {
			return model.ComponentFriendlyName(e.ComponentType) + " not found"
		}

		return "component not found"
	default:
		return ""
	}
}

// allSIEMFieldsPresent returns false when any of the six required
// component-instance fields is nil or empty.
func allSIEMFieldsPresent(input model.SIEMAlertInput) bool {
	present := func(s *string) bool { return s != nil && *s != "" }

	return present(input.Service) && present(input.Region) && present(input.Cluster) &&
		present(input.Namespace) && present(input.Pod) && present(input.Container)
}

// toSIEMOrchestrationInput converts the transport model input to the domain
// type the SIEM orchestration accepts.
func toSIEMOrchestrationInput(input model.SIEMAlertInput) siem_alert.SIEMAlertInput {
	var severity *string

	if input.Severity != nil {
		s := input.Severity.String()
		severity = &s
	}

	links := make([]siem_alert.SIEMLink, 0, len(input.Links))
	for _, l := range input.Links {
		if l != nil {
			links = append(links, siem_alert.SIEMLink{Name: l.Name, URL: l.URL})
		}
	}

	return siem_alert.SIEMAlertInput{
		Service:      input.Service,
		SupportGroup: input.SupportGroup,
		Region:       input.Region,
		Cluster:      input.Cluster,
		Namespace:    input.Namespace,
		Pod:          input.Pod,
		Container:    input.Container,
		Name:         input.Name,
		Description:  input.Description,
		Severity:     severity,
		Links:        links,
		Source:       input.Source,
	}
}

// buildSIEMAlertModel assembles the model.SIEMAlert from the domain result
// and original input. Fields derived from persisted entities (issueVariant,
// componentInstance, service) take precedence over the raw input values.
func buildSIEMAlertModel(input model.SIEMAlertInput, result siem_alert.SIEMAlertResult) model.SIEMAlert {
	var name *string
	if result.Issue != nil {
		name = &result.Issue.PrimaryName
	}

	var description *string
	if result.IssueVariant != nil && result.IssueVariant.Description != "" {
		description = &result.IssueVariant.Description
	} else {
		description = input.Description
	}

	var severity *model.SeverityValues
	if result.IssueVariant != nil && result.IssueVariant.Severity.Value != "" {
		severity = new(model.SeverityValues(result.IssueVariant.Severity.Value))
	} else {
		severity = input.Severity
	}

	var links []*model.SIEMAlertLink
	if result.IssueVariant != nil && result.IssueVariant.ExternalUrl != "" {
		if err := json.Unmarshal([]byte(result.IssueVariant.ExternalUrl), &links); err != nil {
			links = []*model.SIEMAlertLink{{Name: result.IssueVariant.ExternalUrl, URL: result.IssueVariant.ExternalUrl}}
		}
	} else {
		for _, l := range input.Links {
			if l != nil {
				links = append(links, &model.SIEMAlertLink{Name: l.Name, URL: l.URL})
			}
		}
	}

	var servicePtr *string
	if result.Service != nil {
		servicePtr = new(result.Service.CCRN)
	} else {
		servicePtr = input.Service
	}

	var supportGroupPtr *string
	if result.SupportGroup != nil {
		supportGroupPtr = new(result.SupportGroup.CCRN)
	} else {
		supportGroupPtr = input.SupportGroup
	}

	var regionPtr, clusterPtr, namespacePtr, podPtr, containerPtr *string

	if result.ComponentInstance != nil {
		ci := result.ComponentInstance
		if ci.Region != "" {
			regionPtr = new(ci.Region)
		}

		if ci.Cluster != "" {
			clusterPtr = new(ci.Cluster)
		}

		if ci.Namespace != "" {
			namespacePtr = new(ci.Namespace)
		}

		if ci.Pod != "" {
			podPtr = new(ci.Pod)
		}

		if ci.Container != "" {
			containerPtr = new(ci.Container)
		}
	}

	if regionPtr == nil {
		regionPtr = input.Region
	}

	if clusterPtr == nil {
		clusterPtr = input.Cluster
	}

	if namespacePtr == nil {
		namespacePtr = input.Namespace
	}

	if podPtr == nil {
		podPtr = input.Pod
	}

	if containerPtr == nil {
		containerPtr = input.Container
	}

	return model.SIEMAlert{
		Name:         name,
		Description:  description,
		Severity:     severity,
		Links:        links,
		Service:      servicePtr,
		SupportGroup: supportGroupPtr,
		Region:       regionPtr,
		Cluster:      clusterPtr,
		Namespace:    namespacePtr,
		Pod:          podPtr,
		Container:    containerPtr,
		Source:       input.Source,
	}
}

// siemResolverError wraps a domain error from the SIEM orchestration with the
// resolver prefix. SIEMError.Message already contains the part after the
// prefix so the exact client string is preserved.
func siemResolverError(err error) error {
	var siemErr *siem_alert.SIEMError
	if errors.As(err, &siemErr) {
		return baseResolver.NewResolverError("CreateSIEMAlertMutationResolver", siemErr.Message)
	}

	return baseResolver.NewResolverError("CreateSIEMAlertMutationResolver", err.Error())
}
