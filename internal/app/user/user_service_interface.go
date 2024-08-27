package user

import "github.wdf.sap.corp/cc/heureka/internal/entity"

type UserService interface {
	ListUsers(*entity.UserFilter, *entity.ListOptions) (*entity.List[entity.UserResult], error)
	CreateUser(*entity.User) (*entity.User, error)
	UpdateUser(*entity.User) (*entity.User, error)
	DeleteUser(int64) error
	ListUserNames(*entity.UserFilter, *entity.ListOptions) ([]string, error)
	ListUniqueUserIDs(*entity.UserFilter, *entity.ListOptions) ([]string, error)
}
