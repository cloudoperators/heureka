// Code generated by github.com/Khan/genqlient, DO NOT EDIT.

package client

import (
	"context"

	"github.com/Khan/genqlient/graphql"
)

// AddComponentVersionToIssueResponse is returned by AddComponentVersionToIssue on success.
type AddComponentVersionToIssueResponse struct {
	AddComponentVersionToIssue *Issue `json:"addComponentVersionToIssue"`
}

// GetAddComponentVersionToIssue returns AddComponentVersionToIssueResponse.AddComponentVersionToIssue, and is useful for accessing the field via an interface.
func (v *AddComponentVersionToIssueResponse) GetAddComponentVersionToIssue() *Issue {
	return v.AddComponentVersionToIssue
}

// CompleteScannerRunResponse is returned by CompleteScannerRun on success.
type CompleteScannerRunResponse struct {
	CompleteScannerRun bool `json:"completeScannerRun"`
}

// GetCompleteScannerRun returns CompleteScannerRunResponse.CompleteScannerRun, and is useful for accessing the field via an interface.
func (v *CompleteScannerRunResponse) GetCompleteScannerRun() bool { return v.CompleteScannerRun }

// Component includes the requested fields of the GraphQL type Component.
type Component struct {
	Id   string              `json:"id"`
	Ccrn string              `json:"ccrn"`
	Type ComponentTypeValues `json:"type"`
}

// GetId returns Component.Id, and is useful for accessing the field via an interface.
func (v *Component) GetId() string { return v.Id }

// GetCcrn returns Component.Ccrn, and is useful for accessing the field via an interface.
func (v *Component) GetCcrn() string { return v.Ccrn }

// GetType returns Component.Type, and is useful for accessing the field via an interface.
func (v *Component) GetType() ComponentTypeValues { return v.Type }

// ComponentAggregate includes the requested fields of the GraphQL type Component.
type ComponentAggregate struct {
	Id                string              `json:"id"`
	Ccrn              string              `json:"ccrn"`
	Type              ComponentTypeValues `json:"type"`
	ComponentVersions *ComponentVersions  `json:"componentVersions"`
}

// GetId returns ComponentAggregate.Id, and is useful for accessing the field via an interface.
func (v *ComponentAggregate) GetId() string { return v.Id }

// GetCcrn returns ComponentAggregate.Ccrn, and is useful for accessing the field via an interface.
func (v *ComponentAggregate) GetCcrn() string { return v.Ccrn }

// GetType returns ComponentAggregate.Type, and is useful for accessing the field via an interface.
func (v *ComponentAggregate) GetType() ComponentTypeValues { return v.Type }

// GetComponentVersions returns ComponentAggregate.ComponentVersions, and is useful for accessing the field via an interface.
func (v *ComponentAggregate) GetComponentVersions() *ComponentVersions { return v.ComponentVersions }

// ComponentConnection includes the requested fields of the GraphQL type ComponentConnection.
type ComponentConnection struct {
	TotalCount int                                      `json:"totalCount"`
	Edges      []*ComponentConnectionEdgesComponentEdge `json:"edges"`
}

// GetTotalCount returns ComponentConnection.TotalCount, and is useful for accessing the field via an interface.
func (v *ComponentConnection) GetTotalCount() int { return v.TotalCount }

// GetEdges returns ComponentConnection.Edges, and is useful for accessing the field via an interface.
func (v *ComponentConnection) GetEdges() []*ComponentConnectionEdgesComponentEdge { return v.Edges }

// ComponentConnectionEdgesComponentEdge includes the requested fields of the GraphQL type ComponentEdge.
type ComponentConnectionEdgesComponentEdge struct {
	Node   *ComponentAggregate `json:"node"`
	Cursor string              `json:"cursor"`
}

// GetNode returns ComponentConnectionEdgesComponentEdge.Node, and is useful for accessing the field via an interface.
func (v *ComponentConnectionEdgesComponentEdge) GetNode() *ComponentAggregate { return v.Node }

// GetCursor returns ComponentConnectionEdgesComponentEdge.Cursor, and is useful for accessing the field via an interface.
func (v *ComponentConnectionEdgesComponentEdge) GetCursor() string { return v.Cursor }

type ComponentFilter struct {
	ComponentCcrn []string `json:"componentCcrn"`
}

