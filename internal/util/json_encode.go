package util

import "encoding/json"

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
