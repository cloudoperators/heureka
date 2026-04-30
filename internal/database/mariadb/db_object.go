// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	sq "github.com/Masterminds/squirrel"
)

// DbObject
type DbObject[ET entity.Entity] struct {
	Prefix           string
	TableName        string
	Properties       []*Property
	FilterProperties []*FilterProperty
	JoinDefs         []*JoinDef
}

func (do *DbObject[ET]) InsertQuery(entityItem ET) (string, []any, error) {
	columns := lo.Map(do.Properties, func(p *Property, _ int) string {
		return p.GetName()
	})

	values := lo.Map(do.Properties, func(p *Property, _ int) any {
		return p.GetValue(entityItem)
	})

	qb := sq.
		Insert(do.TableName).
		Columns(columns...).
		Values(values...)

	return qb.ToSql()
}

func (do *DbObject[ET]) GetFilterQuery(filter any) string {
	var fl []string
	for _, v := range do.FilterProperties {
		fl = append(fl, v.GetQuery(filter))
	}

	return combineFilterQueries(fl, OP_AND)
}

func (do *DbObject[ET]) GetFilterParameters(
	filter entity.HasPagination,
	withCursor bool,
	cursorFields []Field,
) []any {
	var filterParameters []any
	for _, v := range do.FilterProperties {
		filterParameters = v.AppendParameters(filterParameters, filter)
	}

	if withCursor {
		paginated := filter.GetPaginated()
		filterParameters = append(
			filterParameters,
			GetCursorQueryParameters(paginated.First, cursorFields)...)
	}

	return filterParameters
}

func (do *DbObject[ET]) Create(db Db, entityItem ET) (ET, error) {
	var zero ET

	l := logrus.WithFields(logrus.Fields{
		do.Prefix: entityItem,
		"event":   fmt.Sprintf("database.Create%s", do.TableName),
	})

	sqlQuery, args, err := do.InsertQuery(entityItem)
	if err != nil {
		return zero, err
	}

	id, err := PerformInsertArgs(db, sqlQuery, args, l)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Error 1062") {
			return zero, database.NewDuplicateEntryDatabaseError(
				fmt.Sprintf("%s element already exists", do.TableName),
			)
		}

		return zero, err
	}

	entityItem.SetId(id)

	return entityItem, nil
}

func (do *DbObject[ET]) Update(db Db, entityItem ET) error {
	l := logrus.WithFields(logrus.Fields{
		do.Prefix: entityItem,
		"event":   fmt.Sprintf("database.Update%s", do.TableName),
	})

	updateValues := do.getUpdateMap(entityItem)
	qb := sq.
		Update(do.TableName).
		SetMap(updateValues).
		Where(sq.Eq{fmt.Sprintf("%s_id", do.Prefix): entityItem.GetId()})

	sqlQuery, args, err := qb.ToSql()
	if err != nil {
		return err
	}

	_, err = PerformExecArgs(db, sqlQuery, args, l)

	return err
}

func (do *DbObject[ET]) getUpdateMap(f any) map[string]any {
	m := make(map[string]any)

	for _, v := range do.Properties {
		val, isUpdatePresent := v.GetUpdateData(f)
		if isUpdatePresent {
			m[v.GetName()] = val
		}
	}

	return m
}

func (do *DbObject[ET]) Delete(db Db, id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": fmt.Sprintf("database.Delete%s", do.TableName),
	})

	deletedAtColumn := do.Prefix + "_deleted_at"
	updatedByColumn := do.Prefix + "_updated_by"
	idColumn := do.Prefix + "_id"

	qb := sq.
		Update(do.TableName).
		Set(deletedAtColumn, sq.Expr("NOW()")).
		Set(updatedByColumn, userId).
		Where(sq.Eq{idColumn: id})

	sqlQuery, args, err := qb.ToSql()
	if err != nil {
		return err
	}

	_, err = PerformExecArgs(db, sqlQuery, args, l)

	return err
}

func (do *DbObject[ET]) GetJoins(filter any, order *Order) string {
	return NewJoinResolver(do.JoinDefs).Build(filter, order)
}

func (do *DbObject[ET]) GetFilterWhereClause(filter any, withCursor bool) (string, bool) {
	filterStr := do.GetFilterQuery(filter)
	hasFilter := filterStr != ""
	if hasFilter || withCursor {
		return fmt.Sprintf("WHERE %s", filterStr), hasFilter
	}

	return "", false
}

func (do *DbObject[ET]) GetCursorQuery(hasFilter *bool, cursorFields []Field, withCursor *bool, aggregated bool) string {
	cursorQuery := CreateCursorQuery("", cursorFields)

	if aggregated {
		if (withCursor == nil || *withCursor) && (hasFilter == nil || *hasFilter) && cursorQuery != "" {
			cursorQuery = fmt.Sprintf("HAVING (%s)", cursorQuery)
		}
	} else {
		if hasFilter != nil {
			if *hasFilter && *withCursor && cursorQuery != "" {
				cursorQuery = fmt.Sprintf(" AND (%s)", cursorQuery)
			}
		} else {
			panic("hasFilter for not aggregated query has to be passed (has to be not nil).")
		}
	}

	return cursorQuery
}

// Property
const NoUpdate = false

func NewProperty(name string, access func(any) (any, bool)) *Property {
	return &Property{Name: name, Access: access}
}

type Property struct {
	Name   string
	Access func(any) (any, bool)
}

func (p Property) GetName() string {
	return p.Name
}

func (p Property) GetValue(f any) any {
	val, _ := p.Access(f)
	return val
}

