// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"context"
	"net/http"

	"fmt"

	"github.com/Khan/genqlient/graphql"
	log "github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/scanner/nvd/client"
	"github.wdf.sap.corp/cc/heureka/scanner/nvd/models"
	"time"
)

type Processor struct {
	GraphqlClient       graphql.Client
	IssueRepositoryName string
	IssueRepositoryId   string
	IssueRepositoryUrl  string
}

// NewProcessor
func NewProcessor(cfg Config) *Processor {
	httpClient := http.Client{Timeout: time.Duration(10) * time.Second}
	return &Processor{
		GraphqlClient:       graphql.NewClient(cfg.HeurekaUrl, &httpClient),
		IssueRepositoryName: cfg.IssueRepositoryName,
		IssueRepositoryUrl:  cfg.IssueRepositoryUrl,
	}
}

func (p *Processor) Setup() error {
	// Check if there is already an IssueRepository with the same name
	queryFilter := client.IssueRepositoryFilter{
		Name: []string{p.IssueRepositoryName},
	}
	issueRepositoryResp, err := client.GetIssueRepositories(context.TODO(), p.GraphqlClient, &queryFilter)

	if (err == nil) && (issueRepositoryResp.IssueRepositories.TotalCount == 0) {
		log.Warnf("There is no IssueRepository: %s", err)

		// Create new IssueRepository
		issueRepositoryInput := client.IssueRepositoryInput{
			Name: p.IssueRepositoryName,
			Url:  p.IssueRepositoryUrl,
		}
		resp, err := client.CreateIssueRepository(context.TODO(), p.GraphqlClient, &issueRepositoryInput)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Couldn't create new IssueRepository")
		}

		// Save IssueRepositoryId
		p.IssueRepositoryId = resp.CreateIssueRepository.Id
		log.WithFields(log.Fields{
			"issueRepositoryId": p.IssueRepositoryId,
		}).Info("Created new IssueRepository")
	} else {

		// Extract IssueRepositoryId
		for _, ir := range issueRepositoryResp.IssueRepositories.Edges {
			log.Debugf("nodeId: %s", ir.Node.Id)
			p.IssueRepositoryId = ir.Node.Id
			break
		}
		log.Debugf("IssueRepositoryId: %s", p.IssueRepositoryId)
	}
	return nil
}

func (p *Processor) Process(cve *models.Cve) error {
	var issueId string

	// Create new Issue
	createIssueInput := client.IssueInput{
		PrimaryName: cve.Id,
		Description: cve.GetDescription("en"),
		Type:        "Vulnerability",
	}
	createIssueResp, err := client.CreateIssue(context.TODO(), p.GraphqlClient, &createIssueInput)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Couldn't create new Issue")
		return fmt.Errorf("Couldn't create new Issue")
	}

	issueId = createIssueResp.CreateIssue.Id
	log.WithFields(log.Fields{
		"issueID": issueId,
	}).Info("Created new Issue")

	// Create new IssueVariant
	issueVariantInput := client.IssueVariantInput{
		SecondaryName:     cve.Id,
		Description:       cve.GetDescription("en"),
		IssueRepositoryId: p.IssueRepositoryId,
		IssueId:           issueId,
		Severity: &client.SeverityInput{
			Vector: cve.SeverityVector(),
		},
	}
	createIssueVariantResp, err := client.CreateIssueVariant(context.TODO(), p.GraphqlClient, &issueVariantInput)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Couldn't create new IssueVariant")
		return fmt.Errorf("couldn't create new IssueVariant")
	}

	log.WithFields(log.Fields{
		"issueVariantId": createIssueVariantResp.CreateIssueVariant.Id,
	}).Info("Created new IssueVariant")

	return nil
}
