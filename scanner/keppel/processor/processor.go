// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"context"

	"github.com/cloudoperators/heureka/scanners/keppel/models"
	"github.com/machinebox/graphql"
)

type Processor struct {
	Client *graphql.Client
}

func NewProcessor(cfg Config) *Processor {
	return &Processor{
		Client: graphql.NewClient(cfg.HeurekaUrl),
	}
}

func (p *Processor) Setup() error {
	return nil
}

func (p *Processor) ProcessRepository(name string) (models.Component, error) {
	var componentRespData struct {
		Component models.Component `json:"createComponent"`
	}

	// Create new Component
	req := graphql.NewRequest(CreateComponentQuery)
	req.Var("input", map[string]string{
		"name": name,
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
		"type":        "containerImage",
		"componentId": componentId,
	})

	err := p.Client.Run(context.Background(), req, &componentVersionRespData)

	return componentVersionRespData.ComponentVersion, err
}

func (p *Processor) ProcessReport(report models.TrivyReport) {
}
