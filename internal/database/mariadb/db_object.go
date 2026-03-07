// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

// DbObject
// TODO: type DbObject[ET EntityType, RT RowType] struct {
// TODO: implement Update
// TODO: create Entity and Row interface, extract ToEntity ToRow, reuse
type DbObject struct {
	Prefix           string
	TableName        string
	Properties       []*Property
	FilterProperties []*FilterProperty
}

func (do *DbObject) InsertQuery() string {
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		do.TableName,
		strings.Join(lo.Map(do.Properties, func(p *Property, _ int) string { return p.GetName() }), ","),
		strings.Join(lo.Map(do.Properties, func(p *Property, _ int) string { return ":" + p.GetName() }), ","))
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
		paginatedX := filter.GetPaginated()
		filterParameters = append(filterParameters, GetCursorQueryParameters(paginatedX.First, cursorFields)...)
	}
	return filterParameters
}

func (do *DbObject) Delete(db Db, id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": fmt.Sprintf("database.Delete%s", do.TableName),
	})
	query := fmt.Sprintf(
		"UPDATE %s SET %s_deleted_at = NOW(), %s_updated_by = :userId WHERE %s_id = :id",
		do.TableName,
		do.Prefix,
		do.Prefix,
		do.Prefix)

	args := map[string]interface{}{
		"userId": userId,
		"id":     id,
	}

	stmt, err := db.PrepareNamed(query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
			}).Error(msg)
		return fmt.Errorf("%s", msg)
	}

	defer stmt.Close()
	_, err = stmt.Exec(args)
	if err != nil {
		msg := err.Error()
		l.WithFields(
			logrus.Fields{
				"error": err,
			}).Error(msg)
		return fmt.Errorf("%s", msg)
	}
	return nil
}

// Property
func NewProperty(name string, isUpdatePresent func(any) bool) *Property {
	return &Property{Name: name, IsUpdatePresent: isUpdatePresent}
}

func NewImmutableProperty(name string) *Property {
	return &Property{Name: name}
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

// FilterProperty
type FilterProperty struct {
	QueryBuilder  func([]any) string
	Param         func(any) []any
	ParamAppender func([]any, any) []any
}

func (fp FilterProperty) AppendParameters(params []any, filter any) []any {
	return fp.ParamAppender(params, filter)
}

func (fp FilterProperty) GetQuery(filter any) string {
	return fp.QueryBuilder(fp.Param(filter))
}

func doNotAppendParameters(params []any, _ any) []any {
	return params
}

func NewFilterProperty(query string, param func(any) []any) *FilterProperty {
	return NewNFilterProperty(query, param, 1)
}

func NewNFilterProperty(query string, param func(any) []any, nparam int) *FilterProperty {
	return &FilterProperty{
		QueryBuilder:  func(filter []any) string { return buildFilterQuery(filter, query, OP_OR) },
		Param:         param,
		ParamAppender: func(params []any, filter any) []any { return buildQueryParametersCount(params, param(filter), nparam) },
	}
}

func NewStateFilterProperty(prefix string, param func(any) []entity.StateFilterType) *FilterProperty {
	return &FilterProperty{
		QueryBuilder:  func(state []any) string { return buildStateFilterQuery(ToStateSlice(state), prefix) },
		Param:         WrapRetSlice(param),
		ParamAppender: doNotAppendParameters,
	}
}

func NewJsonFilterProperty(query string, param func(any) []*entity.Json) *FilterProperty {
	return &FilterProperty{
		QueryBuilder:  func(json []any) string { return buildJsonFilterQuery(ToJsonSlice(json), query, OP_OR) },
		Param:         WrapRetSlice(param),
		ParamAppender: func(params []any, filter any) []any { return buildJsonQueryParameters(params, param(filter)) },
	}
}

func NewCustomFilterProperty(queryBuilder func([]any) string, param func(any) []any) *FilterProperty {
	return &FilterProperty{
		QueryBuilder:  queryBuilder,
		Param:         param,
		ParamAppender: doNotAppendParameters,
	}
}

// DB helpers
func EnsurePagination[T entity.HasPagination](filter T) T {
	var first = 1000
	var after = ""

	px := filter.GetPaginated()

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
		typedVal, ok := val.(T)
		if !ok {
			panic(fmt.Sprintf("WrapChecker: expected %T but got %T", *new(T), val))
		}
		return check(typedVal)
	}
}

// WrapBuilder turns a type-specific builder function into a generic builder function
func WrapBuilder[T any](build func([]T) string) func([]any) string {
	return func(values []any) string {
		typed := make([]T, len(values))

		for i, v := range values {
			tv, ok := v.(T)
			if !ok {
				panic(fmt.Sprintf(
					"WrapBuilderSlice: expected %T but got %T",
					*new(T), v,
				))
			}
			typed[i] = tv
		}

		return build(typed)
	}
}

// WrapRetSlice turns a type-specific accessor into a generic one
func WrapRetSlice[T any, E any](fn func(T) []E) func(any) []any {
	return func(input any) []any {
		val, ok := input.(T)
		if !ok {
			panic(fmt.Sprintf("WrapRetSlice: expected %T but got %T", *new(T), input))
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
			panic(fmt.Sprintf("WrapRetState: expected %T but got %T", *new(T), input))
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
			panic(fmt.Sprintf("WrapRetJson: expected %T but got %T", *new(T), input))
		}

		res := fn(val)

		out := make([]*entity.Json, len(res))
		for i := range res {
			out[i] = res[i]
		}
		return out
	}
}

func ToStateSlice(in []any) []entity.StateFilterType { //REMOVE TEMPLATE?
	out := make([]entity.StateFilterType, len(in))
	for i := range in {
		s, ok := in[i].(entity.StateFilterType)
		if !ok {
			panic(fmt.Sprintf("ToStateSlice: expected %T but got %T", new(entity.StateFilterType), in[i]))
		}
		out[i] = s
	}
	return out
}

func ToJsonSlice(in []any) []*entity.Json { //REMOVE TEMPLATE?
	out := make([]*entity.Json, len(in))
	for i := range in {
		s, ok := in[i].(*entity.Json)
		if !ok {
			panic(fmt.Sprintf("ToJsonSlice: expected %T but got %T", new(*entity.Json), in[i]))
		}
		out[i] = s
	}
	return out
}
