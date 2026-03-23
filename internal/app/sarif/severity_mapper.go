// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package sarif

import (
	"strconv"

	"github.com/cloudoperators/heureka/internal/entity"
)

func MapSeverity(sarifLevel string, rule *SARIFRule, properties map[string]interface{}) entity.SeverityValues {
	if cvss := extractCVSSScore(properties); cvss >= 0 {
		return cvssToSeverity(cvss)
	}

	return sarifLevelToSeverity(sarifLevel, rule.DefaultConfiguration.Level)
}

func extractCVSSScore(properties map[string]interface{}) float64 {
	if properties == nil {
		return -1
	}

	if cvssStr, ok := properties["security-severity"].(string); ok {
		if cvss, err := strconv.ParseFloat(cvssStr, 64); err == nil {
			return cvss
		}
	}

	cvssNames := []string{"cvssScore", "cvss_score", "severity_score"}
	for _, name := range cvssNames {
		if cvssVal, ok := properties[name]; ok {
			switch v := cvssVal.(type) {
			case float64:
				return v
			case string:
				if cvss, err := strconv.ParseFloat(v, 64); err == nil {
					return cvss
				}
			}
		}
	}

	return -1
}

func cvssToSeverity(cvss float64) entity.SeverityValues {
	switch {
	case cvss >= 9.0:
		return entity.SeverityValuesCritical
	case cvss >= 7.0:
		return entity.SeverityValuesHigh
	case cvss >= 4.0:
		return entity.SeverityValuesMedium
	case cvss > 0:
		return entity.SeverityValuesLow
	default:
		return entity.SeverityValuesNone
	}
}

func sarifLevelToSeverity(resultLevel, ruleLevel string) entity.SeverityValues {
	level := resultLevel
	if level == "" || level == "none" {
		level = ruleLevel
	}

	switch level {
	case "error":
		return entity.SeverityValuesHigh
	case "warning":
		return entity.SeverityValuesMedium
	case "note":
		return entity.SeverityValuesLow
	default:
		return entity.SeverityValuesMedium
	}
}
