package util

import (
	"encoding/json"
	"fmt"
	"sort"
)

func ConvertStrToJsonNoError(jsonStr *string) *map[string]interface{} {
	if jsonStr == nil {
		return nil
	}
	var jsonResult *map[string]interface{}
	err := json.Unmarshal([]byte(*jsonStr), &jsonResult)
	if err != nil {
		return nil
	}
	return jsonResult
}

func ConvertJsonToStrNoError(jsonVar *map[string]interface{}) string {
	if jsonVar == nil {
		return ""
	}
	jsonBytes, err := json.Marshal(jsonVar)
	if err != nil {
		return ""
	}
	jsonStr := string(jsonBytes)
	return jsonStr
}

func ConvertJsonPointerToValue(jsonVar *map[string]interface{}) map[string]interface{} {
	if jsonVar == nil {
		return map[string]interface{}{}
	}
	return *jsonVar
}

type JsonAttribute struct {
	Key  string
	Attr interface{}
}

func SeparateJsonAttributes(jsonVar map[string]interface{}) []JsonAttribute {
	return separateJsonAttributesRecursive(jsonVar, "")
}

func separateJsonAttributesRecursive(jsonVar interface{}, prefix string) []JsonAttribute {
	attributes := []JsonAttribute{}

	switch v := jsonVar.(type) {
	case map[string]interface{}:
		var mapKeys []string
		for k := range v {
			mapKeys = append(mapKeys, k)
		}
		sort.Strings(mapKeys)

		for _, k := range mapKeys {
			var fullKey string
			if prefix == "" {
				fullKey = k
			} else {
				fullKey = fmt.Sprintf("%s.%s", prefix, k)
			}
			attributes = append(attributes, separateJsonAttributesRecursive(v[k], fullKey)...)
		}
	case []interface{}:
		for i, item := range v {
			fullKey := fmt.Sprintf("%s[%d]", prefix, i)
			attributes = append(attributes, separateJsonAttributesRecursive(item, fullKey)...)
		}
	default:
		attributes = append(attributes, JsonAttribute{
			Key:  prefix,
			Attr: v,
		})
	}

	return attributes
}
