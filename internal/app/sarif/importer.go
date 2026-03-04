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
	return nil, nil // Not implemented for POC
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

	result := &ImportResult{
		Errors: []ImportError{},
	}

	parsed, err := si.parser.ParseSARIFDocument(input.SARIFDocument)
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

	uniqueArtifacts := make(map[string]bool)

	for _, parsedResult := range parsed.Results {
		artifactUri := parsedResult.ArtifactUri

		// Resolve asset
		// For POC, we mock the asset resolution
		asset := &entity.ComponentInstance{Id: 123}

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
