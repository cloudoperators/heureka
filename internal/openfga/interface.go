package openfga

import (
	"io/ioutil"

	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/sirupsen/logrus"
)

type Authorization interface {
	EventInjection(db database.Database, e event.Event)
	// check if userId has permission on resourceId
	CheckPermission(userId, resourceId, permission string) (bool, error)
	// add relationship between userId and resourceId
	AddRelation(userId, resourceId string) error
	// remove relationship between userId and resourceId
	RemoveRelation(userId, resourceId string) error
	// ListAccessibleResources returns a list of resource Ids that the user can access.
	ListAccessibleResources(userID string, resourceType string, permission string, relation string) ([]string, error)
}

func NewAuthorizationHandler(cfg *util.Config, enablelog bool) Authorization {
	l := newLogger(enablelog)

	switch cfg.AuthzEnabled {
	case true:
		return NewAuthz(l, cfg)
	case false:
		return NewNoAuthz(cfg)
	}
	return NewNoAuthz(cfg)
}

func newLogger(enableLog bool) *logrus.Logger {
	l := logrus.New()
	if !enableLog {
		l.SetOutput(ioutil.Discard)
	}
	return l
}
