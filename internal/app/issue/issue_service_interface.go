package issue

import "github.wdf.sap.corp/cc/heureka/internal/entity"

type IssueService interface {
	ListIssues(*entity.IssueFilter, *entity.IssueListOptions) (*entity.IssueList, error)
	CreateIssue(*entity.Issue) (*entity.Issue, error)
	UpdateIssue(*entity.Issue) (*entity.Issue, error)
	DeleteIssue(int64) error
	AddComponentVersionToIssue(int64, int64) (*entity.Issue, error)
	RemoveComponentVersionFromIssue(int64, int64) (*entity.Issue, error)
	ListIssueNames(*entity.IssueFilter, *entity.ListOptions) ([]string, error)
}
