// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package sarif

import (
	"context"
	"testing"
)

func TestSARIFImportWithUnresolvedPackages(t *testing.T) {
	exampleTrivySARIF := `{
  "version": "2.1.0",
  "$schema": "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Trivy",
          "version": "0.49.1",
          "rules": [
            {
              "id": "CVE-2024-58251",
              "name": "CVE-2024-58251",
              "shortDescription": { "text": "Example vulnerability" },
              "fullDescription": { "text": "Example" },
              "defaultConfiguration": { "level": "warning" },
              "properties": { "security-severity": "6.1" }
            },
            {
              "id": "CVE-2025-99999",
              "name": "CVE-2025-99999",
              "shortDescription": { "text": "Unknown package vulnerability" },
              "fullDescription": { "text": "This package is not in our system" },
              "defaultConfiguration": { "level": "error" },
              "properties": { "security-severity": "9.0" }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "CVE-2024-58251",
          "level": "warning",
          "message": { "text": "Found in known package" },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": { "uri": "pkg:deb/debian/example/lib@1.0.0" }
              }
            }
          ],
          "properties": { "PkgName": "example/lib", "InstalledVersion": "1.0.0", "VulnerabilityID": "CVE-2024-58251" }
        },
        {
          "ruleId": "CVE-2025-99999",
          "level": "error",
          "message": { "text": "Found in unknown package" },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": { "uri": "pkg:npm/unknown/package@5.0.0" }
              }
            }
          ],
          "properties": { "PkgName": "unknown/package", "InstalledVersion": "5.0.0", "VulnerabilityID": "CVE-2025-99999" }
        }
      ]
    }
  ]
}`

	importer := NewSARIFImporter()
	ctx := context.Background()

	// Only define one component - the second package won't be found
	serviceComponents := []ComponentMatch{
		{
			ComponentInstanceId: 10,
			PackageName:         "example/lib",
			Version:             "1.0.0",
			Purl:               "pkg:deb/debian/example/lib@1.0.0",
		},
	}

	input := &ImportInput{
		SARIFDocument:     exampleTrivySARIF,
		ScannerName:       "trivy",
		ServiceId:         1,
		Tag:               "test-scan-with-unresolved",
		ServiceComponents: serviceComponents,
	}

	result, err := importer.ImportSARIF(ctx, input)
	if err != nil {
		t.Fatalf("ImportSARIF failed: %v", err)
	}

	// We should have 1 successful issue match and 1 warning about unresolved package
	if result.IssueMatchesCreated != 1 {
		t.Errorf("Expected 1 issue match to be created, got %d", result.IssueMatchesCreated)
	}

	if result.IssuesCreated != 1 {
		t.Errorf("Expected 1 unique issue to be processed, got %d", result.IssuesCreated)
	}

	// Should have at least 1 warning about unresolved package
	if len(result.Errors) == 0 {
		t.Errorf("Expected errors from unresolved package, got %d errors", len(result.Errors))
	}

	// Check that the error mentions the unresolved package
	foundUnresolvedError := false
	for _, errMsg := range result.Errors {
		if errMsg.Message == "Could not resolve package unknown/package (version 5.0.0) to a component instance" {
			foundUnresolvedError = true
			break
		}
	}

	if !foundUnresolvedError {
		t.Errorf("Expected error about unresolved package 'unknown/package', got errors: %v", result.Errors)
	}

	t.Logf("SARIF Import with unresolved packages test successful. IssuesCreated: %d, IssueMatchesCreated: %d, Errors: %+v",
		result.IssuesCreated, result.IssueMatchesCreated, result.Errors)
}

func TestSARIFImportMissingServiceComponents(t *testing.T) {
	exampleTrivySARIF := `{
  "version": "2.1.0",
  "$schema": "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Trivy",
          "version": "0.49.1",
          "rules": []
        }
      },
      "results": []
    }
  ]
}`

	importer := NewSARIFImporter()
	ctx := context.Background()

	// Empty service components - this should fail
	input := &ImportInput{
		SARIFDocument:     exampleTrivySARIF,
		ScannerName:       "trivy",
		ServiceId:         999, // Use a ServiceId that doesn't return anything in the mock
		Tag:               "test-scan",
		ServiceComponents: []ComponentMatch{}, // Empty!
	}

	result, err := importer.ImportSARIF(ctx, input)

	// Should fail because we don't provide service components
	if err == nil {
		t.Fatalf("Expected ImportSARIF to fail with missing service components, but got success")
	}

	if result != nil {
		t.Logf("Got error as expected: %v", err)
	}
}
