// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cloudoperators/heureka/scanners/keppel/client"
	"github.com/cloudoperators/heureka/scanners/keppel/models"
	log "github.com/sirupsen/logrus"
)

// newVuln holds the data needed to create an Issue + IssueVariant for a CVE
// that does not yet exist in Heureka.
type newVuln struct {	primaryName string
	description string
	cvssVector  string
	severity    string
}

type linkPair struct {
	primaryName        string
	issueId            string
	componentVersionId string
}

// batchGQLResponse is the raw GraphQL response envelope for multi-alias
// mutations where partial failures are expected.
type batchGQLResponse struct {
	Data   map[string]json.RawMessage `json:"data"`
	Errors []batchGQLError            `json:"errors"`
}

type batchGQLError struct {
	Message string   `json:"message"`
	Path    []string `json:"path"`
}

func errByAlias(errs []batchGQLError) map[string]string {
	m := make(map[string]string, len(errs))
	for _, e := range errs {
		if len(e.Path) > 0 {
			m[e.Path[0]] = e.Message
		}
	}
	return m
}

// doRawGQL sends a raw GraphQL mutation string to the Heureka endpoint and
// returns the decoded envelope. A non-empty errors[] is not treated as fatal —
// the caller inspects partial results itself.
func (p *Processor) doRawGQL(ctx context.Context, query string) (batchGQLResponse, error) {
	body, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return batchGQLResponse{}, fmt.Errorf("marshal batch request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.heurekaURL, bytes.NewReader(body))
	if err != nil {
		return batchGQLResponse{}, fmt.Errorf("create batch request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return batchGQLResponse{}, fmt.Errorf("execute batch request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return batchGQLResponse{}, fmt.Errorf("read batch response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return batchGQLResponse{}, fmt.Errorf("batch request HTTP %d: %s", resp.StatusCode, respBody)
	}

	var gqlResp batchGQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return batchGQLResponse{}, fmt.Errorf("decode batch response: %w", err)
	}
	return gqlResp, nil
}

// gqlString JSON-encodes s, producing a valid quoted GraphQL string literal.
func gqlString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func chunks[T any](s []T, n int) [][]T {
	if n <= 0 || len(s) == 0 {
		return nil
	}
	result := make([][]T, 0, (len(s)+n-1)/n)
	for len(s) > 0 {
		k := n
		if k > len(s) {
			k = len(s)
		}
		result = append(result, s[:k])
		s = s[k:]
	}
	return result
}

func (p *Processor) batchCreateIssues(ctx context.Context, vulns []newVuln) map[string]*client.Issue {
	results := make(map[string]*client.Issue, len(vulns))

	for _, chunk := range chunks(vulns, p.BatchSize) {
		var sb strings.Builder
		sb.WriteString("mutation {")
		for i, v := range chunk {
			fmt.Fprintf(&sb,
				" i%d: createIssue(input: {primaryName: %s, description: %s, type: Vulnerability, uuid: %s}) { id primaryName description type }",
				i,
				gqlString(v.primaryName),
				gqlString(v.description),
				gqlString(p.uuid),
			)
		}
		sb.WriteString(" }")

		gqlResp, err := p.doRawGQL(ctx, sb.String())
		if err != nil {
			for _, v := range chunk {
				log.WithFields(log.Fields{
					"primaryName": v.primaryName,
				}).WithError(err).Error("batch createIssue request failed")
			}
			continue
		}

		aliasErrors := errByAlias(gqlResp.Errors)
		for i, v := range chunk {
			alias := fmt.Sprintf("i%d", i)
			raw, ok := gqlResp.Data[alias]
			if !ok || string(raw) == "null" {
				log.WithFields(log.Fields{
					"primaryName": v.primaryName,
					"error":       aliasErrors[alias],
				}).Error("createIssue failed in batch")
				continue
			}
			var issue client.Issue
			if err := json.Unmarshal(raw, &issue); err != nil {
				log.WithFields(log.Fields{
					"primaryName": v.primaryName,
				}).WithError(err).Error("decode createIssue batch result")
				continue
			}
			log.WithFields(log.Fields{
				"issueId":     issue.Id,
				"primaryName": issue.PrimaryName,
			}).Info("Issue created (batch)")
			results[v.primaryName] = &issue
		}
	}

	return results
}

// batchCreateIssueVariants returns the set of primaryNames for which variant creation failed.
func (p *Processor) batchCreateIssueVariants(
	ctx context.Context,
	vulns []newVuln,
	issueIDByName map[string]string,
) map[string]struct{} {
	failed := make(map[string]struct{})

	for _, chunk := range chunks(vulns, p.BatchSize) {
		var sb strings.Builder
		sb.WriteString("mutation {")
		for i, v := range chunk {
			severityValue := string(models.GetSeverityValue(v.severity))
			fmt.Fprintf(&sb,
				" i%d: createIssueVariant(input: {secondaryName: %s, description: %s, externalUrl: %s, issueId: %s, issueRepositoryId: %s, severity: {vector: %s, rating: %s}}) { id secondaryName issueId }",
				i,
				gqlString(v.primaryName),
				gqlString(v.description),
				gqlString(p.CveDetailsUrl+v.primaryName),
				gqlString(issueIDByName[v.primaryName]),
				gqlString(p.IssueRepositoryId),
				gqlString(v.cvssVector),
				severityValue, // enum value — no quotes in GraphQL
			)
		}
		sb.WriteString(" }")

		gqlResp, err := p.doRawGQL(ctx, sb.String())
		if err != nil {
			for _, v := range chunk {
				log.WithFields(log.Fields{
					"primaryName": v.primaryName,
				}).WithError(err).Error("batch createIssueVariant request failed")
				failed[v.primaryName] = struct{}{}
			}
			continue
		}

		aliasErrors := errByAlias(gqlResp.Errors)
		for i, v := range chunk {
			alias := fmt.Sprintf("i%d", i)
			raw, ok := gqlResp.Data[alias]
			if !ok || string(raw) == "null" {
				log.WithFields(log.Fields{
					"primaryName": v.primaryName,
					"issueId":     issueIDByName[v.primaryName],
					"error":       aliasErrors[alias],
				}).Error("createIssueVariant failed in batch")
				failed[v.primaryName] = struct{}{}
				continue
			}
			var variant client.IssueVariant
			if err := json.Unmarshal(raw, &variant); err != nil {
				log.WithFields(log.Fields{
					"primaryName": v.primaryName,
				}).WithError(err).Error("decode createIssueVariant batch result")
				failed[v.primaryName] = struct{}{}
				continue
			}
			log.WithFields(log.Fields{
				"issueVariantId": variant.Id,
			}).Info("IssueVariant created (batch)")
		}
	}

	return failed
}

func (p *Processor) batchAddComponentVersionToIssue(ctx context.Context, pairs []linkPair) {
	for _, chunk := range chunks(pairs, p.BatchSize) {
		var sb strings.Builder
		sb.WriteString("mutation {")
		for i, pair := range chunk {
			fmt.Fprintf(&sb,
				" i%d: addComponentVersionToIssue(issueId: %s, componentVersionId: %s) { id }",
				i,
				gqlString(pair.issueId),
				gqlString(pair.componentVersionId),
			)
		}
		sb.WriteString(" }")

		gqlResp, err := p.doRawGQL(ctx, sb.String())
		if err != nil {
			for _, pair := range chunk {
				log.WithFields(log.Fields{
					"primaryName":        pair.primaryName,
					"issueId":            pair.issueId,
					"componentVersionId": pair.componentVersionId,
				}).WithError(err).Error("batch addComponentVersionToIssue request failed")
			}
			continue
		}

		aliasErrors := errByAlias(gqlResp.Errors)
		for i, pair := range chunk {
			alias := fmt.Sprintf("i%d", i)
			raw, ok := gqlResp.Data[alias]
			if !ok || string(raw) == "null" {
				log.WithFields(log.Fields{
					"primaryName":        pair.primaryName,
					"issueId":            pair.issueId,
					"componentVersionId": pair.componentVersionId,
					"error":              aliasErrors[alias],
				}).Error("addComponentVersionToIssue failed in batch")
				continue
			}
			log.WithFields(log.Fields{
				"issueId":            pair.issueId,
				"componentVersionId": pair.componentVersionId,
			}).Info("Added issue to componentVersion (batch)")
		}
	}
}