// GetComponentCcrn returns ComponentFilter.ComponentCcrn, and is useful for accessing the field via an interface.
func (v *ComponentFilter) GetComponentCcrn() []string { return v.ComponentCcrn }

type ComponentInput struct {
	Ccrn string              `json:"ccrn"`
	Type ComponentTypeValues `json:"type"`
}

// GetCcrn returns ComponentInput.Ccrn, and is useful for accessing the field via an interface.
func (v *ComponentInput) GetCcrn() string { return v.Ccrn }

// GetType returns ComponentInput.Type, and is useful for accessing the field via an interface.
func (v *ComponentInput) GetType() ComponentTypeValues { return v.Type }

type ComponentTypeValues string

const (
	ComponentTypeValuesContainerimage      ComponentTypeValues = "containerImage"
	ComponentTypeValuesVirtualmachineimage ComponentTypeValues = "virtualMachineImage"
	ComponentTypeValuesRepository          ComponentTypeValues = "repository"
)

// ComponentVersion includes the requested fields of the GraphQL type ComponentVersion.
type ComponentVersion struct {
	Id          string `json:"id"`
	Version     string `json:"version"`
	ComponentId string `json:"componentId"`
}

// GetId returns ComponentVersion.Id, and is useful for accessing the field via an interface.
func (v *ComponentVersion) GetId() string { return v.Id }

// GetVersion returns ComponentVersion.Version, and is useful for accessing the field via an interface.
func (v *ComponentVersion) GetVersion() string { return v.Version }

// GetComponentId returns ComponentVersion.ComponentId, and is useful for accessing the field via an interface.
func (v *ComponentVersion) GetComponentId() string { return v.ComponentId }

// ComponentVersionConnection includes the requested fields of the GraphQL type ComponentVersionConnection.
type ComponentVersionConnection struct {
	Edges []*ComponentVersionConnectionEdgesComponentVersionEdge `json:"edges"`
}

// GetEdges returns ComponentVersionConnection.Edges, and is useful for accessing the field via an interface.
func (v *ComponentVersionConnection) GetEdges() []*ComponentVersionConnectionEdgesComponentVersionEdge {
	return v.Edges
}

// ComponentVersionConnectionEdgesComponentVersionEdge includes the requested fields of the GraphQL type ComponentVersionEdge.
type ComponentVersionConnectionEdgesComponentVersionEdge struct {
	Node *ComponentVersion `json:"node"`
}

// GetNode returns ComponentVersionConnectionEdgesComponentVersionEdge.Node, and is useful for accessing the field via an interface.
func (v *ComponentVersionConnectionEdgesComponentVersionEdge) GetNode() *ComponentVersion {
	return v.Node
}

type ComponentVersionFilter struct {
	ComponentId   []string `json:"componentId"`
	ComponentCcrn []string `json:"componentCcrn"`
	IssueId       []string `json:"issueId"`
	Version       []string `json:"version"`
}

// GetComponentId returns ComponentVersionFilter.ComponentId, and is useful for accessing the field via an interface.
func (v *ComponentVersionFilter) GetComponentId() []string { return v.ComponentId }

// GetComponentCcrn returns ComponentVersionFilter.ComponentCcrn, and is useful for accessing the field via an interface.
func (v *ComponentVersionFilter) GetComponentCcrn() []string { return v.ComponentCcrn }

// GetIssueId returns ComponentVersionFilter.IssueId, and is useful for accessing the field via an interface.
func (v *ComponentVersionFilter) GetIssueId() []string { return v.IssueId }

// GetVersion returns ComponentVersionFilter.Version, and is useful for accessing the field via an interface.
func (v *ComponentVersionFilter) GetVersion() []string { return v.Version }

type ComponentVersionInput struct {
	Version     string `json:"version"`
	ComponentId string `json:"componentId"`
}

// GetVersion returns ComponentVersionInput.Version, and is useful for accessing the field via an interface.
func (v *ComponentVersionInput) GetVersion() string { return v.Version }

// GetComponentId returns ComponentVersionInput.ComponentId, and is useful for accessing the field via an interface.
func (v *ComponentVersionInput) GetComponentId() string { return v.ComponentId }

// ComponentVersions includes the requested fields of the GraphQL type ComponentVersionConnection.
type ComponentVersions struct {
	Edges []*ComponentVersionsEdgesComponentVersionEdge `json:"edges"`
}

