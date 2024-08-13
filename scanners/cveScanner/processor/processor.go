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
	repositoryId, err := p.GetIssueRepositoryId()
	if err != nil {
		fmt.Println(err)
		return err
	}

	if repositoryId == "-1" {
		fmt.Printf("Invalid issue repository ID")

		// Create new issue repository
		issueRepositoryId, err := p.CreateIssueRepository()
		if err != nil {
			return fmt.Errorf("couldn't create new issue repository")
		}
		p.IssueRepositoryId = issueRepositoryId
	} else {
		p.IssueRepositoryId = repositoryId
	}

	return nil
}

func (p *Processor) GetIssueRepositoryId() (string, error) {
	query := p.GetIssueRepositoryIdQuery()
	req := graphql.NewRequest(query)

	req.Var("filter", map[string][]string{
		"name": {p.IssueRepositoryName},
	})

	var issueRepositoryRespData struct {
		IssueRepository models.IssueRepository `json:"IssueRepositories"`
	}

	err := p.Client.Run(context.Background(), req, &issueRepositoryRespData)
	if err != nil {
		return "", err
	}

	return issueRepositoryRespData.IssueRepository.Id, nil
}

func (p *Processor) CreateIssueRepository() (string, error) {
	var issueRepositoryRespData struct {
		IssueRepository models.IssueRepository `json:"createIssueRepository"`
	}
	query := p.CreateIssueRepositoryQuery()
	req := graphql.NewRequest(query)

	req.Var("input", map[string]string{
		"name": p.IssueRepositoryName,
		"url":  p.IssueRepositoryUrl,
	})

	err := p.Client.Run(context.Background(), req, &issueRepositoryRespData)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return issueRepositoryRespData.IssueRepository.Id, nil
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
