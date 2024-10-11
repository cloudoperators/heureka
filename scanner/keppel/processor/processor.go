// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"github.com/cloudoperators/heureka/scanners/keppel/client"
	"github.com/cloudoperators/heureka/scanners/keppel/models"
	log "github.com/sirupsen/logrus"
)

type Processor struct {
	Client *graphql.Client
}

func NewProcessor(cfg Config) *Processor {
	httpClient := http.Client{}
	gClient := graphql.NewClient(cfg.HeurekaUrl, &httpClient)
	return &Processor{
		Client: &gClient,
	}
}

func (p *Processor) ProcessRepository(registry string, account models.Account, repository models.Repository) (*client.Component, error) {
	r, err := client.CreateComponent(context.Background(), *p.Client, &client.ComponentInput{
		Name: fmt.Sprintf("%s/%s/%s", registry, account.Name, repository.Name),
		Type: client.ComponentTypeValuesContainerimage,
	})

	if err != nil {
		return nil, err
	}

	component := r.GetCreateComponent()

	log.WithFields(log.Fields{
		"componentId": component.Id,
		"component":   component,
	}).Info("Component created")

	return component, nil
}

func (p *Processor) ProcessManifest(manifest models.Manifest, componentId string) (*client.ComponentVersion, error) {
	r, err := client.CreateComponentVersion(context.Background(), *p.Client, &client.ComponentVersionInput{
		Version:     manifest.Digest,
		ComponentId: componentId,
	})

	if err != nil {
		log.WithError(err).Error("Error while creating component")
		return nil, err
	}

	componentVersion := r.GetCreateComponentVersion()

	log.WithFields(log.Fields{
		"componentVersionId": componentVersion.Id,
		"componentVersion":   componentVersion,
	}).Info("ComponentVersion created")

	return componentVersion, nil
}

func (p *Processor) ProcessReport(report models.TrivyReport, componentVersionId string) {
	for _, result := range report.Results {
		for _, vulnerability := range result.Vulnerabilities {
			issue, err := p.GetIssue(vulnerability.VulnerabilityID)
			if err != nil {
				log.WithFields(log.Fields{
					"vulnerabilityID":    vulnerability.VulnerabilityID,
					"issueID":            issue.Id,
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
			_, err = client.AddComponentVersionToIssue(context.Background(), *p.Client, issue.Id, componentVersionId)

			if err != nil {
				log.WithFields(log.Fields{
					"issueId":            issue.Id,
					"componentVersionId": componentVersionId,
				}).WithError(err).Error("Could not add component version to issue")
			} else {
				log.WithFields(log.Fields{
					"issueId":            issue.Id,
					"componentVersionId": componentVersionId,
				}).Info("Added issue to componentVersion")
			}
		}

	}
}

func (p *Processor) GetComponent(name string) (*client.Component, error) {
	r, err := client.ListComponents(context.Background(), *p.Client, &client.ComponentFilter{
		ComponentName: []string{name},
	}, 1)

	if err != nil {
		return nil, err
	}

	var component *client.Component
	if len(r.Components.Edges) > 0 {
		component = r.Components.Edges[0].GetNode()
	}

	return component, nil
}

func (p *Processor) GetComponentVersion(version string) (*client.ComponentVersion, error) {
	r, err := client.ListComponentVersions(context.Background(), *p.Client, &client.ComponentVersionFilter{
		Version: []string{version},
	}, 1)

	if err != nil {
		return nil, err
	}

	var componentVersion *client.ComponentVersion
	if len(r.ComponentVersions.Edges) > 0 {
		componentVersion = r.ComponentVersions.Edges[0].GetNode()
	}

	return componentVersion, nil
}

func (p *Processor) GetIssue(primaryName string) (*client.Issue, error) {
	r, err := client.ListIssues(context.Background(), *p.Client, &client.IssueFilter{
		PrimaryName: []string{primaryName},
	}, 1)

	if err != nil {
		return nil, err
	}

	var issue *client.Issue
	if len(r.Issues.Edges) > 0 {
		issue = r.Issues.Edges[0].GetNode()
	}

	return issue, nil
}

func (p *Processor) CreateIssue(primaryName string, description string) (*client.Issue, error) {
	r, err := client.CreateIssue(context.Background(), *p.Client, &client.IssueInput{
		PrimaryName: primaryName,
		Description: description,
		Type:        "Vulnerability",
	})

	if err != nil {
		return nil, err
	}

	issue := r.GetCreateIssue()

	log.WithFields(log.Fields{
		"issueId": issue.Id,
		"issue":   issue,
	}).Info("Issue created")

	return issue, nil
}
