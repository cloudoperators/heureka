// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package models

// Schemas
// CVE 2.0
// https://csrc.nist.gov/schema/nvd/api/2.0/cve_api_json_2.0.schema
// CVS 3.1
// https://csrc.nist.gov/schema/nvd/api/2.0/external/cvss-v3.1.json
// CVS 3.0
// https://csrc.nist.gov/schema/nvd/api/2.0/external/cvss-v3.0.json
// CVS 2.0
// https://csrc.nist.gov/schema/nvd/api/2.0/external/cvss-v2.0.json

type CveResponse struct {
	ResultsPerPage  int       `json:"resultsPerPage"`
	StartIndex      int       `json:"startIndex"`
	TotalResults    int       `json:"totalResults"`
	Format          string    `json:"format"`
	Version         string    `json:"version"`
	Timestamp       string    `json:"timestamp"`
	Vulnerabilities []CveItem `json:"vulnerabilities"`
}

type CveItem struct {
	Cve Cve `json:"cve"`
}

type Cve struct {
	Id                string          `json:"id,omitempty"`
	SourceIdentifier  string          `json:"sourceIdentifier,omitempty"`
	Published         string          `json:"published,omitempty"`
	LastModified      string          `json:"lastModified,omitempty"`
	VulnStatus        string          `json:"vulnStatus,omitempty"`
	EvaluatorComment  string          `json:"evaluatorComment,omitempty"`
	EvaluatorSolution string          `json:"evaluatorSolution,omitempty"`
	EvaluatorImpact   string          `json:"evaluatorImpact,omitempty"`
	CISAExploitAdd    string          `json:"cisaExploitAdd,omitempty"`
	CISAActionDue     string          `json:"cisaActionDue,omitempty"`
	CveTags           []CveTag        `json:"cveTags"`
	Descriptions      []LangString    `json:"descriptions"`
	References        []CveReference  `json:"references"`
	Metrics           Metrics         `json:"metrics,omitempty"`
	Weaknesses        []Weakness      `json:"weaknesses,omitempty"`
	Configurations    []Configuration `json:"configurations,omitempty"`
	VendorComments    []VendorComment `json:"vendorComments,omitempty"`
}

func (cve Cve) GetDescription(language string) string {
	for _, description := range cve.Descriptions {
		if description.Lang == language {
			return description.Value
		}
	}
	return ""
}

// GetSeverityVector tries to fetch a CvssV31 vector string from the CVE
func (cve Cve) GetSeverityVector() string {
	var vector string
	// Try to first fetch Cvssv31 score
	for _, metric := range cve.Metrics.CvssMetricV31 {
		if len(metric.CvsssData.VectorString) > 0 {
			vector = metric.CvsssData.VectorString
			break
		} else {
			continue
		}
	}
	return vector
}

type CveTag struct {
	SourceIdentifier string   `json:"sourceIdentifier,omitempty"`
	Tags             []string `json:"tags,omitempty"`
}

type LangString struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

type CveReference struct {
	Url    string   `json:"url"`
	Source string   `json:"source,omitempty"`
	Tags   []string `json:"tags.omitempty"`
}

type Metrics struct {
	CvssMetricV40 []CvssMetricV40 `json:"cvssMetricV40"`
	CvssMetricV31 []CvssMetricV31 `json:"cvssMetricV31"`
	CvssMetricV30 []CvssMetricV30 `json:"cvssMetricV30"`
	CvssMetricV2  []CvssMetricV2  `json:"cvssMetricV2"`
}

type CvssMetricV40 struct {
	Source   string  `json:"source"`
	Type     string  `json:"type"`
	CvssData CvssV40 `json:"cvssData"`
}

type CvssMetricV31 struct {
	Source              string  `json:"source"`
	Type                string  `json:"type"`
	CvsssData           CvssV31 `json:"cvssData"`
	ExploitabilityScore float64 `json:"exploitabilityScore,omitempty"`
	ImpactScore         float64 `json:"impactScore,omitempty"`
}

type CvssMetricV30 struct {
	Source              string  `json:"source"`
	Type                string  `json:"type"`
	CvssData            CvssV30 `json:"cvssData"`
	ExploitabilityScore float64 `json:"exploitabilityScore,omitempty"`
	ImpactScore         float64 `json:"impactScore,omitempty"`
}