// GetEdges returns ComponentVersions.Edges, and is useful for accessing the field via an interface.
func (v *ComponentVersions) GetEdges() []*ComponentVersionsEdgesComponentVersionEdge { return v.Edges }

// ComponentVersionsEdgesComponentVersionEdge includes the requested fields of the GraphQL type ComponentVersionEdge.
type ComponentVersionsEdgesComponentVersionEdge struct {
	Node *ComponentVersion `json:"node"`
}

// GetNode returns ComponentVersionsEdgesComponentVersionEdge.Node, and is useful for accessing the field via an interface.
func (v *ComponentVersionsEdgesComponentVersionEdge) GetNode() *ComponentVersion { return v.Node }

// CreateComponentResponse is returned by CreateComponent on success.
type CreateComponentResponse struct {
	CreateComponent *Component `json:"createComponent"`
}

// GetCreateComponent returns CreateComponentResponse.CreateComponent, and is useful for accessing the field via an interface.
func (v *CreateComponentResponse) GetCreateComponent() *Component { return v.CreateComponent }

// CreateComponentVersionResponse is returned by CreateComponentVersion on success.
type CreateComponentVersionResponse struct {
	CreateComponentVersion *ComponentVersion `json:"createComponentVersion"`
}

// GetCreateComponentVersion returns CreateComponentVersionResponse.CreateComponentVersion, and is useful for accessing the field via an interface.
func (v *CreateComponentVersionResponse) GetCreateComponentVersion() *ComponentVersion {
	return v.CreateComponentVersion
}

// CreateIssueResponse is returned by CreateIssue on success.
type CreateIssueResponse struct {
	CreateIssue *Issue `json:"createIssue"`
}

// GetCreateIssue returns CreateIssueResponse.CreateIssue, and is useful for accessing the field via an interface.
func (v *CreateIssueResponse) GetCreateIssue() *Issue { return v.CreateIssue }

// CreateScannerRunCreateScannerRun includes the requested fields of the GraphQL type ScannerRun.
type CreateScannerRunCreateScannerRun struct {
	Id   string `json:"id"`
	Tag  string `json:"tag"`
	Uuid string `json:"uuid"`
}

// GetId returns CreateScannerRunCreateScannerRun.Id, and is useful for accessing the field via an interface.
func (v *CreateScannerRunCreateScannerRun) GetId() string { return v.Id }

// GetTag returns CreateScannerRunCreateScannerRun.Tag, and is useful for accessing the field via an interface.
func (v *CreateScannerRunCreateScannerRun) GetTag() string { return v.Tag }

// GetUuid returns CreateScannerRunCreateScannerRun.Uuid, and is useful for accessing the field via an interface.
func (v *CreateScannerRunCreateScannerRun) GetUuid() string { return v.Uuid }

// CreateScannerRunResponse is returned by CreateScannerRun on success.
type CreateScannerRunResponse struct {
	CreateScannerRun *CreateScannerRunCreateScannerRun `json:"createScannerRun"`
}

// GetCreateScannerRun returns CreateScannerRunResponse.CreateScannerRun, and is useful for accessing the field via an interface.
func (v *CreateScannerRunResponse) GetCreateScannerRun() *CreateScannerRunCreateScannerRun {
	return v.CreateScannerRun
}

// Issue includes the requested fields of the GraphQL type Issue.
type Issue struct {
	Id          string     `json:"id"`
	PrimaryName string     `json:"primaryName"`
	Description string     `json:"description"`
	Type        IssueTypes `json:"type"`
}

// GetId returns Issue.Id, and is useful for accessing the field via an interface.
func (v *Issue) GetId() string { return v.Id }

// GetPrimaryName returns Issue.PrimaryName, and is useful for accessing the field via an interface.
func (v *Issue) GetPrimaryName() string { return v.PrimaryName }

// GetDescription returns Issue.Description, and is useful for accessing the field via an interface.
func (v *Issue) GetDescription() string { return v.Description }

// GetType returns Issue.Type, and is useful for accessing the field via an interface.
func (v *Issue) GetType() IssueTypes { return v.Type }

// IssueConnection includes the requested fields of the GraphQL type IssueConnection.
type IssueConnection struct {
	Edges []*IssueConnectionEdgesIssueEdge `json:"edges"`
}

