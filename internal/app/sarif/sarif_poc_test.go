// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package sarif

import (
	"context"
	"testing"
)

func TestSARIFImportPOC(t *testing.T) {
	// Example Trivy SARIF output (truncated for brevity, real one is in docs/SARIF_IMPLEMENTATION_GUIDE.md)
	// This is a minimal valid SARIF document for testing purposes
	exampleTrivySARIF := `{
  "version": "2.1.0",
  "$schema": "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Trivy",
          "fullName": "Trivy Vulnerability Scanner",
          "informationUri": "https://aquasecurity.github.io/trivy",
          "version": "0.49.1",
          "rules": [
            {
              "id": "CVE-2024-58251",
              "name": "CVE-2024-58251",
              "shortDescription": { "text": "Improper Neutralization of Input During Web Page Generation ('Cross-site Scripting')" },
              "fullDescription": { "text": "Cross-site Scripting (XSS) in some web applications." },
              "defaultConfiguration": { "level": "warning" },
              "properties": { "security-severity": "6.1" }
            },
            {
              "id": "CVE-2025-46394",
              "name": "CVE-2025-46394",
              "shortDescription": { "text": "Directory Traversal in ExampleLib" },
              "fullDescription": { "text": "Path traversal vulnerability in ExampleLib affects versions < 1.2.3." },
              "defaultConfiguration": { "level": "note" },
              "properties": { "security-severity": "3.7" }
            },
            {
              "id": "CVE-2025-15467",
              "name": "CVE-2025-15467",
              "shortDescription": { "text": "Deserialization of Untrusted Data" },
              "fullDescription": { "text": "Remote code execution due to deserialization of untrusted data." },
              "defaultConfiguration": { "level": "error" },
              "properties": { "security-severity": "9.8" }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "CVE-2024-58251",
          "ruleIndex": 0,
          "level": "warning",
          "message": { "text": "Vulnerability CVE-2024-58251 found in package example/lib@1.0.0" },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": { "uri": "pkg:deb/debian/example/lib@1.0.0?arch=amd64" },
                "region": { "startLine": 1 }
              }
            }
          ],
          "properties": { "PkgName": "example/lib", "InstalledVersion": "1.0.0", "VulnerabilityID": "CVE-2024-58251" }
        },
        {
          "ruleId": "CVE-2025-46394",
          "ruleIndex": 1,
          "level": "note",
          "message": { "text": "Vulnerability CVE-2025-46394 found in package another/dep@2.1.0" },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": { "uri": "pkg:npm/another/dep@2.1.0" },
                "region": { "startLine": 1 }
              }
            }
          ],
          "properties": { "PkgName": "another/dep", "InstalledVersion": "2.1.0", "VulnerabilityID": "CVE-2025-46394" }
        },
        {
          "ruleId": "CVE-2025-15467",
          "ruleIndex": 2,
          "level": "error",
          "message": { "text": "Vulnerability CVE-2025-15467 found in package critical/app@0.5.0" },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": { "uri": "pkg:golang/critical/app@0.5.0" },
                "region": { "startLine": 1 }
              }
            }
          ],
          "properties": { "PkgName": "critical/app", "InstalledVersion": "0.5.0", "VulnerabilityID": "CVE-2025-15467" }
        }
      ]
    }
  ]
}`

	importer := NewSARIFImporter()
	ctx := context.Background()

	input := &ImportInput{
		SARIFDocument:    exampleTrivySARIF,
		ScannerName:      "trivy",
		ServiceId:        1,
		Tag:              "test-scan-1",
	}

	result, err := importer.ImportSARIF(ctx, input)
	if err != nil {
		t.Fatalf("ImportSARIF failed: %v", err)
	}

	if len(result.Errors) > 0 {
		t.Errorf("ImportSARIF returned errors: %v", result.Errors)
	}

	if result.IssueMatchesCreated != 3 {
		t.Errorf("Expected 3 issue matches to be created, got %d", result.IssueMatchesCreated)
	}

	// Because of how mockIssueHandler works, each CreateIssue call will return a new mocked issue.
	// If Issue creation was properly deduplicated, this would be 3. For POC, it's fine.
	if result.IssuesCreated != 3 {
		t.Errorf("Expected 3 unique issues to be processed, got %d", result.IssuesCreated)
	}

	t.Logf("SARIF Import POC successful. ScannerRunId: %d, IssuesCreated: %d, IssueMatchesCreated: %d",
		result.ScannerRunId, result.IssuesCreated, result.IssueMatchesCreated)
}
