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

type Json map[string]any

func (m Json) MarshalGQL(w io.Writer) {
	if m == nil {
		if _, err := w.Write([]byte("null")); err != nil {
			panic(fmt.Errorf("failed to write null value: %w", err))
		}

		return
	}

	data, err := json.Marshal(map[string]any(m))
	if err != nil {
		if _, err := w.Write([]byte("null")); err != nil {
			panic(fmt.Errorf("failed to write null value: %w", err))
		}

		return
	}

	if _, err := w.Write(data); err != nil {
		panic(fmt.Errorf("failed to write data: %w", err))
	}
}

func (m *Json) UnmarshalGQL(v any) error {
	if v == nil {
		*m = nil
		return nil
	}

	switch val := v.(type) {
	case map[string]any:
		*m = Json(val)
		return nil
	case string:
		jsonVal := util.ConvertStrToJsonNoError(&val)
		if jsonVal == nil {
			return fmt.Errorf("cannot unmarshal %T into Json", v)
		}

		*m = Json(*jsonVal)

		return nil
	default:
		return fmt.Errorf("cannot unmarshal %T into Json", v)
	}
}

func MarshalJson(val map[string]any) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if val == nil {
			if _, err := w.Write([]byte("null")); err != nil {
				panic(fmt.Errorf("failed to write null value: %w", err))
			}

			return
		}

		data, err := json.Marshal(val)
		if err != nil {
			panic(fmt.Errorf("failed to marshal json: %w", err))
		}

		if _, err := w.Write(data); err != nil {
			panic(fmt.Errorf("failed to write data: %w", err))
		}
	})
}

func UnmarshalJson(v any) (map[string]any, error) {
	if v == nil {
		return nil, nil
	}

	switch val := v.(type) {
	case map[string]any:
		return val, nil
	case string:
		jsonVal := util.ConvertStrToJsonNoError(&val)
		if jsonVal == nil {
			return nil, fmt.Errorf("cannot unmarshal %T into Map", v)
		}
		return *jsonVal, nil
	default:
		return nil, fmt.Errorf("cannot unmarshal %T into map[string]any", v)
	}
}
