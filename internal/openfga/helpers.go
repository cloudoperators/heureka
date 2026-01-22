package openfga

import "strconv"

// ObjectIdFromInt converts a numeric ID to an OpenFGA ObjectId.
func ObjectIdFromInt(id int64) ObjectId {
	return ObjectId(strconv.FormatInt(id, 10))
}

// UserIdFromInt converts an int ID to an OpenFGA UserId.
func UserIdFromInt(id int64) UserId {
	return UserId(strconv.FormatInt(id, 10))
}

// AddRelations accepts a slice of RelationInput and adds them
func AddRelations(authz Authorization, relations []RelationInput) error {
	for _, r := range relations {
		err := authz.AddRelation(r)
		if err != nil {
			return err
		}
	}
	return nil
}

// matchesFilter checks if the given userParts and objectParts match the filters specified in RelationInput.
func matchesFilter(userParts, objectParts []string, r RelationInput, relation string) bool {
	if r.UserType != "" && (len(userParts) < 1 || userParts[0] != string(r.UserType)) {
		return false
	}
	if r.UserId != "" && (len(userParts) < 2 || userParts[1] != string(r.UserId)) {
		return false
	}
	if r.Relation != "" && relation != string(r.Relation) {
		return false
	}
	if r.ObjectType != "" && (len(objectParts) < 1 || objectParts[0] != string(r.ObjectType)) {
		return false
	}
	if r.ObjectId != "" && (len(objectParts) < 2 || objectParts[1] != string(r.ObjectId)) {
		return false
	}
	return true
}
