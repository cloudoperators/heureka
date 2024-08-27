package issue_match

import "github.wdf.sap.corp/cc/heureka/internal/entity"

type IssueMatchService interface {
	ListIssueMatches(filter *entity.IssueMatchFilter, options *entity.ListOptions) (*entity.List[entity.IssueMatchResult], error)
	GetIssueMatch(int64) (*entity.IssueMatch, error)
	CreateIssueMatch(*entity.IssueMatch) (*entity.IssueMatch, error)
	UpdateIssueMatch(*entity.IssueMatch) (*entity.IssueMatch, error)
	DeleteIssueMatch(int64) error
	AddEvidenceToIssueMatch(int64, int64) (*entity.IssueMatch, error)
	RemoveEvidenceFromIssueMatch(int64, int64) (*entity.IssueMatch, error)
}