func (p Property) GetUpdateData(f any) (any, bool) {
	return p.Access(f)
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

func NewStateFilterProperty(
	prefix string,
	param func(any) []entity.StateFilterType,
) *FilterProperty {
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

func NewCustomFilterProperty(
	queryBuilder func([]any) string,
	param func(any) []any,
) *FilterProperty {
	return &FilterProperty{
		QueryBuilder:  queryBuilder,
		Param:         param,
		ParamAppender: doNotAppendParameters,
	}
}

// Join
type JoinType string

const (
	LeftJoin  JoinType = "LEFT JOIN"
	RightJoin JoinType = "RIGHT JOIN"
	InnerJoin JoinType = "JOIN"
)

func DependentJoin(any, *Order) bool { return false }

type JoinDef struct {
	Name      string
	Type      JoinType
	Table     string
	On        string
	DependsOn []string
	Condition func(any, *Order) bool
}

type JoinResolver struct {
	defs     []*JoinDef
	included map[string]bool
	order    []string
}

func NewJoinResolver(defs []*JoinDef) *JoinResolver {
	r := &JoinResolver{
		defs:     defs,
		included: map[string]bool{},
	}

	return r
}

func (jr *JoinResolver) require(name string) {
	if jr.included[name] {
		return
	}

	def, ok := lo.Find(jr.defs, func(jd *JoinDef) bool {
		return jd.Name == name
	})
	if !ok {
		panic("JoinResolver::require(...) Unknown join: " + name)
	}

	// resolve dependencies first
	for _, dep := range def.DependsOn {
		jr.require(dep)
	}

	jr.included[name] = true
	jr.order = append(jr.order, name)
}

func (jr *JoinResolver) Build(filter any, order *Order) string {
	for _, def := range jr.defs {
		if def.Condition != nil && def.Condition(filter, order) {
			jr.require(def.Name)
		}
	}

	var result []string

	// This is little tricky part, but we need to deal with that this way
	// until we have stateful join pattern which is created for issue.go
	// with non-uniq tablename 'IM IssueMatch' which join operation
	// depends on filter pattern with precedence for some members (there
	// is if...else if which cannot be replaced by if... and if... what
	// is a mess and misconception
	uniqTableName := make(map[string]struct{})

	for _, name := range jr.order {
		j, ok := lo.Find(jr.defs, func(jd *JoinDef) bool {
			return jd.Name == name
		})
		if !ok {
			panic("JoinResolver::Build(...) Unknown join: " + name)
		}

		if _, ok := uniqTableName[j.Table]; ok {
			continue
		}

		uniqTableName[j.Table] = struct{}{}

		joinSQL := fmt.Sprintf("%s %s ON %s",
			j.Type,
			j.Table,
			j.On,
		)

		result = append(result, joinSQL)
	}

	return strings.Join(result, "\n")
}

// DB helpers
func EnsurePagination[T entity.HasPagination](filter T) T {
	first := 1000
	after := ""

	px := filter.GetPaginated()

	if px.First == nil {
		px.First = &first
	}

	if px.After == nil {
		px.After = &after
	}

	return filter
}

func PerformExecArgs(db Db, query string, args []any, l *logrus.Entry) (sql.Result, error) {
	res, err := db.Exec(query, args...)
	if err != nil {
		msg := err.Error()
		l.WithFields(logrus.Fields{
			"error": err,
			"query": query,
			"args":  args,
		}).Error(msg)

		return nil, fmt.Errorf("%s", msg)
	}

	return res, nil
}

func PerformInsertArgs(db Db, query string, args []any, l *logrus.Entry) (int64, error) {
	res, err := PerformExecArgs(db, query, args, l)
	if err != nil {
		return -1, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		msg := "Error while getting last insert id"
		l.WithFields(logrus.Fields{
			"error": err,
		}).Error(msg)

		return -1, fmt.Errorf("%s", msg)
	}

	l.WithFields(logrus.Fields{
		"id": id,
	}).Debug("Successfully performed insert")

	return id, nil
}

// Helpers

// WrapAccess turns a type-specific data access into a generic data access
func WrapAccess[T any, TRet any](access func(T) (TRet, bool)) func(any) (any, bool) {
	return func(val any) (any, bool) {
		typedVal, ok := val.(T)
		if !ok {
			panic(fmt.Sprintf("WrapAccess: expected %T but got %T", *new(T), val))
		}

		return access(typedVal)
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

		copy(out, res)

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
		copy(out, res)

		return out
	}
}

// WrapJoinCondition turns a type-specific join planner condition using filter and order
func WrapJoinCondition[T any](joinCond func(T, *Order) bool) func(any, *Order) bool {
	return func(filter any, order *Order) bool {
		typedFilter, ok := filter.(T)
		if !ok {
			panic(fmt.Sprintf("WrapJoinCondition: expected %T but got %T", *new(T), filter))
		}

		return joinCond(typedFilter, order)
	}
}

func ToStateSlice(in []any) []entity.StateFilterType {
	out := make([]entity.StateFilterType, len(in))
	for i := range in {
		s, ok := in[i].(entity.StateFilterType)
		if !ok {
			panic(
				fmt.Sprintf(
					"ToStateSlice: expected %T but got %T",
					new(entity.StateFilterType),
					in[i],
				),
			)
		}

		out[i] = s
	}

	return out
}

func ToJsonSlice(in []any) []*entity.Json {
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

func ValueOrDefault[T any](p *T, def T) T {
	if p == nil {
		return def
	}

	return *p
}
