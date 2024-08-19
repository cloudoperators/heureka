// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"context"
	"fmt"

	"github.com/cloudoperators/heureka/scanners/keppel/models"
	"github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"
)

type Processor struct {
	Client *graphql.Client
}

func NewProcessor(cfg Config) *Processor {
	return &Processor{
		Client: graphql.NewClient(cfg.HeurekaUrl),
	}
}

func (p *Processor) ProcessRepository(repository models.Repository) (models.Component, error) {
	var componentRespData struct {
		Component models.Component `json:"createComponent"`
	}

	// Create new Component
	req := graphql.NewRequest(CreateComponentQuery)
	req.Var("input", map[string]string{
		"name": repository.Name,
		"type": "containerImage",
	})

	err := p.Client.Run(context.Background(), req, &componentRespData)

	return componentRespData.Component, err
}

func (p *Processor) ProcessManifest(manifest models.Manifest, componentId string) (models.ComponentVersion, error) {
	var componentVersionRespData struct {
		ComponentVersion models.ComponentVersion `json:"createComponentVersion"`
	}

	// Create new ComponentVersion
	req := graphql.NewRequest(CreateComponentVersionQuery)
	req.Var("input", map[string]string{
		"version":     manifest.Digest,
		"componentId": componentId,
	})

	err := p.Client.Run(context.Background(), req, &componentVersionRespData)

	return componentVersionRespData.ComponentVersion, err
}

func (p *Processor) ProcessReport(report models.TrivyReport, componentVersionId string) {
	for _, result := range report.Results {
		for _, vulnerability := range result.Vulnerabilities {
			issue, err := p.GetIssue(vulnerability.VulnerabilityID)
			if err != nil {
				log.WithFields(log.Fields{
					"vulnerabilityID":    vulnerability.VulnerabilityID,
					"issueID":            issue.ID,
					"issuePrimaryName":   issue.PrimaryName,
					"componentVersionID": componentVersionId,
					"report":             report.ArtifactName,
				}).WithError(err).Error("Error while getting issue")
				continue
			}
			if issue == nil {
				log.WithFields(log.Fields{
					"vulnerabilityID": vulnerability.VulnerabilityID,
				}).Warning("Issue not found")
				continue
				// use this for inserting issues, necessary to test without nvd scanner
				// i, err := p.CreateIssue(vulnerability.VulnerabilityID, vulnerability.Description)
				// if err != nil {
				// 	fmt.Println(err)
				// 	continue
				// }
				// issue = &i
			}
			var respData struct {
				ComponentVersion models.ComponentVersion `json:"addComponentVersionToIssue"`
			}
			req := graphql.NewRequest(AddComponentVersionToIssueQuery)
			req.Var("componentVersionId", componentVersionId)
			req.Var("issueId", issue.ID)
			err = p.Client.Run(context.Background(), req, &respData)
			if err != nil {
				fmt.Println(err)
				log.WithFields(log.Fields{
					"issueID":            issue.ID,
					"componentVersionID": componentVersionId,
				}).WithError(err).Error("Could not add component version to issue")
			}
		}

	}
}

func (p *Processor) GetComponent(name string) (*models.Component, error) {
	var respData struct {
		Component models.ComponentConnection `json:"Components"`
	}

	req := graphql.NewRequest(ListComponentsQuery)
	req.Var("filter", map[string][]string{
		"componentName": {name},
	})
	req.Var("first", 1)

	err := p.Client.Run(context.Background(), req, &respData)

	if err != nil {
		return nil, err
	}

	var component *models.Component
	if len(respData.Component.Edges) > 0 {
		component = respData.Component.Edges[0].Node
	}

	return component, nil
}

func (p *Processor) GetComponentVersion(version string) (*models.ComponentVersion, error) {
	var respData struct {
		ComponentVersion models.ComponentVersionConnection `json:"ComponentVersions"`
	}

	req := graphql.NewRequest(ListComponentVersionsQuery)
	req.Var("filter", map[string][]string{
		"version": {version},
	})
	req.Var("first", 1)

	err := p.Client.Run(context.Background(), req, &respData)

	if err != nil {
		return nil, err
	}

	var componentVersion *models.ComponentVersion
	if len(respData.ComponentVersion.Edges) > 0 {
		componentVersion = respData.ComponentVersion.Edges[0].Node
	}

	return componentVersion, nil
}

func (p *Processor) GetIssue(primaryName string) (*models.Issue, error) {
	var respData struct {
		Issue models.IssueConnection `json:"Issues"`
	}

	req := graphql.NewRequest(ListIssueQuery)
	req.Var("filter", map[string][]string{
		"primaryName": {primaryName},
	})
	req.Var("first", 1)

	err := p.Client.Run(context.Background(), req, &respData)

	if err != nil {
		return nil, err
	}

	var issue *models.Issue
	if len(respData.Issue.Edges) > 0 {
		issue = respData.Issue.Edges[0].Node
	}

	return issue, nil
}

func (p *Processor) CreateIssue(primaryName string, description string) (models.Issue, error) {
	var issueRespData struct {
		Issue models.Issue `json:"createIssue"`
	}

	// Create new Issue
	req := graphql.NewRequest(CreateIssueQuery)
	req.Var("input", map[string]string{
		"primaryName": primaryName,
		"description": description,
		"type":        "Vulnerability",
	})

	err := p.Client.Run(context.Background(), req, &issueRespData)

	return issueRespData.Issue, err
}
