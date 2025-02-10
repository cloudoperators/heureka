// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type UserType int

const (
	InvalidUserType   UserType = 0
	HumanUserType     UserType = 1
	TechnicalUserType UserType = 2
)

type User struct {
	Metadata
	Id           int64    `json:"id"`
	Name         string   `json:"name"`
	UniqueUserID string   `json:"uniqueUserId"`
	Type         UserType `json:"type"`
}

type UserFilter struct {
	Paginated
	Name           []*string         `json:"name"`
	UniqueUserID   []*string         `json:"uniqueUserId"`
	Type           []UserType        `json:"type"`
	Id             []*int64          `json:"id"`
	SupportGroupId []*int64          `json:"support_group_id"`
	ServiceId      []*int64          `json:"service_id"`
	State          []StateFilterType `json:"state"`
}

type UserAggregations struct {
}

type UserResult struct {
	WithCursor
	*UserAggregations
	*User
}

func GetUserTypeFromString(uts string) UserType {
	if uts == "user" {
		return HumanUserType
	} else if uts == "technical" {
		return TechnicalUserType
	}
	return InvalidUserType
}

func GetUserTypeString(ut UserType) string {
	if ut == HumanUserType {
		return "user"
	} else if ut == TechnicalUserType {
		return "technical"
	}
	return ""
}
