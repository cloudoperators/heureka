package openfga

import (
	"io/ioutil"

	"github.com/cloudoperators/heureka/internal/util"
	"github.com/sirupsen/logrus"
)

type UserType string
type UserId string
type ObjectType string
type RelationType string
type ObjectId string

type PermissionInput struct {
	UserType   UserType
	UserId     UserId
	Relation   RelationType
	ObjectType ObjectType
	ObjectId   string
}

type RelationInput struct {
	UserType   UserType
	UserId     UserId
	Relation   RelationType
	ObjectType ObjectType
	ObjectId   string
}

type AccessibleResource struct {
	ObjectType ObjectType
	ObjectId   ObjectId
}

type Authorization interface {
	// check if userId has permission on resourceId
	CheckPermission(p PermissionInput) (bool, error)
	// add relationship between userId and resourceId
	AddRelation(r RelationInput) error
	// remove relationship between userId and resourceId
	RemoveRelation(r RelationInput) error
	// ListAccessibleResources returns a list of resource Ids that the user can access.
	ListAccessibleResources(p PermissionInput) ([]AccessibleResource, error)
}

func NewAuthorizationHandler(cfg *util.Config, enablelog bool) Authorization {
	l := newLogger(enablelog)

	if cfg.AuthzOpenFGApiUrl != "" {
		return NewAuthz(l, cfg)
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
