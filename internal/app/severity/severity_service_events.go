package severity

import (
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

const (
	GetSeverityEventName event.EventName = "GetSeverity"
)

type GetSeverityEvent struct {
	Filter *entity.SeverityFilter
	Result *entity.Severity
}

func (e *GetSeverityEvent) Name() event.EventName {
	return GetSeverityEventName
}
