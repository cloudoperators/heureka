// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package sarif

// ComponentMatch represents a component instance that can be matched against package names/versions
type ComponentMatch struct {
	ComponentInstanceId int64
	PackageName        string
	Version            string
	Purl               string
}

type PackageResolver struct {
	components []ComponentMatch
}

func NewPackageResolver(components []ComponentMatch) *PackageResolver {
	return &PackageResolver{
		components: components,
	}
}

// Resolve attempts to find a ComponentInstanceId for a package with the following strategies:
// 1. Exact PURL match
// 2. Exact (Name + Version) match
// 3. Name-only match
func (pr *PackageResolver) Resolve(info *PackageInfo) (int64, bool) {
	if pr == nil || len(pr.components) == 0 || info == nil {
		return 0, false
	}

	if info.Purl != "" {
		for _, comp := range pr.components {
			if comp.Purl == info.Purl {
				return comp.ComponentInstanceId, true
			}
		}
	}

	if info.Name != "" && info.Version != "" {
		for _, comp := range pr.components {
			if comp.PackageName == info.Name && comp.Version == info.Version {
				return comp.ComponentInstanceId, true
			}
		}
	}

	if info.Name != "" {
		for _, comp := range pr.components {
			if comp.PackageName == info.Name {
				return comp.ComponentInstanceId, true
			}
		}
	}

	return 0, false
}
