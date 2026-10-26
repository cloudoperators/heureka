// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package models_test

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/cloudoperators/heureka/scanner/nvd/models"
)

func TestParseEpssValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *float64
	}{
		{
			name:  "valid decimal string",
			input: "0.005440000",
			want:  new(0.00544),
		},
		{
			name:  "zero string",
			input: "0.000000000",
			want:  new(0.0),
		},
		{
			name:  "empty string returns nil",
			input: "",
			want:  nil,
		},
		{
			name:  "non-numeric string returns nil",
			input: "n/a",
			want:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := models.ParseEpssValue(tc.input)
			if tc.want == nil {
				if got != nil {
					t.Errorf("ParseEpssValue(%q) = %v, want nil", tc.input, *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("ParseEpssValue(%q) = nil, want %v", tc.input, *tc.want)
			}
			if math.Abs(*got-*tc.want) > 1e-9 {
				t.Errorf("ParseEpssValue(%q) = %v, want ~%v", tc.input, *got, *tc.want)
			}
		})
	}
}

func TestDerefFloat64(t *testing.T) {
	f := 3.14
	if got := models.DerefFloat64(&f); math.Abs(got-3.14) > 1e-9 {
		t.Errorf("DerefFloat64(&3.14) = %v, want 3.14", got)
	}
	if got := models.DerefFloat64(nil); got != 0.0 {
		t.Errorf("DerefFloat64(nil) = %v, want 0.0", got)
	}
}

func TestEpssResponseUnmarshal(t *testing.T) {
	raw := `{
		"status": "OK",
		"status-code": 200,
		"total": 1,
		"offset": 0,
		"limit": 100,
		"data": [
			{
				"cve": "CVE-2022-27225",
				"epss": "0.005440000",
				"percentile": "0.438950000",
				"date": "2026-10-07"
			}
		]
	}`

	var resp models.EpssResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if resp.Status != "OK" {
		t.Errorf("Status = %q, want OK", resp.Status)
	}
	if resp.Total != 1 {
		t.Errorf("Total = %d, want 1", resp.Total)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(resp.Data))
	}
	d := resp.Data[0]
	if d.Cve != "CVE-2022-27225" {
		t.Errorf("Cve = %q, want CVE-2022-27225", d.Cve)
	}
	if d.Epss != "0.005440000" {
		t.Errorf("Epss = %q, want 0.005440000", d.Epss)
	}
	if d.Percentile != "0.438950000" {
		t.Errorf("Percentile = %q, want 0.438950000", d.Percentile)
	}
}
