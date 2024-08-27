package component

import "github.wdf.sap.corp/cc/heureka/internal/entity"

type ComponentService interface {
	ListComponents(*entity.ComponentFilter, *entity.ListOptions) (*entity.List[entity.ComponentResult], error)
	CreateComponent(*entity.Component) (*entity.Component, error)
	UpdateComponent(*entity.Component) (*entity.Component, error)
	DeleteComponent(int64) error
	ListComponentNames(*entity.ComponentFilter, *entity.ListOptions) ([]string, error)
}
