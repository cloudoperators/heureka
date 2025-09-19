package common

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/openfga"
)

type HandlerContext struct {
	DB       database.Database
	EventReg event.EventRegistry
	Cache    cache.Cache
	Authz    openfga.Authorization
}
