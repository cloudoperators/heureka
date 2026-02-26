// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
)

type DbObject struct {
	Properties       []PropertySpec
	FilterProperties []FilterPropertySpec
}

func (do *DbObject) InsertQuery(insertTable string) string {
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		insertTable,
		strings.Join(lo.Map(do.Properties, func(p PropertySpec, _ int) string { return p.GetName() }), ","),
		strings.Join(lo.Map(do.Properties, func(p PropertySpec, _ int) string { return ":" + p.GetName() }), ","))
}

func (do *DbObject) GetUpdateFields(f any) string {
	fl := []string{}
	for _, v := range do.Properties {
		updateField := v.GetUpdateExpression(f)
		if updateField != "" {
			fl = append(fl, updateField)
		}
	}
	return strings.Join(fl, ", ")
}

func (do *DbObject) GetFilterQuery(filter any) string {
	var fl []string
	for _, v := range do.FilterProperties {
		fl = append(fl, v.GetQuery(filter))
	}
	return combineFilterQueries(fl, OP_AND)
}

func (do *DbObject) GetFilterParameters(filter entity.HasPagination, withCursor bool, cursorFields []Field) []any {
	var filterParameters []interface{}
	for _, v := range do.FilterProperties {
		filterParameters = v.AppendParameters(filterParameters, filter)
	}
	if withCursor {
		paginatedX := filter.GetPaginatedX()
		filterParameters = append(filterParameters, GetCursorQueryParameters(paginatedX.First, cursorFields)...)
	}
	return filterParameters
}

type PropertySpec interface {
	GetName() string
	GetUpdateExpression(any) string
}

type Property struct {
	Name            string
	IsUpdatePresent func(any) bool
}

func (p Property) GetName() string {
	return p.Name
}

func (p Property) GetUpdateExpression(f any) string {
	if p.IsUpdatePresent != nil && p.IsUpdatePresent(f) {
		return fmt.Sprintf("%s = :%s", p.Name, p.Name)
	}
	return ""
}

type FilterPropertySpec interface {
	AppendParameters([]any, any) []any
	GetQuery(any) string
}

type FilterProperty struct {
	Query string
	Param func(any) []any
}

func (fp FilterProperty) AppendParameters(params []any, filter any) []any {
	return buildQueryParameters(params, fp.Param(filter))
}

func (fp FilterProperty) GetQuery(item any) string {
	return buildFilterQuery(fp.Param(item), fp.Query, OP_OR)
}

type StateFilterProperty struct {
	Prefix string
	Param  func(any) []entity.StateFilterType
}

func (sfp StateFilterProperty) AppendParameters(params []any, filter any) []any {
	return params // There is no parameter needed for State so let's not append any
}

func (sfp StateFilterProperty) GetQuery(item any) string {
	// State query has to be modified according to parameter
	return buildStateFilterQuery(sfp.Param(item), sfp.Prefix)
}

type JsonFilterProperty struct {
	Query string
	Param func(any) []*entity.Json
}

func (jfp JsonFilterProperty) AppendParameters(params []any, filter any) []any {
	return buildJsonQueryParameters(params, jfp.Param(filter))
}

func (jfp JsonFilterProperty) GetQuery(item any) string {
	return buildJsonFilterQuery(jfp.Param(item), jfp.Query, OP_OR)
}

// DB helpers
func EnsurePagination[T entity.HasPagination](filter T) T {
	var first = 1000
	var after = ""

	px := filter.GetPaginatedX()

	if px.First == nil {
		px.First = &first
	}
	if px.After == nil {
		px.After = &after
	}

	return filter
}

// Helpers

// WrapChecker turns a type-specific check into a generic check
func WrapChecker[T any](check func(T) bool) func(any) bool {
	return func(val any) bool {
		if typedVal, ok := val.(T); ok {
			return check(typedVal)
		}
		return false
	}
}

// WrapRetSlice turns a type-specific accessor into a generic one
func WrapRetSlice[T any, E any](fn func(T) []E) func(any) []any {
	return func(input any) []any {
		val, ok := input.(T)
		if !ok {
			return nil
		}

		res := fn(val)

		out := make([]any, len(res))
		for i := range res {
			out[i] = res[i]
		}
		return out
	}
}

// WrapRetState turns a type-specific accessor into a generic one for StateFilter slice
func WrapRetState[T any](fn func(T) []entity.StateFilterType) func(any) []entity.StateFilterType {
	return func(input any) []entity.StateFilterType {
		val, ok := input.(T)
		if !ok {
			return nil
		}

		res := fn(val)

		out := make([]entity.StateFilterType, len(res))
		for i := range res {
			out[i] = res[i]
		}
		return out
	}
}

// WrapRetJson turns a type-specific accessor into a generic one for Json slice
func WrapRetJson[T any](fn func(T) []*entity.Json) func(any) []*entity.Json {
	return func(input any) []*entity.Json {
		val, ok := input.(T)
		if !ok {
			return nil
		}

		res := fn(val)

		out := make([]*entity.Json, len(res))
		for i := range res {
			out[i] = res[i]
		}
		return out
	}
}

func ToAnySlice[T any](in []T) []any {
	out := make([]any, len(in))
	for i := range in {
		out[i] = in[i]
	}
	return out
}