// GetEdges returns IssueConnection.Edges, and is useful for accessing the field via an interface.
func (v *IssueConnection) GetEdges() []*IssueConnectionEdgesIssueEdge { return v.Edges }

// IssueConnectionEdgesIssueEdge includes the requested fields of the GraphQL type IssueEdge.
type IssueConnectionEdgesIssueEdge struct {
	Node *Issue `json:"node"`
}

// GetNode returns IssueConnectionEdgesIssueEdge.Node, and is useful for accessing the field via an interface.
func (v *IssueConnectionEdgesIssueEdge) GetNode() *Issue { return v.Node }

type IssueFilter struct {
	AffectedService    []string                 `json:"affectedService"`
	PrimaryName        []string                 `json:"primaryName"`
	IssueMatchStatus   []IssueMatchStatusValues `json:"issueMatchStatus"`
	IssueType          []IssueTypes             `json:"issueType"`
	ComponentVersionId []string                 `json:"componentVersionId"`
	Search             []string                 `json:"search"`
}

// GetAffectedService returns IssueFilter.AffectedService, and is useful for accessing the field via an interface.
func (v *IssueFilter) GetAffectedService() []string { return v.AffectedService }

// GetPrimaryName returns IssueFilter.PrimaryName, and is useful for accessing the field via an interface.
func (v *IssueFilter) GetPrimaryName() []string { return v.PrimaryName }

// GetIssueMatchStatus returns IssueFilter.IssueMatchStatus, and is useful for accessing the field via an interface.
func (v *IssueFilter) GetIssueMatchStatus() []IssueMatchStatusValues { return v.IssueMatchStatus }

// GetIssueType returns IssueFilter.IssueType, and is useful for accessing the field via an interface.
func (v *IssueFilter) GetIssueType() []IssueTypes { return v.IssueType }

// GetComponentVersionId returns IssueFilter.ComponentVersionId, and is useful for accessing the field via an interface.
func (v *IssueFilter) GetComponentVersionId() []string { return v.ComponentVersionId }

// GetSearch returns IssueFilter.Search, and is useful for accessing the field via an interface.
func (v *IssueFilter) GetSearch() []string { return v.Search }

type IssueInput struct {
	PrimaryName string     `json:"primaryName"`
	Description string     `json:"description"`
	Uuid        string     `json:"uuid"`
	Type        IssueTypes `json:"type"`
}

// GetPrimaryName returns IssueInput.PrimaryName, and is useful for accessing the field via an interface.
func (v *IssueInput) GetPrimaryName() string { return v.PrimaryName }

// GetDescription returns IssueInput.Description, and is useful for accessing the field via an interface.
func (v *IssueInput) GetDescription() string { return v.Description }

// GetUuid returns IssueInput.Uuid, and is useful for accessing the field via an interface.
func (v *IssueInput) GetUuid() string { return v.Uuid }

// GetType returns IssueInput.Type, and is useful for accessing the field via an interface.
func (v *IssueInput) GetType() IssueTypes { return v.Type }

type IssueMatchStatusValues string

const (
	IssueMatchStatusValuesNew           IssueMatchStatusValues = "new"
	IssueMatchStatusValuesRiskAccepted  IssueMatchStatusValues = "risk_accepted"
	IssueMatchStatusValuesFalsePositive IssueMatchStatusValues = "false_positive"
	IssueMatchStatusValuesMitigated     IssueMatchStatusValues = "mitigated"
)

type IssueTypes string

const (
	IssueTypesVulnerability   IssueTypes = "Vulnerability"
	IssueTypesPolicyviolation IssueTypes = "PolicyViolation"
	IssueTypesSecurityevent   IssueTypes = "SecurityEvent"
)

// ListComponentVersionsResponse is returned by ListComponentVersions on success.
type ListComponentVersionsResponse struct {
	ComponentVersions *ComponentVersionConnection `json:"ComponentVersions"`
}

// GetComponentVersions returns ListComponentVersionsResponse.ComponentVersions, and is useful for accessing the field via an interface.
func (v *ListComponentVersionsResponse) GetComponentVersions() *ComponentVersionConnection {
	return v.ComponentVersions
}

// ListComponentsResponse is returned by ListComponents on success.
type ListComponentsResponse struct {
	Components *ComponentConnection `json:"Components"`
}

// GetComponents returns ListComponentsResponse.Components, and is useful for accessing the field via an interface.
func (v *ListComponentsResponse) GetComponents() *ComponentConnection { return v.Components }