type CvssMetricV2 struct {
	Source                  string  `json:"source"`
	Type                    string  `json:"type"`
	CvssData                CvssV20 `json:"cvssData"`
	BaseSeverity            string  `json:"baseSeverity,omitempty"`
	ExploitabilityScore     float64 `json:"exploitabilityScore,omitempty"`
	ImpactScore             float64 `json:"impactScore,omitempty"`
	ACInsufInfo             bool    `json:"acInsufInfo,omitempty"`
	ObtainAllPrivilege      bool    `json:"obtainAllPrivilege,omitempty"`
	ObtainUserPrivilege     bool    `json:"obtainUserPrivilege,omitempty"`
	ObtainOtherPrivilege    bool    `json:"obtainOtherPrivilege,omitempty"`
	UserInteractionRequired bool    `json:"userInteractionRequired,omitempty"`
}

type CvssV20 struct {
	Version                    string  `json:"version"`
	VectorString               string  `json:"vectorString"`
	AccessVector               string  `json:"accessVector,omitempty"`
	AccessComplexity           string  `json:"accessComplexity,omitempty"`
	Authentication             string  `json:"authentication,omitempty"`
	ConfidentialityImpact      string  `json:"confidentialityImpact,omitempty"`
	IntegrityImpact            string  `json:"integrityImpact,omitempty"`
	AvailabilityImpact         string  `json:"availabilityImpact,omitempty"`
	BaseScore                  float64 `json:"baseScore,omitempty"`
	Exploitability             string  `json:"exploitability,omitempty"`
	RemediationLevel           string  `json:"remediationLevel,omitempty"`
	ReportConfidence           string  `json:"reportConfidence,omitempty"`
	TemporalScore              float64 `json:"temporalScore,omitempty"`
	CollateralDamagePotential  string  `json:"collateralDamagePotential,omitempty"`
	TargetDistribution         string  `json:"targetDistribution,omitempty"`
	ConfidentialityRequirement string  `json:"confidentialityRequirement,omitempty"`
	IntegrityRequirement       string  `json:"integrityRequirement,omitempty"`
	AvailabilityRequirement    string  `json:"availabilityRequirement,omitempty"`
	EnvironmentalScore         float64 `json:"environmentalScore,omitempty"`
}

type CvssV30 struct {
	Version                       string  `json:"version"`
	VectorString                  string  `json:"vectorString"`
	AttackVector                  string  `json:"attackVector,omitempty"`
	AttackComplexity              string  `json:"attackComplexity,omitempty"`
	PrivilegesRequired            string  `json:"privilegesRequired,omitempty"`
	UserInteraction               string  `json:"userInteraction,omitempty"`
	Scope                         string  `json:"scope,omitempty"`
	ConfidentialityImpact         string  `json:"confidentialityImpact,omitempty"`
	IntegrityImpact               string  `json:"integrityImpact,omitempty"`
	AvailabilityImpact            string  `json:"availabilityImpact,omitempty"`
	BaseScore                     float64 `json:"baseScore"`
	BaseSeverity                  string  `json:"baseSeverity"`
	ExploitCodeMaturity           string  `json:"exploitCodeMaturity,omitempty"`
	RemediationLevel              string  `json:"remediationLevel,omitempty"`
	ReportConfidence              string  `json:"reportConfidence,omitempty"`
	TemporalScore                 float64 `json:"temporalScore,omitempty"`
	TemporalSeverity              string  `json:"temporalSeverity,omitempty"`
	ConfidentialityRequirement    string  `json:"confidentialityRequirement,omitempty"`
	IntegrityRequirement          string  `json:"integrityRequirement,omitempty"`
	AvailabilityRequirement       string  `json:"availabilityRequirement,omitempty"`
	ModifiedAttackVector          string  `json:"modifiedAttackVector,omitempty"`
	ModifiedAttackComplexity      string  `json:"modifiedAttackComplexity,omitempty"`
	ModifiedPrivilegesRequired    string  `json:"modifiedPrivilegesRequired,omitempty"`
	ModifiedUserInteraction       string  `json:"modifiedUserInteraction,omitempty"`
	ModifiedScope                 string  `json:"modifiedScope,omitempty"`
	ModifiedConfidentialityImpact string  `json:"modifiedConfidentialityImpact,omitempty"`
	ModifiedIntegrityImpact       string  `json:"modifiedIntegrityImpact,omitempty"`
	ModifiedAvailabilityImpact    string  `json:"modifiedAvailabilityImpact,omitempty"`
	EnvironmentalScore            float64 `json:"environmentalScore,omitempty"`
	EnvironmentalSeverity         string  `json:"environmentalSeverity,omitempty"`
}

