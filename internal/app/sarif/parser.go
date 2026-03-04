// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package sarif

import (
	"encoding/json"
	"fmt"

	appErrors "github.com/cloudoperators/heureka/internal/errors" // Corrected import path
)

type Parser struct{}

func (p *Parser) ParseSARIFDocument(sarifJSON string) (*ParsedSARIFData, error) {
	op := appErrors.Op("Parser.ParseSARIFDocument")

	var doc SARIFDocument
	if err := json.Unmarshal([]byte(sarifJSON), &doc); err != nil {
		return nil, appErrors.E(op, "Invalid SARIF JSON", err)
	}

	// Validate SARIF structure
	if err := p.validateSARIF(&doc); err != nil {
		return nil, appErrors.E(op, "SARIF validation failed", err)
	}

	parsed := &ParsedSARIFData{
		Rules:   make(map[string]*SARIFRule),
		Results: []ParsedSARIFResult{},
		Errors:  []ParseError{},
	}

	// Process all runs
	for _, run := range doc.Runs {
		parsed.ScannerName = run.Tool.Driver.Name

		// Index all rules by ID
		for i := range run.Tool.Driver.Rules {
			rule := &run.Tool.Driver.Rules[i]
			parsed.Rules[rule.Id] = rule
		}

		// Process all results
		for resultIdx, result := range run.Results {
			parsedResult, err := p.parseResult(result, parsed.Rules)
			if err != nil {
				parsed.Errors = append(parsed.Errors, ParseError{
					Line:     resultIdx,
					Message:  err.Error(),
					Severity: "error",
				})
				continue
			}
			parsed.Results = append(parsed.Results, parsedResult)
		}
	}

	return parsed, nil
}

// validateSARIF checks SARIF document structure
func (p *Parser) validateSARIF(doc *SARIFDocument) error {
	op := appErrors.Op("Parser.validateSARIF")

	if doc.Version != "2.1.0" {
		return appErrors.E(op, fmt.Sprintf("Unsupported SARIF version: %s (expected 2.1.0)", doc.Version))
	}

	if len(doc.Runs) == 0 {
		return appErrors.E(op, "SARIF document must contain at least one run")
	}

	for _, run := range doc.Runs {
		if run.Tool.Driver.Name == "" {
			return appErrors.E(op, "Tool driver name is required")
		}
	}

	return nil
}

// parseResult extracts and maps a single SARIF result
func (p *Parser) parseResult(result SARIFResult, rules map[string]*SARIFRule) (ParsedSARIFResult, error) {
	op := appErrors.Op("Parser.parseResult")

	if result.RuleId == "" {
		return ParsedSARIFResult{}, appErrors.E(op, "Result rule ID is required")
	}

	rule, exists := rules[result.RuleId]
	if !exists {
		return ParsedSARIFResult{}, appErrors.E(op, fmt.Sprintf("Rule not found: %s", result.RuleId))
	}

	// Extract artifact URI
	artifactUri := ""
	if len(result.Locations) > 0 && result.Locations[0].PhysicalLocation.ArtifactLocation.Uri != "" {
		artifactUri = result.Locations[0].PhysicalLocation.ArtifactLocation.Uri
	}

	if artifactUri == "" {
		return ParsedSARIFResult{}, appErrors.E(op, "No artifact location found in result")
	}

	// Map severity
	severity := MapSeverity(result.Level, rule, result.Properties)

	return ParsedSARIFResult{
		Rule:        rule,
		Result:      &result,
		ArtifactUri: artifactUri,
		Severity:    severity,
		Message:     result.Message.Text,
	}, nil
}

// GetRuleById retrieves a rule by ID from parsed data
func (p *ParsedSARIFData) GetRuleById(ruleId string) *SARIFRule {
	return p.Rules[ruleId]
}

// GetResultsByArtifact returns all results for a specific artifact
func (p *ParsedSARIFData) GetResultsByArtifact(artifactUri string) []ParsedSARIFResult {
	var results []ParsedSARIFResult
	for _, result := range p.Results {
		if result.ArtifactUri == artifactUri {
			results = append(results, result)
		}
	}
	return results
}

// HasErrors returns true if there were parsing errors
func (p *ParsedSARIFData) HasErrors() bool {
	return len(p.Errors) > 0
}

// ErrorCount returns the number of errors
func (p *ParsedSARIFData) ErrorCount() int {
	return len(p.Errors)
}