// ListIssuesResponse is returned by ListIssues on success.
type ListIssuesResponse struct {
	Issues *IssueConnection `json:"Issues"`
}

// GetIssues returns ListIssuesResponse.Issues, and is useful for accessing the field via an interface.
func (v *ListIssuesResponse) GetIssues() *IssueConnection { return v.Issues }

type ScannerRunInput struct {
	Uuid string `json:"uuid"`
	Tag  string `json:"tag"`
}

// GetUuid returns ScannerRunInput.Uuid, and is useful for accessing the field via an interface.
func (v *ScannerRunInput) GetUuid() string { return v.Uuid }

// GetTag returns ScannerRunInput.Tag, and is useful for accessing the field via an interface.
func (v *ScannerRunInput) GetTag() string { return v.Tag }

// __AddComponentVersionToIssueInput is used internally by genqlient
type __AddComponentVersionToIssueInput struct {
	IssueId            string `json:"issueId"`
	ComponentVersionId string `json:"componentVersionId"`
}

// GetIssueId returns __AddComponentVersionToIssueInput.IssueId, and is useful for accessing the field via an interface.
func (v *__AddComponentVersionToIssueInput) GetIssueId() string { return v.IssueId }

// GetComponentVersionId returns __AddComponentVersionToIssueInput.ComponentVersionId, and is useful for accessing the field via an interface.
func (v *__AddComponentVersionToIssueInput) GetComponentVersionId() string {
	return v.ComponentVersionId
}

// __CompleteScannerRunInput is used internally by genqlient
type __CompleteScannerRunInput struct {
	Uuid string `json:"uuid"`
}

// GetUuid returns __CompleteScannerRunInput.Uuid, and is useful for accessing the field via an interface.
func (v *__CompleteScannerRunInput) GetUuid() string { return v.Uuid }

// __CreateComponentInput is used internally by genqlient
type __CreateComponentInput struct {
	Input *ComponentInput `json:"input,omitempty"`
}

// GetInput returns __CreateComponentInput.Input, and is useful for accessing the field via an interface.
func (v *__CreateComponentInput) GetInput() *ComponentInput { return v.Input }

// __CreateComponentVersionInput is used internally by genqlient
type __CreateComponentVersionInput struct {
	Input *ComponentVersionInput `json:"input,omitempty"`
}

// GetInput returns __CreateComponentVersionInput.Input, and is useful for accessing the field via an interface.
func (v *__CreateComponentVersionInput) GetInput() *ComponentVersionInput { return v.Input }

// __CreateIssueInput is used internally by genqlient
type __CreateIssueInput struct {
	Input *IssueInput `json:"input,omitempty"`
}

// GetInput returns __CreateIssueInput.Input, and is useful for accessing the field via an interface.
func (v *__CreateIssueInput) GetInput() *IssueInput { return v.Input }

// __CreateScannerRunInput is used internally by genqlient
type __CreateScannerRunInput struct {
	Input *ScannerRunInput `json:"input,omitempty"`
}

// GetInput returns __CreateScannerRunInput.Input, and is useful for accessing the field via an interface.
func (v *__CreateScannerRunInput) GetInput() *ScannerRunInput { return v.Input }

// __ListComponentVersionsInput is used internally by genqlient
type __ListComponentVersionsInput struct {
	Filter *ComponentVersionFilter `json:"filter,omitempty"`
}

// GetFilter returns __ListComponentVersionsInput.Filter, and is useful for accessing the field via an interface.
func (v *__ListComponentVersionsInput) GetFilter() *ComponentVersionFilter { return v.Filter }

// __ListComponentsInput is used internally by genqlient
type __ListComponentsInput struct {
	Filter *ComponentFilter `json:"filter,omitempty"`
	First  int              `json:"first"`
	After  string           `json:"after"`
}

// GetFilter returns __ListComponentsInput.Filter, and is useful for accessing the field via an interface.
func (v *__ListComponentsInput) GetFilter() *ComponentFilter { return v.Filter }

// GetFirst returns __ListComponentsInput.First, and is useful for accessing the field via an interface.
func (v *__ListComponentsInput) GetFirst() int { return v.First }

// GetAfter returns __ListComponentsInput.After, and is useful for accessing the field via an interface.
func (v *__ListComponentsInput) GetAfter() string { return v.After }

