// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"fmt"
	"sort"
)

func ConvertStrToJsonNoError(jsonStr *string) *map[string]any {
	if jsonStr == nil {
		return nil
	}

	var jsonResult *map[string]any

	err := json.Unmarshal([]byte(*jsonStr), &jsonResult)
	if err != nil {
		return nil
	}

	return jsonResult
}

func ConvertJsonToStrNoError(jsonVar *map[string]any) string {
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

func ConvertJsonPointerToValue(jsonVar *map[string]any) map[string]any {
	if jsonVar == nil {
		return map[string]any{}
	}

	return *jsonVar
}

type JsonAttribute struct {
	Key  string
	Attr any
}

func SeparateJsonAttributes(jsonVar map[string]any) []JsonAttribute {
	return separateJsonAttributesRecursive(jsonVar, "")
}

func separateJsonAttributesRecursive(jsonVar any, prefix string) []JsonAttribute {
	attributes := []JsonAttribute{}

	switch v := jsonVar.(type) {
	case map[string]any:
		var mapKeys []string
		for k := range v {
			mapKeys = append(mapKeys, k)
		}

		sort.Strings(mapKeys)

		for _, k := range mapKeys {
			fullKey := k
			if prefix != "" {
				fullKey = fmt.Sprintf("%s.%s", prefix, k)
			}

			attributes = append(attributes, separateJsonAttributesRecursive(v[k], fullKey)...)
		}
	case []any:
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
