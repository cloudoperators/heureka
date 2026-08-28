// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"context"
	"fmt"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/cloudoperators/heureka/pkg/util"
	"github.com/cloudoperators/heureka/scanner/nvd/client"
	"github.com/cloudoperators/heureka/scanner/nvd/models"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type Processor struct {
	GraphqlClient       graphql.Client
	IssueRepositoryName string
	IssueRepositoryId   string
	IssueRepositoryUrl  string
	CveDetailsUrl       string
}

// NewProcessor
func NewProcessor(cfg Config) *Processor {
	httpClient := util.NewRateLimitedHTTPClient(
		rate.Limit(cfg.HeurekaRateLimit),
		cfg.HeurekaRateBurst,
		nil,
	)
	httpClient.Timeout = time.Duration(10) * time.Second
	return &Processor{
		GraphqlClient:       graphql.NewClient(cfg.HeurekaUrl, httpClient),
		IssueRepositoryName: cfg.IssueRepositoryName,
		IssueRepositoryUrl:  cfg.IssueRepositoryUrl,
		CveDetailsUrl:       cfg.CveDetailsUrl,
	}
}

func (p *Processor) Setup() error {
	// Check if there is already an IssueRepository with the same name
	queryFilter := client.IssueRepositoryFilter{
		Name: []string{p.IssueRepositoryName},
	}
	listRepositoriesResp, err := client.GetIssueRepositories(
		context.TODO(),
		p.GraphqlClient,
		&queryFilter,
	)
	if err != nil {
		return err
	}

	if listRepositoriesResp.IssueRepositories.TotalCount == 0 {
		log.Warnf("There is no IssueRepository: %s", err)

		// Create new IssueRepository
		issueRepositoryInput := client.IssueRepositoryInput{
			Name: p.IssueRepositoryName,
			Url:  p.IssueRepositoryUrl,
		}
		issueMutationResp, err := client.CreateIssueRepository(
			context.TODO(),
			p.GraphqlClient,
			&issueRepositoryInput,
		)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Couldn't create new IssueRepository")
		}

		// Save IssueRepositoryId
		p.IssueRepositoryId = issueMutationResp.CreateIssueRepository.Id
		log.WithFields(log.Fields{
			"issueRepositoryId": p.IssueRepositoryId,
		}).Info("Created new IssueRepository")
	} else {
		// Extract IssueRepositoryId
		for _, ir := range listRepositoriesResp.IssueRepositories.Edges {
			log.Debugf("nodeId: %s", ir.Node.Id)
			p.IssueRepositoryId = ir.Node.Id
			break
		}
		log.Debugf("IssueRepositoryId: %s", p.IssueRepositoryId)
	}
	return nil
}

func (p *Processor) Process(ctx context.Context, cve *models.Cve) error {
	var issueId string

	// Create new Issue
	createIssueInput := client.IssueInput{
		PrimaryName: cve.Id,
		Description: cve.GetDescription("en"),
		Type:        "Vulnerability",
	}
	issueMutationResp, err := client.CreateIssue(ctx, p.GraphqlClient, &createIssueInput)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Couldn't create new Issue")
		return fmt.Errorf("Couldn't create new Issue")
	}

	issueId = issueMutationResp.CreateIssue.Id
	log.WithFields(log.Fields{
		"issueID": issueId,
	}).Info("Created new Issue")

	// Create new IssueVariant
	issueVariantInput := client.IssueVariantInput{
		SecondaryName:     cve.Id,
		Description:       cve.GetDescription("en"),
		ExternalUrl:       p.CveDetailsUrl + cve.Id,
		IssueRepositoryId: p.IssueRepositoryId,
		IssueId:           issueId,
		Severity: &client.SeverityInput{
			Vector: cve.SeverityVector(),
			Rating: "None",
		},
	}
	variantMutationResp, err := client.CreateIssueVariant(
		ctx,
		p.GraphqlClient,
		&issueVariantInput,
	)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Couldn't create new IssueVariant")
		return fmt.Errorf("couldn't create new IssueVariant")
	}

	log.WithFields(log.Fields{
		"issueVariantId": variantMutationResp.CreateIssueVariant.Id,
	}).Info("Created new IssueVariant")

	return nil
}