type CvssV31 struct {
	Version                       string  `json:"version"`
	VectorString                  string  `json:"vectorString"`
	AttackVector                  string  `json:"attackVector,omitempty"`
	AttackComplexity              string  `json:"attackComplexity,omitempty"`
	PrivilegesRequired            string  `json:"privilegesRequired,omitempty"`
	UserInteraction               string  `json:"userInteraction,omitempty"`
	Scope                         string  `json:"scope,omitempty"`
	ConfidentialityImpact         string  `json:"confidentialityImpact,omitempty"`
	IntegrityImpact               string  `json:"integrityImpact,omitempty"`
	AvailabilityImpact            string  `json:"availabilityImpact,omitempty"`
	BaseScore                     float64 `json:"baseScore"`
	BaseSeverity                  string  `json:"baseSeverity"`
	ExploitCodeMaturity           string  `json:"exploitCodeMaturity,omitempty"`
	RemediationLevel              string  `json:"remediationLevel,omitempty"`
	ReportConfidence              string  `json:"reportConfidence,omitempty"`
	TemporalScore                 float64 `json:"temporalScore,omitempty"`
	TemporalSeverity              string  `json:"temporalSeverity,omitempty"`
	ConfidentialityRequirement    string  `json:"confidentialityRequirement,omitempty"`
	IntegrityRequirement          string  `json:"integrityRequirement,omitempty"`
	AvailabilityRequirement       string  `json:"availabilityRequirement,omitempty"`
	ModifiedAttackVector          string  `json:"modifiedAttackVector,omitempty"`
	ModifiedAttackComplexity      string  `json:"modifiedAttackComplexity,omitempty"`
	ModifiedPrivilegesRequired    string  `json:"modifiedPrivilegesRequired,omitempty"`
	ModifiedUserInteraction       string  `json:"modifiedUserInteraction,omitempty"`
	ModifiedScope                 string  `json:"modifiedScope,omitempty"`
	ModifiedConfidentialityImpact string  `json:"modifiedConfidentialityImpact,omitempty"`
	ModifiedIntegrityImpact       string  `json:"modifiedIntegrityImpact,omitempty"`
	ModifiedAvailabilityImpact    string  `json:"modifiedAvailabilityImpact,omitempty"`
	EnvironmentalScore            float64 `json:"environmentalScore,omitempty"`
	EnvironmentalSeverity         string  `json:"environmentalSeverity,omitempty"`
}

type CvssV40 struct {
	Version                           string `json:"version"`
	VectorString                      string `json:"vectorString"`
	AttackVector                      string `json:"attackVector,omitempty"`
	AttackComplexity                  string `json:"attackComplexity,omitempty"`
	AttackRequirements                string `json:"attackRequirements,omitempty"`
	PrivilegesRequired                string `json:"privilegesRequired,omitempty"`
	UserInteraction                   string `json:"userInteraction,omitempty"`
	VulnConfidentialityImpact         string `json:"vulnConfidentialityImpact,omitempty"`
	VulnIntegrityImpact               string `json:"vulnIntegrityImpact,omitempty"`
	VulnAvailabilityImpact            string `json:"vulnAvailabilityImpact,omitempty"`
	SubConfidentialityImpact          string `json:"subConfidentialityImpact,omitempty"`
	SubIntegrityImpact                string `json:"subIntegrityImpact,omitempty"`
	SubAvailabilityImpact             string `json:"subAvailabilityImpact,omitempty"`
	ExploitMaturity                   string `json:"exploitMaturity,omitempty"`
	ConfidentialityRequirement        string `json:"confidentialityRequirement,omitempty"`
	IntegrityRequirement              string `json:"integrityRequirement,omitempty"`
	AvailabilityRequirement           string `json:"availabilityRequirement,omitempty"`
	ModifiedAttackVector              string `json:"modifiedAttackVector,omitempty"`
	ModifiedAttackComplexity          string `json:"modifiedAttackComplexity,omitempty"`
	ModifiedAttackRequirements        string `json:"modifiedAttackRequirements,omitempty"`
	ModifiedPrivilegesRequired        string `json:"modifiedPrivilegesRequired,omitempty"`
	ModifiedUserInteracotion          string `json:"modifiedUserInteraction,omitempty"`
	ModifiedVulnConfidentialityImpact string `json:"modifiedVulnConfidentialityImpact,omitempty"`
	ModifiedVulnIntegrityImpact       string `json:"modifiedVulnIntegrityImpact,omitempty"`
	ModifiedVulnAvailabilityImpact    string `json:"modifiedVulnAvailabilityImpact,omitempty"`
	ModifiedSubConfidentialityImpact  string `json:"modifiedSubConfidentialityImpact,omitempty"`
	ModifiedSubIntegrityImpact        string `json:"modifiedSubIntegrityImpact,omitempty"`
	ModifiedSubAvailabilityImpact     string `json:"modifiedSubAvailabilityImpact,omitempty"`
	Safety                            string `json:"safety,omitempty"`
	Automatable                       string `json:"automatable,omitempty"`
	Recovery                          string `json:"recovery,omitempty"`
	ValueDensity                      string `json:"valueDensity,omitempty"`
	VulnerabilityResponseEffort       string `json:"vulnerabilityResponseEffort,omitempty"`
	ProviderUrgency                   string `json:"providerUrgency,omitempty"`
}

