// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package event

import "github.wdf.sap.corp/cc/heureka/internal/entity"

// Event names
const (
	OnIssueMatchCreateEvent             EventName = "OnIssueMatchCreate"
	OnIssueMatchUpdateEvent             EventName = "OnIssueMatchUpdate"
	OnIssueMatchDeleteEvent             EventName = "OnIssueMatchDelete"
	OnAddEvidenceToIssueMatchEvent      EventName = "OnAddEvidenceToIssueMatch"
	OnRemoveEvidenceFromIssueMatchEvent EventName = "OnRemoveEvidenceFromIssueMatch"
	OnListIssueMatchChangesEvent        EventName = "OnListIssueMatchChanges"
	OnIssueMatchChangeCreateEvent       EventName = "OnIssueMatchChangeCreate"
	OnIssueMatchChangeUpdateEvent       EventName = "OnIssueMatchChangeUpdate"
	OnIssueMatchChangeDeleteEvent       EventName = "OnIssueMatchChangeDelete"

	OnUserCreateEvent        EventName = "OnUserCreate"
	OnUserUpdateEvent        EventName = "OnUserUpdate"
	OnUserDeleteEvent        EventName = "OnUserDelete"
	OnListUsersEvent         EventName = "OnListUsers"
	OnListUserNamesEvent     EventName = "OnListUserNames"
	OnListUniqueUserIDsEvent EventName = "OnListUniqueUserIDs"

	OnSupportGroupCreateEvent            EventName = "OnSupportGroupCreate"
	OnSupportGroupUpdateEvent            EventName = "OnSupportGroupUpdate"
	OnSupportGroupDeleteEvent            EventName = "OnSupportGroupDelete"
	OnAddServiceToSupportGroupEvent      EventName = "OnAddServiceToSupportGroup"
	OnRemoveServiceFromSupportGroupEvent EventName = "OnRemoveServiceFromSupportGroup"
	OnAddUserToSupportGroupEvent         EventName = "OnAddUserToSupportGroup"
	OnRemoveUserFromSupportGroupEvent    EventName = "OnRemoveUserFromSupportGroup"
	OnListSupportGroupsEvent             EventName = "OnListSupportGroups"
	OnListSupportGroupNamesEvent         EventName = "OnListSupportGroupNames"

	OnComponentInstanceCreateEvent EventName = "OnComponentInstanceCreate"
	OnComponentInstanceUpdateEvent EventName = "OnComponentInstanceUpdate"
	OnComponentInstanceDeleteEvent EventName = "OnComponentInstanceDelete"
	OnListComponentInstancesEvent  EventName = "OnListComponentInstances"
	OnListComponentNamesEvent      EventName = "OnListComponentNames"

	OnEvidenceCreateEvent EventName = "OnEvidenceCreate"
	OnEvidenceUpdateEvent EventName = "OnEvidenceUpdate"
	OnEvidenceDeleteEvent EventName = "OnEvidenceDelete"
	OnListEvidencesEvent  EventName = "OnListEvidences"

	OnComponentCreateEvent EventName = "OnComponentCreate"
	OnComponentUpdateEvent EventName = "OnComponentUpdate"
	OnComponentDeleteEvent EventName = "OnComponentDelete"
	OnListComponentsEvent  EventName = "OnListComponents"

	OnComponentVersionCreateEvent EventName = "OnComponentVersionCreate"
	OnComponentVersionUpdateEvent EventName = "OnComponentVersionUpdate"
	OnComponentVersionDeleteEvent EventName = "OnComponentVersionDelete"
	OnListComponentVersionsEvent  EventName = "OnListComponentVersions"

	OnGetSeverityEvent EventName = "OnGetSeverity"
	OnShutdownEvent    EventName = "OnShutdown"
)

// Service events
type OnServiceCreate struct {
	Event
	Service *entity.Service
}

type OnServiceList struct {
	Event
	Filter   *entity.ServiceFilter
	Options  *entity.ListOptions
	Services *entity.List[entity.Service]
}

type OnServiceUpdate struct {
	Event
	Service *entity.Service
}

type OnServiceDelete struct {
	Event
	ServiceID int64
}

type OnAddOwnerToService struct {
	Event
	ServiceID int64
	OwnerID   int64
}

type OnRemoveOwnerFromService struct {
	Event
	ServiceID int64
	OwnerID   int64
}

type OnAddIssueRepositoryToService struct {
	Event
	ServiceID         int64
	IssueRepositoryID int64
}

type OnRemoveIssueRepositoryFromService struct {
	Event
	ServiceID         int64
	IssueRepositoryID int64
}

type OnListServiceNames struct {
	Event
	Filter  *entity.ServiceFilter
	Options *entity.ListOptions
	Names   []string
}

// IssueMatch events
type OnIssueMatchCreate struct {
	Event
	IssueMatch *entity.IssueMatch
}

type OnIssueMatchUpdate struct {
	Event
	IssueMatch *entity.IssueMatch
}

