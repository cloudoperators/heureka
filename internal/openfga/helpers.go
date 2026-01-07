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