// ProcessOrUpdate looks up an existing Issue by CVE ID. If none exists it creates
// both the Issue and IssueVariant (same as Process). If one exists it compares the
// description and CVSS vector and only writes an update when something changed.
func (p *Processor) ProcessOrUpdate(ctx context.Context, cve *models.Cve) error {
	resp, err := client.GetIssues(
		ctx,
		p.GraphqlClient,
		&client.IssueFilter{
			PrimaryName: []string{cve.Id},
		},
	)
	if err != nil {
		return fmt.Errorf("couldn't look up issue for %s: %w", cve.Id, err)
	}

	if resp.Issues == nil || resp.Issues.TotalCount == 0 || len(resp.Issues.Edges) == 0 {
		return p.Process(ctx, cve)
	}

	issueEdge := resp.Issues.Edges[0]
	if issueEdge == nil || issueEdge.Node == nil {
		return fmt.Errorf("unexpected nil node for existing issue %s", cve.Id)
	}
	existing := issueEdge.Node

	newDescription := cve.GetDescription("en")
	newVector := cve.SeverityVector()

	issueChanged := existing.Description != newDescription
	if issueChanged {
		_, err = client.UpdateIssue(
			ctx,
			p.GraphqlClient,
			existing.Id,
			&client.IssueInput{
				PrimaryName: existing.PrimaryName,
				Description: newDescription,
				Type:        "Vulnerability",
			},
		)
		if err != nil {
			return fmt.Errorf("couldn't update issue %s: %w", cve.Id, err)
		}
		log.WithFields(log.Fields{"cve": cve.Id}).Info("Updated Issue description")
	}

	if existing.IssueVariants == nil || len(existing.IssueVariants.Edges) == 0 {
		// Variant is missing — create it
		_, err = client.CreateIssueVariant(
			ctx,
			p.GraphqlClient,
			&client.IssueVariantInput{
				SecondaryName:     cve.Id,
				Description:       newDescription,
				ExternalUrl:       p.CveDetailsUrl + cve.Id,
				IssueRepositoryId: p.IssueRepositoryId,
				IssueId:           existing.Id,
				Severity: &client.SeverityInput{
					Vector: newVector,
					Rating: "None",
				},
			},
		)
		if err != nil {
			return fmt.Errorf("couldn't create missing issue variant for %s: %w", cve.Id, err)
		}
		log.WithFields(log.Fields{"cve": cve.Id}).Info("Created missing IssueVariant")
		return nil
	}

	variantEdge := existing.IssueVariants.Edges[0]
	if variantEdge == nil || variantEdge.Node == nil {
		return fmt.Errorf("unexpected nil variant node for issue %s", cve.Id)
	}
	variant := variantEdge.Node

	existingVector := ""
	if variant.Severity != nil && variant.Severity.Cvss != nil {
		existingVector = variant.Severity.Cvss.Vector
	}

	variantChanged := existingVector != newVector || variant.SecondaryName != cve.Id || issueChanged
	if variantChanged {
		_, err = client.UpdateIssueVariant(
			ctx,
			p.GraphqlClient,
			variant.Id,
			&client.IssueVariantInput{
				SecondaryName:     cve.Id,
				Description:       newDescription,
				ExternalUrl:       p.CveDetailsUrl + cve.Id,
				IssueRepositoryId: p.IssueRepositoryId,
				IssueId:           existing.Id,
				Severity: &client.SeverityInput{
					Vector: newVector,
					Rating: "None",
				},
			},
		)
		if err != nil {
			return fmt.Errorf("couldn't update issue variant for %s: %w", cve.Id, err)
		}
		log.WithFields(log.Fields{"cve": cve.Id}).Info("Updated IssueVariant severity")
	}

	if !issueChanged && !variantChanged {
		log.WithFields(log.Fields{"cve": cve.Id}).Debug("No changes detected, skipping update")
	}

	return nil
}
