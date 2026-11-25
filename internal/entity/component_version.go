// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "github.com/Masterminds/semver/v3"

type ComponentVersionFilter struct {
	Paginated
	Id            []*int64  `json:"id"`
	IssueId       []*int64  `json:"issue_id"`
	ComponentCCRN []*string `json:"component_ccrn"`
	ComponentId   []*int64  `json:"component_id"`
	// Used for ordering by severity/rating
	IssueRepositoryId []*int64          `json:"issue_repository_id"`
	Version           []*string         `json:"version"`
	ServiceId         []*int64          `json:"service_id"`
	ServiceCCRN       []*string         `json:"service_ccrn"`
	State             []StateFilterType `json:"state"`
	Tag               []*string         `json:"tag"`
	Repository        []*string         `json:"repository"`
	Organization      []*string         `json:"organization"`
	EndOfLife         []*bool           `json:"end_of_life"`
	IsOutdated        []*bool           `json:"is_outdated"`
}

func (f *ComponentVersionFilter) Get() any {
	return f
}

func (f *ComponentVersionFilter) Ensure() Filter {
	if f == nil {
		return &ComponentVersionFilter{}
	}

	return f
}

type ComponentVersionAggregations struct{}

type ComponentVersionResult struct {
	WithCursor
	*ComponentVersion
	*ComponentVersionAggregations
}

type ComponentVersion struct {
	Metadata
	Id                 int64               `json:"id"`
	Version            string              `json:"version"`
	Component          *Component          `json:"component,omitempty"`
	ComponentId        int64               `db:"componentversion_component_id"`
	Tag                string              `json:"tag"`
	Repository         string              `json:"repository"`
	Organization       string              `json:"organization"`
	ComponentInstances []ComponentInstance `json:"component_instances,omitempty"`
	Issues             []Issue             `json:"issues,omitempty"`
	EndOfLife          *bool               `json:"end_of_life,omitempty"`
	// LatestVersion is the highest semantic-version tag available in the
	// registry for this image's repository. Nil when unknown or when the
	// repository has no semver tags.
	LatestVersion *string `json:"latest_version,omitempty"`
	// IsOutdated reports whether the running Tag is behind LatestVersion. It is
	// computed and persisted on write (see IsVersionOutdated); nil means the
	// comparison is not meaningful (e.g. a floating "latest" or digest tag).
	IsOutdated *bool `json:"is_outdated,omitempty"`
}

func (cv *ComponentVersion) GetId() int64 {
	return cv.Id
}

func (cv *ComponentVersion) SetId(id int64) {
	cv.Id = id
}

// IsVersionOutdated reports whether the running image tag is behind the highest
// semantic-version tag available in the registry. It returns nil when the
// comparison is not meaningful: when there is no known latest version, or when
// either the running tag or the latest version does not parse as a semantic
// version (e.g. floating tags like "latest" or digest-style tags).
func IsVersionOutdated(tag string, latest *string) *bool {
	if latest == nil || *latest == "" || tag == "" {
		return nil
	}

	current, err := semver.NewVersion(tag)
	if err != nil {
		return nil
	}

	latestVersion, err := semver.NewVersion(*latest)
	if err != nil {
		return nil
	}

	outdated := latestVersion.GreaterThan(current)

	return &outdated
}
