// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"time"

	"github.com/cloudoperators/heureka/internal/entity"
)

// RemediationTimelineConfig holds the configuration for remediation timelines
type RemediationTimelineConfig struct {
	LowDays      int
	MediumDays   int
	HighDays     int
	CriticalDays int
	DefaultDays  int
}

// DefaultRemediationConfig provides default values for remediation timelines
// TODO: Make these values configurable (allow to be specified via ENV variables)
var DefaultRemediationConfig = RemediationTimelineConfig{
	LowDays:      180, // 6 months
	MediumDays:   90,  // 3 months
	HighDays:     20,
	CriticalDays: 7,
	DefaultDays:  365, // 1 year
}

// GetTargetRemediationTimeline calculates the target remediation date based on severity
// and creation date. It uses the DefaultRemediationConfig if no custom config is provided.
func GetTargetRemediationTimeline(severity entity.Severity, creationDate time.Time, config *RemediationTimelineConfig) time.Time {
	if config == nil {
		config = &DefaultRemediationConfig
	}

	var days int
	switch entity.SeverityValues(severity.Value) {
	case entity.SeverityValuesLow:
		days = config.LowDays
	case entity.SeverityValuesMedium:
		days = config.MediumDays
	case entity.SeverityValuesHigh:
		days = config.HighDays
	case entity.SeverityValuesCritical:
		days = config.CriticalDays
	default:
		days = config.DefaultDays
	}

	return creationDate.AddDate(0, 0, days)
}