// __ListIssuesInput is used internally by genqlient
type __ListIssuesInput struct {
	Filter *IssueFilter `json:"filter,omitempty"`
	First  int          `json:"first"`
}

// GetFilter returns __ListIssuesInput.Filter, and is useful for accessing the field via an interface.
func (v *__ListIssuesInput) GetFilter() *IssueFilter { return v.Filter }

// GetFirst returns __ListIssuesInput.First, and is useful for accessing the field via an interface.
func (v *__ListIssuesInput) GetFirst() int { return v.First }

// The query or mutation executed by AddComponentVersionToIssue.
const AddComponentVersionToIssue_Operation = `
mutation AddComponentVersionToIssue ($issueId: ID!, $componentVersionId: ID!) {
	addComponentVersionToIssue(issueId: $issueId, componentVersionId: $componentVersionId) {
		id
		primaryName
		description
		type
	}
}
`

func AddComponentVersionToIssue(
	ctx_ context.Context,
	client_ graphql.Client,
	issueId string,
	componentVersionId string,
) (*AddComponentVersionToIssueResponse, error) {
	req_ := &graphql.Request{
		OpName: "AddComponentVersionToIssue",
		Query:  AddComponentVersionToIssue_Operation,
		Variables: &__AddComponentVersionToIssueInput{
			IssueId:            issueId,
			ComponentVersionId: componentVersionId,
		},
	}
	var err_ error

	var data_ AddComponentVersionToIssueResponse
	resp_ := &graphql.Response{Data: &data_}

	err_ = client_.MakeRequest(
		ctx_,
		req_,
		resp_,
	)

	return &data_, err_
}

// The query or mutation executed by CompleteScannerRun.
const CompleteScannerRun_Operation = `
mutation CompleteScannerRun ($uuid: String!) {
	completeScannerRun(uuid: $uuid)
}
`

func CompleteScannerRun(
	ctx_ context.Context,
	client_ graphql.Client,
	uuid string,
) (*CompleteScannerRunResponse, error) {
	req_ := &graphql.Request{
		OpName: "CompleteScannerRun",
		Query:  CompleteScannerRun_Operation,
		Variables: &__CompleteScannerRunInput{
			Uuid: uuid,
		},
	}
	var err_ error

	var data_ CompleteScannerRunResponse
	resp_ := &graphql.Response{Data: &data_}

	err_ = client_.MakeRequest(
		ctx_,
		req_,
		resp_,
	)

	return &data_, err_
}

// The query or mutation executed by CreateComponent.
const CreateComponent_Operation = `
mutation CreateComponent ($input: ComponentInput!) {
	createComponent(input: $input) {
		id
		ccrn
		type
	}
}
`

func CreateComponent(
	ctx_ context.Context,
	client_ graphql.Client,
	input *ComponentInput,
) (*CreateComponentResponse, error) {
	req_ := &graphql.Request{
		OpName: "CreateComponent",
		Query:  CreateComponent_Operation,
		Variables: &__CreateComponentInput{
			Input: input,
		},
	}
	var err_ error

	var data_ CreateComponentResponse
	resp_ := &graphql.Response{Data: &data_}

	err_ = client_.MakeRequest(
		ctx_,
		req_,
		resp_,
	)

	return &data_, err_
}

// The query or mutation executed by CreateComponentVersion.
const CreateComponentVersion_Operation = `
mutation CreateComponentVersion ($input: ComponentVersionInput!) {
	createComponentVersion(input: $input) {
		id
		version
		componentId
	}
}
`

func CreateComponentVersion(
	ctx_ context.Context,
	client_ graphql.Client,
	input *ComponentVersionInput,
) (*CreateComponentVersionResponse, error) {
	req_ := &graphql.Request{
		OpName: "CreateComponentVersion",
		Query:  CreateComponentVersion_Operation,
		Variables: &__CreateComponentVersionInput{
			Input: input,
		},
	}
	var err_ error

	var data_ CreateComponentVersionResponse
	resp_ := &graphql.Response{Data: &data_}

	err_ = client_.MakeRequest(
		ctx_,
		req_,
		resp_,
	)

	return &data_, err_
}

// The query or mutation executed by CreateIssue.
const CreateIssue_Operation = `
mutation CreateIssue ($input: IssueInput!) {
	createIssue(input: $input) {
		id
		primaryName
		description
		type
	}
}
`

