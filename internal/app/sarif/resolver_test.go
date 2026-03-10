// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package sarif

import (
	"testing"
)

func TestPackageResolver(t *testing.T) {
	components := []ComponentMatch{
		{
			ComponentInstanceId: 10,
			PackageName:         "example/lib",
			Version:             "1.0.0",
			Purl:               "pkg:deb/debian/example/lib@1.0.0",
		},
		{
			ComponentInstanceId: 11,
			PackageName:         "another/dep",
			Version:             "2.1.0",
			Purl:               "pkg:npm/another/dep@2.1.0",
		},
	}

	resolver := NewPackageResolver(components)

	tests := []struct {
		name        string
		pkgName     string
		version     string
		expectedId  int64
		shouldFind  bool
	}{
		{
			name:       "Exact match with name and version",
			pkgName:    "example/lib",
			version:    "1.0.0",
			expectedId: 10,
			shouldFind: true,
		},
		{
			name:       "Exact match second component",
			pkgName:    "another/dep",
			version:    "2.1.0",
			expectedId: 11,
			shouldFind: true,
		},
		{
			name:       "Fallback match - only name matches (different version)",
			pkgName:    "example/lib",
			version:    "2.0.0",
			expectedId: 10,
			shouldFind: true,
		},
		{
			name:       "Not found - package doesn't exist",
			pkgName:    "unknown/package",
			version:    "1.0.0",
			expectedId: 0,
			shouldFind: false,
		},
		{
			name:       "Not found - empty package name",
			pkgName:    "",
			version:    "1.0.0",
			expectedId: 0,
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &PackageInfo{Name: tt.pkgName, Version: tt.version}
				id, found := resolver.Resolve(info)

			if found != tt.shouldFind {
				t.Errorf("expected found=%v, got %v", tt.shouldFind, found)
			}

			if found && id != tt.expectedId {
				t.Errorf("expected id=%d, got %d", tt.expectedId, id)
			}
		})
	}
}

func TestPackageResolverNilSafety(t *testing.T) {
	var resolver *PackageResolver
    id, found := resolver.Resolve(&PackageInfo{Name: "example/lib", Version: "1.0.0"})
	if found {
		t.Errorf("expected found=false for nil resolver, got true")
	}

	if id != 0 {
		t.Errorf("expected id=0 for nil resolver, got %d", id)
	}

	emptyResolver := NewPackageResolver([]ComponentMatch{})
    id, found = emptyResolver.Resolve(&PackageInfo{Name: "example/lib", Version: "1.0.0"})
	if found {
		t.Errorf("expected found=false for empty resolver, got true")
	}

	if id != 0 {
		t.Errorf("expected id=0 for empty resolver, got %d", id)
	}
}