type OnIssueMatchDelete struct {
	Event
	IssueMatchID int64
}

type OnAddEvidenceToIssueMatch struct {
	Event
	IssueMatchID int64
	EvidenceID   int64
}

type OnRemoveEvidenceFromIssueMatch struct {
	Event
	IssueMatchID int64
	EvidenceID   int64
}

type OnListIssueMatchChanges struct {
	Event
	Filter  *entity.IssueMatchChangeFilter
	Options *entity.ListOptions
	Changes *entity.List[entity.IssueMatchChangeResult]
}

type OnIssueMatchChangeCreate struct {
	Event
	IssueMatchChange *entity.IssueMatchChange
}

type OnIssueMatchChangeUpdate struct {
	Event
	IssueMatchChange *entity.IssueMatchChange
}

type OnIssueMatchChangeDelete struct {
	Event
	IssueMatchChangeID int64
}

// User events
type OnUserCreate struct {
	Event
	User *entity.User
}

type OnUserUpdate struct {
	Event
	User *entity.User
}

type OnUserDelete struct {
	Event
	UserID int64
}

type OnListUsers struct {
	Event
	Filter  *entity.UserFilter
	Options *entity.ListOptions
	Users   *entity.List[entity.UserResult]
}

type OnListUserNames struct {
	Event
	Filter  *entity.UserFilter
	Options *entity.ListOptions
	Names   []string
}

type OnListUniqueUserIDs struct {
	Event
	Filter  *entity.UserFilter
	Options *entity.ListOptions
	IDs     []string
}

// SupportGroup events
type OnSupportGroupCreate struct {
	Event
	SupportGroup *entity.SupportGroup
}

type OnSupportGroupUpdate struct {
	Event
	SupportGroup *entity.SupportGroup
}

type OnSupportGroupDelete struct {
	Event
	SupportGroupID int64
}

type OnAddServiceToSupportGroup struct {
	Event
	SupportGroupID int64
	ServiceID      int64
}

type OnRemoveServiceFromSupportGroup struct {
	Event
	SupportGroupID int64
	ServiceID      int64
}

type OnAddUserToSupportGroup struct {
	Event
	SupportGroupID int64
	UserID         int64
}

type OnRemoveUserFromSupportGroup struct {
	Event
	SupportGroupID int64
	UserID         int64
}

type OnListSupportGroups struct {
	Event
	Filter  *entity.SupportGroupFilter
	Options *entity.ListOptions
	Groups  *entity.List[entity.SupportGroupResult]
}

type OnListSupportGroupNames struct {
	Event
	Filter  *entity.SupportGroupFilter
	Options *entity.ListOptions
	Names   []string
}

// ComponentInstance events
type OnComponentInstanceCreate struct {
	Event
	ComponentInstance *entity.ComponentInstance
}

type OnComponentInstanceUpdate struct {
	Event
	ComponentInstance *entity.ComponentInstance
}

type OnComponentInstanceDelete struct {
	Event
	ComponentInstanceID int64
}

type OnListComponentInstances struct {
	Event
	Filter    *entity.ComponentInstanceFilter
	Options   *entity.ListOptions
	Instances *entity.List[entity.ComponentInstanceResult]
}

type OnListComponentNames struct {
	Event
	Filter  *entity.ComponentFilter
	Options *entity.ListOptions
	Names   []string
}

// Evidence events
type OnEvidenceCreate struct {
	Event
	Evidence *entity.Evidence
}

type OnEvidenceUpdate struct {
	Event
	Evidence *entity.Evidence
}

type OnEvidenceDelete struct {
	Event
	EvidenceID int64
}

type OnListEvidences struct {
	Event
	Filter    *entity.EvidenceFilter
	Options   *entity.ListOptions
	Evidences *entity.List[entity.EvidenceResult]
}

// Component events
type OnComponentCreate struct {
	Event
	Component *entity.Component
}

type OnComponentUpdate struct {
	Event
	Component *entity.Component
}

type OnComponentDelete struct {
	Event
	ComponentID int64
}

type OnListComponents struct {
	Event
	Filter     *entity.ComponentFilter
	Options    *entity.ListOptions
	Components *entity.List[entity.ComponentResult]
}

// ComponentVersion events
type OnComponentVersionCreate struct {
	Event
	ComponentVersion *entity.ComponentVersion
}

type OnComponentVersionUpdate struct {
	Event
	ComponentVersion *entity.ComponentVersion
}

type OnComponentVersionDelete struct {
	Event
	ComponentVersionID int64
}

type OnListComponentVersions struct {
	Event
	Filter   *entity.ComponentVersionFilter
	Options  *entity.ListOptions
	Versions *entity.List[entity.ComponentVersionResult]
}

// Severity event
type OnGetSeverity struct {
	Event
	Filter   *entity.SeverityFilter
	Severity *entity.Severity
}

// Shutdown event
type OnShutdown struct {
	Event
}
