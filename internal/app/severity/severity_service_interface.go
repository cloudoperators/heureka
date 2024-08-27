package severity

import "github.wdf.sap.corp/cc/heureka/internal/entity"

type SeverityService interface {
	GetSeverity(*entity.SeverityFilter) (*entity.Severity, error)
}
