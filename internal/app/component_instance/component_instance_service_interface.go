package component_instance

import "github.wdf.sap.corp/cc/heureka/internal/entity"

type ComponentInstanceService interface {
	ListComponentInstances(*entity.ComponentInstanceFilter, *entity.ListOptions) (*entity.List[entity.ComponentInstanceResult], error)
	CreateComponentInstance(*entity.ComponentInstance) (*entity.ComponentInstance, error)
	UpdateComponentInstance(*entity.ComponentInstance) (*entity.ComponentInstance, error)
	DeleteComponentInstance(int64) error
}
