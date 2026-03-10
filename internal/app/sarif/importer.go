// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package sarif

import (
	"context"
	"fmt"
	"time"

	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/app/scanner"
	"github.com/cloudoperators/heureka/internal/entity"

	"github.com/google/uuid"
)

// Mock Database interface for POC
type mockDatabase struct{}

func (m *mockDatabase) CreateScannerRun(run *entity.ScannerRun) (*entity.ScannerRun, error) {
	run.RunID = 1
	return run, nil
}
func (m *mockDatabase) CompleteScannerRun(runUUID string) (*entity.ScannerRun, error) {
	return &entity.ScannerRun{UUID: runUUID}, nil
}
func (m *mockDatabase) FailScannerRun(runUUID, message string) (*entity.ScannerRun, error) {
	return &entity.ScannerRun{UUID: runUUID}, nil
}
func (m *mockDatabase) GetComponentInstance(id int64) (*entity.ComponentInstance, error) {
	return &entity.ComponentInstance{Id: id}, nil
}
func (m *mockDatabase) CreateScannerAssetMapping(mapping *entity.ScannerAssetMapping) (*entity.ScannerAssetMapping, error) {
	mapping.Id = 1
	return mapping, nil
}
func (m *mockDatabase) GetScannerAssetMappingByUri(scannerName, artifactUri string) (*entity.ScannerAssetMapping, error) {
	// purely mock
	if artifactUri == "/path/to/known/asset" {
		return &entity.ScannerAssetMapping{
			ComponentInstanceId: 42,
			ArtifactUri:         artifactUri,
		}, nil
	}
	return nil, nil
}


type mockIssueHandler struct{}

func (m *mockIssueHandler) CreateIssue(ctx context.Context, issue *entity.Issue) (*entity.Issue, error) {
	issue.Id = 1
	return issue, nil
}
func (m *mockIssueHandler) GetIssue(ctx context.Context, id int64) (*entity.Issue, error) { return nil, nil }
func (m *mockIssueHandler) ListIssues(ctx context.Context, options entity.IssueListOptions) ([]*entity.Issue, error) { return nil, nil }
func (m *mockIssueHandler) UpdateIssue(ctx context.Context, issue *entity.Issue) (*entity.Issue, error) { return nil, nil }
func (m *mockIssueHandler) DeleteIssue(ctx context.Context, id int64) error { return nil }

type mockIssueMatchHandler struct{}

func (m *mockIssueMatchHandler) CreateIssueMatch(ctx context.Context, match *entity.IssueMatch) (*entity.IssueMatch, error) {
	match.Id = 1
	return match, nil
}
func (m *mockIssueMatchHandler) GetIssueMatch(ctx context.Context, id int64) (*entity.IssueMatch, error) { return nil, nil }
func (m *mockIssueMatchHandler) ListIssueMatches(filter *entity.IssueMatchFilter, options *entity.ListOptions) (*entity.List[entity.IssueMatchResult], error) { return nil, nil }
func (m *mockIssueMatchHandler) UpdateIssueMatch(ctx context.Context, match *entity.IssueMatch) (*entity.IssueMatch, error) { return nil, nil }
func (m *mockIssueMatchHandler) DeleteIssueMatch(ctx context.Context, id int64) error { return nil }
func (m *mockIssueMatchHandler) ListIssueMatchesByIssue(ctx context.Context, issueId int64) ([]*entity.IssueMatch, error) { return nil, nil }


type sarifImporter struct {
	parser       *Parser
	assetMapper  scanner.AssetMapper
	issueHandler mockIssueHandler
	matchHandler mockIssueMatchHandler
	db           *mockDatabase
}

func NewSARIFImporter() Importer {
	db := &mockDatabase{}
	return &sarifImporter{
		parser:       &Parser{},
		assetMapper:  scanner.NewAssetMapper(db),
		issueHandler: mockIssueHandler{},
		matchHandler: mockIssueMatchHandler{},
		db:           db,
	}
}

func (m *mockDatabase) ListComponentInstances(serviceId int64) ([]ComponentMatch, error) {
	// purely mock
	if serviceId == 1 {
		return []ComponentMatch{
			{ComponentInstanceId: 101, PackageName: "example/lib", Version: "1.0.0", Purl: "pkg:npm/example/lib@1.0.0"},
			{ComponentInstanceId: 102, PackageName: "openssl", Version: "3.0.1", Purl: "pkg:generic/openssl@3.0.1"},
		}, nil
	}
	return []ComponentMatch{}, nil
}