type Weakness struct {
	Source      string       `json:"source"`
	Type        string       `json:"type"`
	Description []LangString `json:"description"`
}

type Configurations struct {
	Nodes []Node `json:"nodes"`
}

type Node struct {
	Operator string     `json:"operator"`
	Negate   bool       `json:"negate"`
	CpeMatch []CpeMatch `json:"cpeMatch"`
}

type CpeMatch struct {
	Vulnerable            bool   `json:"vulnerable"`
	Criteria              string `json:"criteria"`
	MatchCriteriaID       string `json:"matchCriteriaId"`
	VersionStartExcluding string `json:"versionStartExcluding,omitempty"`
	VersionStartIncluding string `json:"versionStartIncluding,omitempty"`
	VersionEndExcluding   string `json:"versionEndExcluding,omitempty"`
	VersionEndIncluding   string `json:"versionEndIncluding,omitempty"`
}

type Configuration struct {
	Operator string `json:"operator,omitempty"`
	Negate   bool   `json:"negate,omitempty"`
	Nodes    []Node `json:"nodes"`
}

type VendorComment struct {
	Organization string `json:"organization"`
	Comment      string `json:"comment"`
	LastModified string `json:"lastModified"`
}

type CveFilter struct {
	// both pubStartDate and pubEndDate are required. The maximum allowable range when using any date range parameters is 120 consecutive days.
	PubStartDate string
	PubEndDate   string
	//  both lastModStartDate and lastModEndDate are required. The maximum allowable range when using any date range parameters is 120 consecutive days.
	ModStartDate string
	ModEndDate   string
}

//
// Models for Issue
//
type Issue struct {
	Id          string `json:"id"`
	Type        string `json:"type"`
	PrimaryName string `json:"primary_name"`
	Description string `json:"description"`
}

type IssueEdge struct {
	Node   *Issue  `json:"node"`
	Cursor *string `json:"cursor,omitempty"`
}

type IssueConnection struct {
	TotalCount int          `json:"totalCount"`
	Edges      []*IssueEdge `json:"edges"`
}

//
// Models for IssueVariant
//
type IssueVariant struct {
	Id            string `json:"id"`
	SecondaryName string `json:"secondary_name"`
	IssueId       int64  `json:"issue_id"`
}

//
// Models for IssueRepository
//
type IssueRepository struct {
	Id        string  `json:"id"`
	Name      *string `json:"name,omitempty"`
	URL       *string `json:"url,omitempty"`
	CreatedAt *string `json:"created_at,omitempty"`
	UpdatedAt *string `json:"updated_at,omitempty"`
}

type IssueRepositoryConnection struct {
	TotalCount int                    `json:"totalCount"`
	Edges      []*IssueRepositoryEdge `json:"edges,omitempty"`
}

type IssueRepositoryEdge struct {
	Node      *IssueRepository `json:"node"`
	Cursor    *string          `json:"cursor,omitempty"`
	Priority  *int             `json:"priority,omitempty"`
	CreatedAt *string          `json:"created_at,omitempty"`
	UpdatedAt *string          `json:"updated_at,omitempty"`
}
