package mariadb

import (
	"database/sql"
	"fmt"
	"github.com/samber/lo"
	"reflect"
	"strings"
	"time"
)

const MAX_RESCURSION_DEPTH = 2

func recursiveMarshalCursor(v interface{}, depth int) string {
	// prevent endless recursion by limiting the depth
	if depth > MAX_RESCURSION_DEPTH {
		return ""
	}

	var cursorParts []string

	// get the value and type
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Check if the field is exported, if not we can not access it and skip it
		if !field.IsExported() {
			continue
		}

		if fieldValue.Kind() == reflect.Ptr {
			// check if we got a nil pointer
			if fieldValue.IsNil() {
				continue
			}
			//otherwise derefernce
			fieldValue = fieldValue.Elem()
		}

		// If field has a cursor tag then add it to the cursor
		cursorTag := field.Tag.Get("cursor")
		if cursorTag != "" {
			var value string
			switch val := fieldValue.Interface().(type) {
			case sql.NullInt64:
				if val.Valid {
					value = fmt.Sprintf("%d", val.Int64)
				}
			case sql.NullString:
				if val.Valid {
					value = val.String
				}
			case sql.NullTime:
				if val.Valid {
					value = val.Time.Format(time.RFC3339)
				}
			case sql.NullBool:
				if val.Valid {
					value = fmt.Sprintf("%t", val.Bool)
				}
				// other types are not used and not properly handled currently so we skip them omiting the field
			}
			if value != "" {
				cursorParts = append(cursorParts, fmt.Sprintf("%s=%s", cursorTag, value))
			}

			//otherwise check if the field is a struct to go into recursion for composite struct support
		} else if fieldValue.Kind() == reflect.Struct {
			recursivePart := recursiveMarshalCursor(fieldValue.Interface(), depth+1)
			if recursivePart != "" {
				cursorParts = append(cursorParts, recursivePart)
			}
		}
	}
	return strings.Join(cursorParts, ", ")
}

func MarshalCursor(v interface{}) string {
	return recursiveMarshalCursor(v, 0)
}

func recursiveUnmarshalCursor(cursor string, v interface{}, depth int) error {
	if depth > MAX_RESCURSION_DEPTH {
		return nil
	}
	// get the value and type
	val := reflect.ValueOf(v).Elem()
	typ := reflect.TypeOf(v).Elem()

	// split the cursor string into parts
	// @todo may move outside of the rescursive function to prevent splitting the string multiple times
	cursorParts := strings.Split(cursor, ", ")

	// iterate over the parts
	for _, part := range cursorParts {
		kv := strings.Split(part, "=")

		// check if the part is valid and containing a key and value side
		if len(kv) != 2 {
			continue
		}
		key, value := kv[0], kv[1]

		// iterate over the fields of the struct
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			cursorTag := field.Tag.Get("cursor")

			// get the enum tag if it exists and extract enum values
			enum_tag := field.Tag.Get("cursor_enum")
			enum := make([]string, 0)
			if enum_tag != "" {
				enum = strings.Split(enum_tag, ",")
			}

			fieldValue := val.Field(i)

			if fieldValue.Kind() == reflect.Ptr {
				// check if we got a nil pointer
				if fieldValue.IsNil() {
					continue
				}
				//otherwise derefernce
				fieldValue = fieldValue.Elem()
			}

			if cursorTag == key {
				// validate against enum and empty value
				isInValid := value == "" || (len(enum) > 0 && !lo.Contains(enum, value))

				switch fieldValue.Interface().(type) {
				case sql.NullInt64:
					if isInValid {
						fieldValue.Set(reflect.ValueOf(sql.NullInt64{Int64: 0, Valid: false}))
					} else {
						var intValue int64
						fmt.Sscanf(value, "%d", &intValue)
						fieldValue.Set(reflect.ValueOf(sql.NullInt64{Int64: intValue, Valid: true}))
					}
				case sql.NullString:
					if isInValid {
						fieldValue.Set(reflect.ValueOf(sql.NullString{String: "", Valid: false}))
					} else {
						fieldValue.Set(reflect.ValueOf(sql.NullString{String: value, Valid: true}))
					}
				case sql.NullTime:
					if isInValid {
						fieldValue.Set(reflect.ValueOf(sql.NullTime{Time: time.Time{}, Valid: false}))
					} else {
						t, err := time.Parse(time.RFC3339, value)
						if err != nil {
							fieldValue.Set(reflect.ValueOf(sql.NullTime{Time: time.Time{}, Valid: false}))
						} else {
							fieldValue.Set(reflect.ValueOf(sql.NullTime{Time: t, Valid: true}))
						}
					}
					// other types are currently not used and not properly handled, we simply skip them omiting the field
				}
			} else if fieldValue.Kind() == reflect.Struct {
				// if the field is a struct go into recursion with the reference of the substruct
				// this currently doess process the WHOLE cursor again
				recursiveUnmarshalCursor(cursor, fieldValue.Addr().Interface(), depth+1)
			}
		}
	}
	return nil
}

func UnmarshalCursor(cursor string, v interface{}) error {
	return recursiveUnmarshalCursor(cursor, v, 0)
}
