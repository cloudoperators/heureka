package component_version

import "github.wdf.sap.corp/cc/heureka/internal/entity"

type ComponentVersionService interface {
	ListComponentVersions(*entity.ComponentVersionFilter, *entity.ListOptions) (*entity.List[entity.ComponentVersionResult], error)
	CreateComponentVersion(*entity.ComponentVersion) (*entity.ComponentVersion, error)
	UpdateComponentVersion(*entity.ComponentVersion) (*entity.ComponentVersion, error)
	DeleteComponentVersion(int64) error
}
