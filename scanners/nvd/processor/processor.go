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
	query := p.GetIssueRepositoryIdQuery()
	req := graphql.NewRequest(query)

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
	query := p.CreateIssueRepositoryQuery()
	req := graphql.NewRequest(query)

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

func (p *Processor) Process(cve *models.Cve) error {
	query := p.GetCreateIssueQuery()
	req := graphql.NewRequest(query)

	req.Var("input", map[string]string{
		"primaryName": cve.Id,
		"description": cve.GetDescription("en"),
		"type":        "Vulnerability",
	})

	var issueRespData struct {
		Issue models.Issue `json:"createIssue"`
	}

	err := p.Client.Run(context.Background(), req, &issueRespData)

	if err != nil {
		fmt.Println(err)
		return err
	}

	query = p.GetCreateIssueVariantQuery()
	req = graphql.NewRequest(query)

	req.Var("input", map[string]interface{}{
		"secondaryName":     cve.Id,
		"description":       cve.GetDescription("en"),
		"issueRepositoryId": p.IssueRepositoryId,
		"issueId":           issueRespData.Issue.Id,
		"severity": map[string]string{
			"vector": "todo",
		},
	})

	var issueVariantRespData struct {
		IssueVariant models.IssueVariant `json:"createIssueVariant"`
	}

	err = p.Client.Run(context.Background(), req, &issueVariantRespData)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return err
}

func (p *Processor) GetCreateIssueQuery() string {
	return `
	mutation ($input: IssueInput!) {
		createIssue (
			input: $input
		) {
			id
			primaryName
			description
			type
		}
	}
	`
}

func (p *Processor) GetCreateIssueVariantQuery() string {
	return `
	mutation ($input: IssueVariantInput!) {
    createIssueVariant (
        input: $input
    ) {
        id
        secondaryName
        issueId
 	   }
	}
	`
}

func (p *Processor) GetIssueRepositoryIdQuery() string {
	return `
	query ($filter: IssueRepositoryFilter) {
		IssueRepositories (
			filter: $filter,
		) {
            totalCount
			edges {
				node {
					id
				}
			}
		}
	}
	`
}

func (p *Processor) CreateIssueRepositoryQuery() string {
	return `
	mutation ($input: IssueRepositoryInput!) {
		createIssueRepository (
			input: $input
		) {
			id
			name
			url
		}
	}
	`
}
