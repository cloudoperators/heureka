// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/cloudoperators/heureka/scanners/keppel/client"
	"github.com/cloudoperators/heureka/scanners/keppel/models"
	"github.com/google/uuid"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
)

const (
	RottenVulnerabilityStatus = "Rotten"
)

type Processor struct {
	uuid                string
	tag                 string
	Client              *graphql.Client
	CveDetailsUrl       string
	IssueRepositoryUrl  string
	IssueRepositoryName string
	IssueRepositoryId   string
}

func NewProcessor(cfg Config, tag string) *Processor {
	httpClient := http.Client{}
	gClient := graphql.NewClient(cfg.HeurekaUrl, &httpClient)
	return &Processor{
		Client:              &gClient,
		uuid:                uuid.New().String(),
		tag:                 tag,
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
	listRepositoriesResp, err := client.GetIssueRepositories(context.TODO(), *p.Client, &queryFilter)
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
		issueMutationResp, err := client.CreateIssueRepository(context.TODO(), *p.Client, &issueRepositoryInput)
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

func (p *Processor) CreateScannerRun(ctx context.Context) error {
	_, err := client.CreateScannerRun(ctx, *p.Client, &client.ScannerRunInput{
		Uuid: p.uuid,
		Tag:  p.tag,
	})
	return err
}

func (p *Processor) CompleteScannerRun(ctx context.Context) error {
	_, err := client.CompleteScannerRun(ctx, *p.Client, p.uuid)
	return err
}

func (p *Processor) ProcessRepository(registry string, account models.Account, repository models.Repository) (*client.Component, error) {
	r, err := client.CreateComponent(context.Background(), *p.Client, &client.ComponentInput{
		Ccrn:         fmt.Sprintf("%s/%s/%s", registry, account.Name, repository.Name),
		Organization: account.Name,
		Repository:   repository.Name,
		Url:          fmt.Sprintf("https://%s/%s/%s", registry, account.Name, repository.Name),
		Type:         client.ComponentTypeValuesContainerimage,
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
		EndOfLife:   manifest.VulnerabilityStatus == RottenVulnerabilityStatus,
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
					"issueID":            lo.Ternary(issue != nil, issue.Id, ""),
					"issuePrimaryName":   lo.Ternary(issue != nil, issue.PrimaryName, ""),
					"componentVersionID": componentVersionId,
					"report":             report.ArtifactName,
				}).WithError(err).Error("Error while getting issue")
				continue
			}
			if issue == nil {
				// create only cve issues
				if !strings.HasPrefix(strings.ToLower(vulnerability.VulnerabilityID), "cve") {
					log.WithFields(log.Fields{
						"vulnerabilityID": vulnerability.VulnerabilityID,
					}).Warning("VulnerabilityID does not start with 'CVE'")
					continue
				}
				i, err := p.CreateIssue(vulnerability.VulnerabilityID, vulnerability.Description)
				if err != nil {
					log.Error(err)
					continue
				}
				issue = i
				cvssVector := ""
				// check if nvd CVSS vector is available
				if vulnerability.CVSS != nil {
					if _, ok := vulnerability.CVSS["nvd"]; ok {
						cvssVector = vulnerability.CVSS["nvd"].V3Vector
					}
				}
				_, err = p.CreateIssueVariant(vulnerability.VulnerabilityID, vulnerability.Description, issue.Id, cvssVector, vulnerability.Severity)
				if err != nil {
					log.Error(err)
					continue
				}
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

	if report.Metadata.OS.Eosl {
		_, err := client.UpdateComponentVersion(context.Background(), *p.Client, componentVersionId, &client.ComponentVersionInput{
			EndOfLife: true,
		})
		if err != nil {
			log.WithFields(log.Fields{
				"componentVersionId": componentVersionId,
			}).WithError(err).Error("Could not update ComponentVersion")
		}
	}
}

// GetAllComponents will fetch all available Components using pagination.
// pageSize specifies how many object we want to fetch in a row. Then we use cursor
// to fetch the next batch.
func (p *Processor) GetAllComponents(filter *client.ComponentFilter, pageSize int) ([]*client.ComponentAggregate, error) {
	var allComponents []*client.ComponentAggregate
	cursor := "" // Set initial cursor to ""

	for {
		// ListComponents also returns the ComponentVersions of each Component
		listComponentsResp, err := client.ListComponents(context.Background(), *p.Client, filter, pageSize, cursor)
		if err != nil {
			return nil, fmt.Errorf("cannot list Components: %w", err)
		}

		if len(listComponentsResp.Components.Edges) == 0 {
			break
		}

		for _, edge := range listComponentsResp.Components.Edges {
			allComponents = append(allComponents, edge.Node)
		}

		if len(listComponentsResp.Components.Edges) < pageSize {
			break
		}

		// Update cursor for the next iteration
		cursor = listComponentsResp.Components.Edges[len(listComponentsResp.Components.Edges)-1].Cursor
	}
	return allComponents, nil
}

func (p *Processor) GetComponentVersions(componentId string) ([]*client.ComponentVersion, error) {
	listCompVersionResp, err := client.ListComponentVersions(context.Background(), *p.Client, &client.ComponentVersionFilter{
		ComponentId: []string{componentId},
	})
	if err != nil {
		return nil, fmt.Errorf("cannot list ComponentVersions: %w", err)
	}

	var componentVersions []*client.ComponentVersion
	if len(listCompVersionResp.ComponentVersions.Edges) > 0 {
		for _, cv := range listCompVersionResp.GetComponentVersions().Edges {
			componentVersions = append(componentVersions, cv.GetNode())
		}
	}

	return componentVersions, nil
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
		Uuid:        p.uuid,
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

func (p *Processor) CreateIssueVariant(secondaryName string, description string, issueId string, vector string, severity string) (*client.IssueVariant, error) {
	severityValue := models.GetSeverityValue(severity)
	r, err := client.CreateIssueVariant(context.Background(), *p.Client, &client.IssueVariantInput{
		SecondaryName:     secondaryName,
		Description:       description,
		ExternalUrl:       p.CveDetailsUrl + secondaryName,
		IssueId:           issueId,
		IssueRepositoryId: p.IssueRepositoryId,
		Severity: &client.SeverityInput{
			Vector: vector,
			Rating: severityValue,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("secondaryName: %s, issueId: %s, severity: %s %w", secondaryName, issueId, severity, err)
	}

	issueVariant := r.GetCreateIssueVariant()

	log.WithFields(log.Fields{
		"issueVariantId": issueVariant.Id,
	}).Info("IssueVariant created")

	return issueVariant, nil
}
