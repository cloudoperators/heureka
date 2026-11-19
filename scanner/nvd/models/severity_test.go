// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package models_test

import (
	"testing"

	"github.com/cloudoperators/heureka/scanner/nvd/models"
)

const emptyCVSS = "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:N"

const (
	vectorV31 = "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"
	vectorV30 = "CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"
	vectorV40 = "CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:H/VI:H/VA:H/SC:N/SI:N/SA:N"
	vectorV2  = "AV:N/AC:L/Au:N/C:C/I:C/A:C"
)

func cveWith(m models.Metrics) models.Cve {
	return models.Cve{Metrics: m}
}

func TestSeverityVector(t *testing.T) {
	tests := []struct {
		name    string
		metrics models.Metrics
		want    string
	}{
		{
			name: "v3.1 only",
			metrics: models.Metrics{
				CvssMetricV31: []models.CvssMetricV31{{CvsssData: models.CvssV31{VectorString: vectorV31}}},
			},
			want: vectorV31,
		},
		{
			name: "v3.0 only",
			metrics: models.Metrics{
				CvssMetricV30: []models.CvssMetricV30{{CvssData: models.CvssV30{VectorString: vectorV30}}},
			},
			want: vectorV30,
		},
		{
			name: "v4.0 only",
			metrics: models.Metrics{
				CvssMetricV40: []models.CvssMetricV40{{CvssData: models.CvssV40{VectorString: vectorV40}}},
			},
			want: vectorV40,
		},
		{
			name: "v2.0 only",
			metrics: models.Metrics{
				CvssMetricV2: []models.CvssMetricV2{{CvssData: models.CvssV20{VectorString: vectorV2}}},
			},
			want: vectorV2,
		},
		{
			name: "v3.1 preferred over v4.0 when both present",
			metrics: models.Metrics{
				CvssMetricV40: []models.CvssMetricV40{{CvssData: models.CvssV40{VectorString: vectorV40}}},
				CvssMetricV31: []models.CvssMetricV31{{CvsssData: models.CvssV31{VectorString: vectorV31}}},
			},
			want: vectorV31,
		},
		{
			name: "v3.0 preferred over v4.0 and v2.0 when no v3.1",
			metrics: models.Metrics{
				CvssMetricV2:  []models.CvssMetricV2{{CvssData: models.CvssV20{VectorString: vectorV2}}},
				CvssMetricV40: []models.CvssMetricV40{{CvssData: models.CvssV40{VectorString: vectorV40}}},
				CvssMetricV30: []models.CvssMetricV30{{CvssData: models.CvssV30{VectorString: vectorV30}}},
			},
			want: vectorV30,
		},
		{
			name: "v4.0 preferred over v2.0 when no v3",
			metrics: models.Metrics{
				CvssMetricV2:  []models.CvssMetricV2{{CvssData: models.CvssV20{VectorString: vectorV2}}},
				CvssMetricV40: []models.CvssMetricV40{{CvssData: models.CvssV40{VectorString: vectorV40}}},
			},
			want: vectorV40,
		},
		{
			name:    "no metric falls back to empty CVSS",
			metrics: models.Metrics{},
			want:    emptyCVSS,
		},
		{
			name: "empty vector strings skipped, falls back to empty CVSS",
			metrics: models.Metrics{
				CvssMetricV31: []models.CvssMetricV31{{CvsssData: models.CvssV31{VectorString: ""}}},
			},
			want: emptyCVSS,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := cveWith(tc.metrics).SeverityVector()
			if got != tc.want {
				t.Errorf("SeverityVector() = %q, want %q", got, tc.want)
			}
		})
	}
}
