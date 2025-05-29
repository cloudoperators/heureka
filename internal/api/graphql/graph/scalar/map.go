// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scalar

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/cloudoperators/heureka/internal/util"
)

type Map map[string]any

func (m Map) MarshalGQL(w io.Writer) {
	if m == nil {
		w.Write([]byte("null"))
		return
	}

	data, err := json.Marshal(map[string]any(m))
	if err != nil {
		w.Write([]byte("null"))
		return
	}

	w.Write(data)
}

func (m *Map) UnmarshalGQL(v any) error {
	fmt.Println(fmt.Sprintf("DEBUG: UnmarshalGQL received value: %+v, type: %T", v, v))
	if v == nil {
		*m = nil
		return nil
	}

	switch val := v.(type) {
	case map[string]any:
		*m = Map(val)
		return nil
	case string:
		mapVal := util.ConvertStrToJsonNoError(&val)
		*m = Map(*mapVal)
		if mapVal == nil {
			return fmt.Errorf("cannot unmarshal %T into Map", v)
		}
		return nil
	default:
		return fmt.Errorf("cannot unmarshal %T into Map", v)
	}
}

func MarshalMap(val map[string]any) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if val == nil {
			w.Write([]byte("null"))
			return
		}

		data, err := json.Marshal(val)
		if err != nil {
			panic(fmt.Errorf("failed to marshal map: %w", err))
		}

		w.Write(data)
	})
}

func UnmarshalMap(v any) (map[string]any, error) {
	fmt.Println(fmt.Sprintf("DEBUG2: UnmarshalGQL received value: %+v, type: %T", v, v))
	if v == nil {
		return nil, nil
	}

	switch val := v.(type) {
	case map[string]any:
		return val, nil
	case string:
		mapVal := util.ConvertStrToJsonNoError(&val)
		if mapVal == nil {
			return nil, fmt.Errorf("cannot unmarshal %T into Map", v)
		}
		return *mapVal, nil
	default:
		return nil, fmt.Errorf("cannot unmarshal %T into map[string]any", v)
	}
}