func (si *sarifImporter) ImportSARIF(ctx context.Context, input *ImportInput) (*ImportResult, error) {
	op := appErrors.Op("SARIFImporter.ImportSARIF")

	if input.SARIFDocument == "" {
		return nil, appErrors.E(op, "SARIF document is required")
	}

	if input.ServiceId == 0 {
		return nil, appErrors.E(op, "Service ID is required")
	}

	if input.Tag == "" {
		return nil, appErrors.E(op, "Scanner run tag is required")
	}

	// AUTOMATION: If no components provided, fetch them from the database automatically
	serviceComponents := input.ServiceComponents
	if len(serviceComponents) == 0 {
		autoComponents, err := si.db.ListComponentInstances(input.ServiceId)
		if err != nil {
			return nil, appErrors.E(op, "Failed to auto-discover service components", err)
		}
		serviceComponents = autoComponents
	}

	if len(serviceComponents) == 0 {
		return nil, appErrors.E(op, "No component instances found for service. Resolution is impossible.")
	}

	result := &ImportResult{
		Errors: []ImportError{},
	}

	parsed, err := si.parser.ParseSARIFDocument(input.SARIFDocument)
	// ... (rest of the function remains the same, using serviceComponents)
	if err != nil {
		return result, appErrors.E(op, "Failed to parse SARIF", err)
	}

	for _, parseErr := range parsed.Errors {
		result.Errors = append(result.Errors, ImportError{
			Line:     parseErr.Line,
			Message:  parseErr.Message,
			Severity: parseErr.Severity,
		})
	}

	scannerRun := &entity.ScannerRun{
		UUID:      uuid.New().String(),
		Tag:       input.Tag,
		StartRun:  time.Now(),
		Completed: false,
	}

	createdRun, err := si.db.CreateScannerRun(scannerRun)
	if err != nil {
		return result, appErrors.E(op, "Failed to create scanner run", err)
	}

	result.ScannerRunId = createdRun.RunID

	// Create package resolver from service components (either provided or auto-discovered)
	resolver := NewPackageResolver(serviceComponents)

	uniqueArtifacts := make(map[string]bool)

	for _, parsedResult := range parsed.Results {
		artifactUri := parsedResult.ArtifactUri
		var componentInstanceId int64

		// Strategy 1: Look up by ScannerAssetMapping (Direct Mapping)
		mapping, err := si.assetMapper.GetAssetMapping(ctx, parsed.ScannerName, artifactUri)
		if err == nil && mapping != nil {
			componentInstanceId = mapping.ComponentInstanceId
		} else {
			// Strategy 2: Extract package info from SARIF and Resolve (Standardized Meta-matching)
			info, found := parsedResult.GetPackageInfo()
			if !found {
				result.Errors = append(result.Errors, ImportError{
					Line:     0,
					Message:  fmt.Sprintf("Failed to extract package info from artifact %s and no pre-mapping found", artifactUri),
					Severity: "warning",
				})
				continue
			}

			// Resolve via PackageResolver (PURL first, then Name/Version)
			id, resolved := resolver.Resolve(info)
			if !resolved {
				result.Errors = append(result.Errors, ImportError{
					Line:     0,
					Message:  fmt.Sprintf("Could not resolve package %s to a component instance", info.String()),
					Severity: "warning",
				})
				continue
			}
			componentInstanceId = id
		}

		asset := &entity.ComponentInstance{Id: componentInstanceId}
// ...

		issueEntity := &entity.Issue{
			Type:        entity.IssueTypeVulnerability,
			PrimaryName: parsedResult.Rule.Id,
			Description: parsedResult.Rule.ShortDescription.Text,
		}

		createdIssue, err := si.issueHandler.CreateIssue(ctx, issueEntity)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Line:     0,
				Message:  fmt.Sprintf("Failed to create issue %s: %v", issueEntity.PrimaryName, err),
				Severity: "error",
			})
			continue
		}

		matchEntity := &entity.IssueMatch{
			IssueId:           createdIssue.Id,
			ComponentInstanceId: asset.Id,
			Status:            entity.IssueMatchStatusValuesNew,
			UserId:            1,
		}

		severity := entity.NewSeverityFromRating(parsedResult.Severity)
		matchEntity.Severity = severity

		_, err = si.matchHandler.CreateIssueMatch(ctx, matchEntity)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Line:     0,
				Message:  fmt.Sprintf("Failed to create issue match for %s on %s: %v", issueEntity.PrimaryName, artifactUri, err),
				Severity: "warning",
			})
			continue
		}

		result.IssueMatchesCreated++

		if !uniqueArtifacts[artifactUri] {
			uniqueArtifacts[artifactUri] = true
			result.IssuesCreated++
		}
	}

	if len(result.Errors) == 0 {
		_, err = si.db.CompleteScannerRun(scannerRun.UUID)
		if err != nil {
			return result, appErrors.E(op, "Failed to complete scanner run", err)
		}
	} else {
		_, err = si.db.FailScannerRun(scannerRun.UUID, fmt.Sprintf("Import completed with %d errors", len(result.Errors)))
		if err != nil {
			return result, appErrors.E(op, "Failed to fail scanner run", err)
		}
	}

	return result, nil
}
