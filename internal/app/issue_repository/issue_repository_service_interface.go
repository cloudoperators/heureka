package issue_repository

import "github.wdf.sap.corp/cc/heureka/internal/entity"

type IssueRepositoryService interface {
	ListIssueRepositories(*entity.IssueRepositoryFilter, *entity.ListOptions) (*entity.List[entity.IssueRepositoryResult], error)
	CreateIssueRepository(*entity.IssueRepository) (*entity.IssueRepository, error)
	UpdateIssueRepository(*entity.IssueRepository) (*entity.IssueRepository, error)
	DeleteIssueRepository(int64) error
}
