// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"math"
	"testing"
)

func TestNewSeverity(t *testing.T) {
	tests := []struct {
		name         string
		vector       string
		wantValue    string
		wantScore    float64
		wantHaveBase bool
	}{
		{
			name:         "CVSS v3.1 scored via goark with full breakdown",
			vector:       "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
			wantValue:    string(SeverityValuesCritical),
			wantScore:    9.8,
			wantHaveBase: true,
		},
		{
			name:         "CVSS v3.0 scored via goark with full breakdown",
			vector:       "CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
			wantValue:    string(SeverityValuesCritical),
			wantScore:    9.8,
			wantHaveBase: true,
		},
		{
			name:         "CVSS v4.0 vulnerable-system-only is Critical 9.3",
			vector:       "CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:H/VI:H/VA:H/SC:N/SI:N/SA:N",
			wantValue:    string(SeverityValuesCritical),
			wantScore:    9.3,
			wantHaveBase: false,
		},
		{
			name:         "CVSS v4.0 maximal is Critical 10.0",
			vector:       "CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:H/VI:H/VA:H/SC:H/SI:H/SA:H",
			wantValue:    string(SeverityValuesCritical),
			wantScore:    10.0,
			wantHaveBase: false,
		},
		{
			name:         "CVSS v2.0 scored via pandatix",
			vector:       "AV:N/AC:L/Au:N/C:C/I:C/A:C",
			wantValue:    string(SeverityValuesCritical),
			wantScore:    10.0,
			wantHaveBase: false,
		},
		{
			name:         "CVSS v4.0 High band",
			vector:       "CVSS:4.0/AV:N/AC:H/AT:N/PR:N/UI:N/VC:H/VI:N/VA:N/SC:N/SI:N/SA:N",
			wantValue:    string(SeverityValuesHigh),
			wantScore:    8.2,
			wantHaveBase: false,
		},
		{
			name:         "CVSS v4.0 Medium band",
			vector:       "CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:L/VI:N/VA:N/SC:N/SI:N/SA:N",
			wantValue:    string(SeverityValuesMedium),
			wantScore:    6.9,
			wantHaveBase: false,
		},
		{
			name:         "CVSS v4.0 Low band",
			vector:       "CVSS:4.0/AV:P/AC:H/AT:P/PR:H/UI:A/VC:L/VI:N/VA:N/SC:N/SI:N/SA:N",
			wantValue:    string(SeverityValuesLow),
			wantScore:    1.0,
			wantHaveBase: false,
		},
		{
			name:         "CVSS v4.0 None band",
			vector:       "CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:N/VI:N/VA:N/SC:N/SI:N/SA:N",
			wantValue:    string(SeverityValuesNone),
			wantScore:    0.0,
			wantHaveBase: false,
		},
		{
			name:         "CVSS v2.0 High band",
			vector:       "AV:N/AC:L/Au:N/C:P/I:P/A:P",
			wantValue:    string(SeverityValuesHigh),
			wantScore:    7.5,
			wantHaveBase: false,
		},
		{
			name:         "CVSS v2.0 Medium band",
			vector:       "AV:N/AC:L/Au:N/C:P/I:N/A:N",
			wantValue:    string(SeverityValuesMedium),
			wantScore:    5.0,
			wantHaveBase: false,
		},
		{
			name:         "CVSS v2.0 Low band",
			vector:       "AV:L/AC:H/Au:N/C:P/I:N/A:N",
			wantValue:    string(SeverityValuesLow),
			wantScore:    1.2,
			wantHaveBase: false,
		},
		{
			name:         "CVSS v2.0 None band",
			vector:       "AV:N/AC:L/Au:N/C:N/I:N/A:N",
			wantValue:    string(SeverityValuesNone),
			wantScore:    0.0,
			wantHaveBase: false,
		},
		{
			name:         "unparseable vector yields unknown 0.0",
			vector:       "not-a-cvss-vector",
			wantValue:    "unknown",
			wantScore:    0.0,
			wantHaveBase: false,
		},
		{
			name:         "empty vector yields unknown 0.0",
			vector:       "",
			wantValue:    "unknown",
			wantScore:    0.0,
			wantHaveBase: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NewSeverity(tc.vector)

			if got.Value != tc.wantValue {
				t.Errorf("Value = %q, want %q", got.Value, tc.wantValue)
			}

			if math.Abs(got.Score-tc.wantScore) > 1e-9 {
				t.Errorf("Score = %v, want %v", got.Score, tc.wantScore)
			}

			if tc.wantHaveBase && got.Cvss.Base == nil {
				t.Errorf("Cvss.Base = nil, want populated goark breakdown")
			}

			if !tc.wantHaveBase && got.Cvss.Base != nil {
				t.Errorf("Cvss.Base populated, want nil")
			}
		})
	}
}

func TestNewSeverityV2PrefixStripped(t *testing.T) {
	got := NewSeverity("CVSS:2.0/AV:N/AC:L/Au:N/C:C/I:C/A:C")
	if got.Value == "unknown" || got.Score == 0.0 {
		t.Errorf("prefixed v2.0 vector not parsed: got value=%q score=%v", got.Value, got.Score)
	}
}

func TestRatingFromScore(t *testing.T) {
	tests := []struct {
		score float64
		want  string
	}{
		{0.0, string(SeverityValuesNone)},
		{0.09, string(SeverityValuesNone)},
		{0.1, string(SeverityValuesLow)},
		{3.9, string(SeverityValuesLow)},
		{4.0, string(SeverityValuesMedium)},
		{6.9, string(SeverityValuesMedium)},
		{7.0, string(SeverityValuesHigh)},
		{8.9, string(SeverityValuesHigh)},
		{9.0, string(SeverityValuesCritical)},
		{10.0, string(SeverityValuesCritical)},
	}

	for _, tc := range tests {
		if got := ratingFromScore(tc.score); got != tc.want {
			t.Errorf("ratingFromScore(%.2f) = %q, want %q", tc.score, got, tc.want)
		}
	}
}
