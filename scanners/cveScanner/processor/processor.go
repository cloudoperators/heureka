// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"context"

	"github.com/machinebox/graphql"
	"github.wdf.sap.corp/cc/heureka/scanners/cveScanner/models"
)

type Processor struct {
	Client *graphql.Client
}

func NewProcessor(heurekaUrl string) *Processor {
	return &Processor{
		Client: graphql.NewClient(heurekaUrl),
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

	err := p.Client.Run(context.Background(), req, nil)

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
