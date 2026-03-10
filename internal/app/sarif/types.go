// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package sarif

import (
		"github.com/cloudoperators/heureka/internal/entity"
 		"fmt"
	)

type SARIFDocument struct {
	Version string      `json:"version"`
	Schema  string      `json:"$schema"`
	Runs    []SARIFRun  `json:"runs"`
}

type SARIFRun struct {
	Tool    SARIFTool       `json:"tool"`
	Results []SARIFResult   `json:"results"`
}

type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

type SARIFDriver struct {
	Name            string      `json:"name"`
	FullName        string      `json:"fullName"`
	InformationUri  string      `json:"informationUri"`
	Version         string      `json:"version"`
	Rules           []SARIFRule `json:"rules"`
	SemanticVersion string      `json:"semanticVersion"`
}

type SARIFRule struct {
	Id                    string                   `json:"id"`
	Name                  string                   `json:"name"`
	ShortDescription      SARIFDescription        `json:"shortDescription"`
	FullDescription       SARIFDescription        `json:"fullDescription"`
	DefaultConfiguration  SARIFConfiguration      `json:"defaultConfiguration"`
	HelpUri               string                   `json:"helpUri"`
	Help                  SARIFHelp               `json:"help"`
	Properties            map[string]interface{}  `json:"properties"`
	Tags                  []string                `json:"tags"`
}

type SARIFDescription struct {
	Text     string `json:"text"`
	Markdown string `json:"markdown"`
}

type SARIFConfiguration struct {
	Level string `json:"level"`
}

type SARIFHelp struct {
	Text     string `json:"text"`
	Markdown string `json:"markdown"`
}

type SARIFResult struct {
	RuleId    string              `json:"ruleId"`
	RuleIndex int                 `json:"ruleIndex"`
	Level     string              `json:"level"`
	Message   SARIFMessage        `json:"message"`
	Locations []SARIFLocation     `json:"locations"`
	Properties map[string]interface{} `json:"properties"`
}

type SARIFMessage struct {
	Text     string `json:"text"`
	Markdown string `json:"markdown"`
	Id       string `json:"id"`
}

type SARIFLocation struct {
	Id                 string                    `json:"id"`
	PhysicalLocation   SARIFPhysicalLocation    `json:"physicalLocation"`
	LogicalLocations   []SARIFLogicalLocation   `json:"logicalLocations"`
}

type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
	Region           SARIFRegion           `json:"region"`
}

type SARIFArtifactLocation struct {
	Uri       string `json:"uri"`
	UriBaseId string `json:"uriBaseId"`
	Index     int    `json:"index"`
}

type SARIFLogicalLocation struct {
	Name           string `json:"name"`
	FullyQualifiedName string `json:"fullyQualifiedName"`
	Kind           string `json:"kind"`
}

type SARIFRegion struct {
	StartLine   int `json:"startLine"`
	EndLine     int `json:"endLine"`
	StartColumn int `json:"startColumn"`
	EndColumn   int `json:"endColumn"`
}

type ParsedSARIFData struct {
	ScannerName string
	Rules       map[string]*SARIFRule
	Results     []ParsedSARIFResult
	Errors      []ParseError
}

type ParsedSARIFResult struct {
	Rule          *SARIFRule
	Result        *SARIFResult
	ArtifactUri   string
	Severity      entity.SeverityValues
	Message       string
}

type ParseError struct {
	Line    int
	Message string
	Severity string
}
type PackageInfo struct {
	Name    string
	Version string
	Purl    string
}

func (pi PackageInfo) String() string {
	if pi.Purl != "" {
		return pi.Purl
	}
	return fmt.Sprintf("%s (version %s)", pi.Name, pi.Version)
}

func (psr *ParsedSARIFResult) GetPackageInfo() (*PackageInfo, bool) {
	if psr == nil || psr.Result == nil || psr.Result.Properties == nil {
		return nil, false
	}

	props := psr.Result.Properties
	info := &PackageInfo{}

	if purl, ok := props["purl"].(string); ok && purl != "" {
		info.Purl = purl
	} else if purl, ok := props["Purl"].(string); ok && purl != "" {
		info.Purl = purl
	}

	nameKeys := []string{"PkgName", "pkgName", "packageName", "name"}
	for _, key := range nameKeys {
		if val, ok := props[key].(string); ok && val != "" {
			info.Name = val
			break
		}
	}

	versionKeys := []string{"InstalledVersion", "version", "installedVersion", "pkgVersion"}
	for _, key := range versionKeys {
		if val, ok := props[key].(string); ok && val != "" {
			info.Version = val
			break
		}
	}

	return info, info.Purl != "" || info.Name != ""
}