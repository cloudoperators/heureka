// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"context"
	"fmt"

	"github.com/machinebox/graphql"
	"github.wdf.sap.corp/cc/heureka/scanners/cveScanner/models"
)

type Processor struct {
	Client              *graphql.Client
	IssueRepositoryName string
	IssueRepositoryId   string
	IssueRepositoryUrl  string
}

func NewProcessor(cfg Config) *Processor {
	return &Processor{
		Client:              graphql.NewClient(cfg.HeurekaUrl),
		IssueRepositoryName: cfg.IssueRepositoryName,
		IssueRepositoryUrl:  cfg.IssueRepositoryUrl,
	}
}

func (p *Processor) Setup() error {
	// Check if there is already an IssueRepository with the same name
	existentingIssueRepositoryId, err := p.GetIssueRepositoryId()
	if err != nil {
		fmt.Printf("err: %s", err)

		// Create new IssueRepository
		newIssueRepositoryId, err := p.CreateIssueRepository()
		if err != nil {
			fmt.Println(err)
		}
		p.IssueRepositoryId = newIssueRepositoryId
	} else {
		p.IssueRepositoryId = existentingIssueRepositoryId
	}

	return nil
}

// GetIssueRepositoryId fetches an IssueRepository based on the name of the scanner
func (p *Processor) GetIssueRepositoryId() (string, error) {
	var (
		newIssueRepositoryId          string
		issueRepositoryConnectionResp struct {
			IssueRepositoryConnection models.IssueRepositoryConnection `json:"IssueRepositories"`
		}
	)

	// Fetch IssueRepositoryId by name
	req := graphql.NewRequest(GetIssueRepositoryIdQuery)
	req.Var("filter", map[string][]string{
		"name": {p.IssueRepositoryName},
	})

	err := p.Client.Run(context.Background(), req, &issueRepositoryConnectionResp)
	if err != nil {
		return "", err
	}

	// TODO: What to do if multiple edges/Ids available?
	if issueRepositoryConnectionResp.IssueRepositoryConnection.TotalCount > 0 {
		for _, repositoryEdge := range issueRepositoryConnectionResp.IssueRepositoryConnection.Edges {
			fmt.Printf("id: %s", repositoryEdge.Node.Id)
			newIssueRepositoryId = repositoryEdge.Node.Id
		}
	} else {
		return "", fmt.Errorf("didn't get any repository ids")
	}

	return newIssueRepositoryId, nil
}

// CreateIssueRepository creates a new IssueRepository based on
// - the name and
// - the URL
// of the current scanner
func (p *Processor) CreateIssueRepository() (string, error) {
	var repositoryId string
	var createIssueRepositoryResp struct {
		IssueRepository models.IssueRepository `json:"createIssueRepository"`
	}

	req := graphql.NewRequest(CreateIssueRepositoryQuery)
	req.Var("input", map[string]string{
		"name": p.IssueRepositoryName,
		"url":  p.IssueRepositoryUrl,
	})

	err := p.Client.Run(context.Background(), req, &createIssueRepositoryResp)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	repositoryId = createIssueRepositoryResp.IssueRepository.Id

	if len(repositoryId) > 0 {
		return repositoryId, nil
	} else {
		return "", fmt.Errorf("repositoryId is empty")
	}
}

// GetIssueId ...
func (p *Processor) GetIssueId(cve *models.Cve) (string, error) {
	var issueConnectionResp struct {
		IssueConnection models.IssueConnection `json:"Issues"`
	}

	// Fetch Issue by CVE name
	req := graphql.NewRequest(GetIssueIdQuery)
	req.Var("filter", map[string][]string{
		"primaryName": {p.IssueRepositoryName},
	})

	err := p.Client.Run(context.Background(), req, &issueRepositoryConnectionResp)
	if err != nil {
		return "", err
	}

}

// CreateIssue creates a new Issue based on a CVE
func (p *Processor) CreateIssue(cve *models.Cve) (string, error) {
	var issueRespData struct {
		Issue models.Issue `json:"createIssue"`
	}

	// Create new Issue
	req := graphql.NewRequest(CreateIssueQuery)
	req.Var("input", map[string]string{
		"primaryName": cve.Id,
		"description": cve.GetDescription("en"),
		"type":        "Vulnerability",
	})

	err := p.Client.Run(context.Background(), req, &issueRespData)

	return issueRespData.Issue.Id, err
}

// CreateIssueVariant ...
func (p *Processor) CreateIssueVariant(issueId string, issueRepositoryId string, cve *models.Cve) (string, error) {
	var issueVariantRespData struct {
		IssueVariant models.IssueVariant `json:"createIssueVariant"`
	}

	// Create new IssueVariant
	req := graphql.NewRequest(CreateIssueVariantQuery)
	req.Var("input", map[string]interface{}{
		"secondaryName":     cve.Id,
		"description":       cve.GetDescription("en"),
		"issueRepositoryId": issueRepositoryId,
		"issueId":           issueId,
		"severity": map[string]string{
			"vector": cve.GetSeverityVector(),
		},
	})

	err := p.Client.Run(context.Background(), req, &issueVariantRespData)

	return issueVariantRespData.IssueVariant.Id, err
}

func (p *Processor) Process(cve *models.Cve) error {
	// Create new Issue
	issueId, err := p.CreateIssue(cve)
	if err != nil {
		fmt.Printf("couldn't create new Issue")
	}

	// Create new IssueVariant
	issueVariantId, err := p.CreateIssueVariant(issueId, p.IssueRepositoryId, cve)
	if err != nil {
		return fmt.Errorf("couldn't create new IssueVariant")
	}

	fmt.Printf("Created new Issue: %s", issueId)
	fmt.Printf("Created new IssueVariant: %s", issueVariantId)
	return nil
}