func CreateIssue(
	ctx_ context.Context,
	client_ graphql.Client,
	input *IssueInput,
) (*CreateIssueResponse, error) {
	req_ := &graphql.Request{
		OpName: "CreateIssue",
		Query:  CreateIssue_Operation,
		Variables: &__CreateIssueInput{
			Input: input,
		},
	}
	var err_ error

	var data_ CreateIssueResponse
	resp_ := &graphql.Response{Data: &data_}

	err_ = client_.MakeRequest(
		ctx_,
		req_,
		resp_,
	)

	return &data_, err_
}

// The query or mutation executed by CreateScannerRun.
const CreateScannerRun_Operation = `
mutation CreateScannerRun ($input: ScannerRunInput!) {
	createScannerRun(input: $input) {
		id
		tag
		uuid
	}
}
`

func CreateScannerRun(
	ctx_ context.Context,
	client_ graphql.Client,
	input *ScannerRunInput,
) (*CreateScannerRunResponse, error) {
	req_ := &graphql.Request{
		OpName: "CreateScannerRun",
		Query:  CreateScannerRun_Operation,
		Variables: &__CreateScannerRunInput{
			Input: input,
		},
	}
	var err_ error

	var data_ CreateScannerRunResponse
	resp_ := &graphql.Response{Data: &data_}

	err_ = client_.MakeRequest(
		ctx_,
		req_,
		resp_,
	)

	return &data_, err_
}

// The query or mutation executed by ListComponentVersions.
const ListComponentVersions_Operation = `
query ListComponentVersions ($filter: ComponentVersionFilter) {
	ComponentVersions(filter: $filter) {
		edges {
			node {
				id
				version
				componentId
			}
		}
	}
}
`

func ListComponentVersions(
	ctx_ context.Context,
	client_ graphql.Client,
	filter *ComponentVersionFilter,
) (*ListComponentVersionsResponse, error) {
	req_ := &graphql.Request{
		OpName: "ListComponentVersions",
		Query:  ListComponentVersions_Operation,
		Variables: &__ListComponentVersionsInput{
			Filter: filter,
		},
	}
	var err_ error

	var data_ ListComponentVersionsResponse
	resp_ := &graphql.Response{Data: &data_}

	err_ = client_.MakeRequest(
		ctx_,
		req_,
		resp_,
	)

	return &data_, err_
}

// The query or mutation executed by ListComponents.
const ListComponents_Operation = `
query ListComponents ($filter: ComponentFilter, $first: Int, $after: String) {
	Components(filter: $filter, first: $first, after: $after) {
		totalCount
		edges {
			node {
				id
				ccrn
				type
				componentVersions {
					edges {
						node {
							id
							version
							componentId
						}
					}
				}
			}
			cursor
		}
	}
}
`

func ListComponents(
	ctx_ context.Context,
	client_ graphql.Client,
	filter *ComponentFilter,
	first int,
	after string,
) (*ListComponentsResponse, error) {
	req_ := &graphql.Request{
		OpName: "ListComponents",
		Query:  ListComponents_Operation,
		Variables: &__ListComponentsInput{
			Filter: filter,
			First:  first,
			After:  after,
		},
	}
	var err_ error

	var data_ ListComponentsResponse
	resp_ := &graphql.Response{Data: &data_}

	err_ = client_.MakeRequest(
		ctx_,
		req_,
		resp_,
	)

	return &data_, err_
}

// The query or mutation executed by ListIssues.
const ListIssues_Operation = `
query ListIssues ($filter: IssueFilter, $first: Int) {
	Issues(filter: $filter, first: $first) {
		edges {
			node {
				id
				primaryName
				description
				type
			}
		}
	}
}
`

func ListIssues(
	ctx_ context.Context,
	client_ graphql.Client,
	filter *IssueFilter,
	first int,
) (*ListIssuesResponse, error) {
	req_ := &graphql.Request{
		OpName: "ListIssues",
		Query:  ListIssues_Operation,
		Variables: &__ListIssuesInput{
			Filter: filter,
			First:  first,
		},
	}
	var err_ error

	var data_ ListIssuesResponse
	resp_ := &graphql.Response{Data: &data_}

	err_ = client_.MakeRequest(
		ctx_,
		req_,
		resp_,
	)

	return &data_, err_
}
