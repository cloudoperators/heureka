// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/go-sql-driver/mysql"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	sq "github.com/Masterminds/squirrel"
)

// DbObject
type DbObject[ET entity.Entity, ETFilter entity.Filter, ETResult entity.HeurekaEntity, ETAgg any] struct {
	Prefix               string
	TableName            string
	TableKey             string
	DefaultOrder         entity.Order
	OrderPrefix          string
	Properties           []*Property[ET]
	FilterProperties     []*FilterProperty[ETFilter]
	JoinDefs             []*JoinDef[ETFilter]
	Attributes           []Attr
	Aggregated           bool
	ExtraColumnsSelector func(ETFilter, *Order) []string
	RowToData            func(RowComposite, []entity.Order) (ET, string)
	RowToAggregatedData  func(RowComposite, []entity.Order) (ET, ETAgg, string)
	NewResult            func(ET, ETAgg, string) ETResult
	ForceFrom            string
}

type Attr struct {
	Name  string
	Order entity.Order
}

// private TODO: Fix testing and unexport
func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) InsertQuery(entityItem ET) (string, []any, error) {
	columns := lo.Map(do.Properties, func(p *Property[ET], _ int) string {
		return p.GetName()
	})

	values := lo.Map(do.Properties, func(p *Property[ET], _ int) any {
		return p.GetValue(entityItem)
	})

	qb := sq.
		Insert(do.TableName).
		Columns(columns...).
		Values(values...)

	return qb.ToSql()
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) Create(db Db, entityItem ET) (ET, error) {
	if do.TableName == "" || do.Prefix == "" {
		panic("database.Create (" + do.TableName + ") - not allowed")
	}

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

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) Update(db Db, entityItem ET) error {
	if do.TableName == "" || do.Prefix == "" {
		panic("database.Create (" + do.TableName + ") - not allowed")
	}

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

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) getUpdateMap(e ET) map[string]any {
	m := make(map[string]any)

	for _, v := range do.Properties {
		val, isUpdatePresent := v.GetUpdateData(e)
		if isUpdatePresent {
			m[v.GetName()] = val
		}
	}

	return m
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) Delete(db Db, id int64, userId int64) error {
	if do.TableName == "" || do.Prefix == "" {
		panic("database.Delete (" + do.TableName + ") - not allowed")
	}

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

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) Count(ctx context.Context, db Db, filter ETFilter) (int64, error) {
	if do.TableName == "" || do.TableKey == "" || do.Prefix == "" {
		panic("database.Count (" + do.TableName + ") - not allowed")
	}

	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  fmt.Sprintf("database.Count (%s)", do.TableName),
	})

	ord := NewOrder([]entity.Order{}, do.DefaultOrder)

	baseQuery := do.countSelectBuilder()
	baseQuery = do.fromBuilder(baseQuery)

	stmt, filterParameters, err := do.BuildStatement(ctx, l, db, baseQuery, filter, ord, false)
	if err != nil {
		return -1, fmt.Errorf("failed to build %s count query: %w", do.TableName, err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error during close stmt: %s", err)
		}
	}()

	count, err := performCountScan(ctx, stmt, filterParameters, l)
	if err != nil {
		return -1, fmt.Errorf("failed to count %s: %w", do.TableName, err)
	}

	return count, nil
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) countSelectBuilder() sq.SelectBuilder {
	return sq.Select(fmt.Sprintf("count(distinct %s)", do.getAttrStr("id")))
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) Get(ctx context.Context, db Db, filter ETFilter, order []entity.Order) ([]ETResult, error) {
	if do.TableName == "" || (do.TableKey != "" && do.Prefix == "") || do.RowToData == nil || do.NewResult == nil {
		panic("database.Get (" + do.TableName + ") - not allowed")
	}

	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  fmt.Sprintf("database.Get (%s)", do.TableName),
	})

	ord := NewOrderWithCounterPrefix(order, do.DefaultOrder, do.OrderPrefix)

	baseQuery := do.getSelectBuilder(filter, ord)
	baseQuery = do.fromBuilder(baseQuery)
	baseQuery = do.groupByBuilder(baseQuery)

	stmt, filterParameters, err := do.BuildStatement(ctx, l, db, baseQuery, filter, ord, true)
	if err != nil {
		return nil, fmt.Errorf("failed to build %s Get query: %w", do.TableName, err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error during close stmt: %s", err)
		}
	}()

	return performListScan(
		ctx,
		stmt,
		filterParameters,
		l,
		func(l []ETResult, e RowComposite) []ETResult {
			result, cursor := do.RowToData(e, order)
			return append(l, do.NewResult(result, *new(ETAgg), cursor))
		},
	)
}

type AggregationDef struct {
	Aggregations  []Aggregation
	From          string
	Joins         []string
	OrderByPrefix string
}

type Aggregation struct {
	Table       string
	TableKey    string
	Columns     []string
	ForcedJoins []string
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) GetWithAggregations(ctx context.Context, db Db, aggDef AggregationDef, filter ETFilter, order []entity.Order) ([]ETResult, error) {
	if do.TableName == "" || (do.TableKey != "" && do.Prefix == "") || do.RowToAggregatedData == nil || do.NewResult == nil {
		panic("database.GetWithAggregations (" + do.TableName + ") - not allowed")
	}

	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  fmt.Sprintf("database.GetWithAggregations (%s)", do.TableName),
	})

	ord := NewOrder(order, do.DefaultOrder)

	baseQuery := do.getSelectBuilder(filter, ord)
	baseQuery = do.fromBuilder(baseQuery)
	baseQuery = do.groupByBuilder(baseQuery)

	qbs, err := lo.MapErr(aggDef.Aggregations, func(agg Aggregation, _ int) (any, error) {
		q := baseQuery.Columns(agg.Columns...)
		return do.BuildQuery(q, filter, ord, true, agg.ForcedJoins)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build query builder: %w", err)
	}

	qb := sq.Select(lo.Map(aggDef.Aggregations, func(agg Aggregation, _ int) string { return agg.TableKey + ".*" })...)

	qb = qb.From(aggDef.From)
	for _, j := range aggDef.Joins {
		qb = qb.JoinClause(j)
	}

	if aggDef.OrderByPrefix != "" {
		qb = qb.OrderBy(ord.ToSqlWithPrefixAll(aggDef.OrderByPrefix))
	}

	qb = qb.Prefix("WITH "+strings.Join(lo.Map(aggDef.Aggregations, func(agg Aggregation, _ int) string { return agg.Table + " AS ( ? )" }), ", "), qbs...)

	query, params, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query using squirrel: %w", err)
	}

	stmt, err := db.PreparexContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare context for get with aggregations: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error during close stmt: %s", err)
		}
	}()

	return performListScan(
		ctx,
		stmt,
		params,
		l,
		func(l []ETResult, e RowComposite) []ETResult {
			return append(l, do.NewResult(do.RowToAggregatedData(e, order)))
		},
	)
}

// TODO: The only difference between Get/GetAllCursors is withCursor(false/true), logging and picked Appender (Extract Method)
func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) GetAllCursors(ctx context.Context, db Db, filter ETFilter, order []entity.Order) ([]string, error) {
	if do.TableName == "" || do.TableKey == "" || do.Prefix == "" || do.RowToData == nil {
		panic("database.GetAllCursors (" + do.TableName + ") - not allowed")
	}

	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  fmt.Sprintf("database.GetAllCursors (%s)", do.TableName),
	})

	ord := NewOrderWithCounterPrefix(order, do.DefaultOrder, do.OrderPrefix)

	baseQuery := do.getSelectBuilder(filter, ord)
	baseQuery = do.fromBuilder(baseQuery)
	baseQuery = do.groupByBuilder(baseQuery)

	stmt, filterParameters, err := do.BuildStatement(ctx, l, db, baseQuery, filter, ord, false)
	if err != nil {
		return nil, fmt.Errorf("failed to build %s GetAllCursors query: %w", do.TableName, err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error during close stmt: %s", err)
		}
	}()

	return performListScan(
		ctx,
		stmt,
		filterParameters,
		l,
		func(l []string, e RowComposite) []string {
			_, cursor := do.RowToData(e, order)
			return append(l, cursor)
		},
	)
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) GetAttr(ctx context.Context, db Db, attrName string, filter ETFilter) ([]string, error) {
	if do.TableName == "" || do.TableKey == "" || do.Prefix == "" {
		panic("database.GetAttr (" + do.TableName + ") - not allowed")
	}

	attr, ok := lo.Find(do.Attributes, func(x Attr) bool {
		return x.Name == attrName
	})
	if !ok {
		panic("database.GetAttr (" + do.TableName + ") - not allowed for: '" + attrName + "'")
	}

	return do.queryAttr(ctx, db, attr, filter)
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) queryAttr(ctx context.Context, db Db, attr Attr, filter ETFilter) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  fmt.Sprintf("database.GetAttr (%s) -> %s", do.TableName, attr.Name),
	})

	ord := NewOrder([]entity.Order{}, attr.Order)

	baseQuery := sq.Select(do.getAttrStr(attr.Name))
	baseQuery = do.fromBuilder(baseQuery)

	stmt, filterParameters, err := do.BuildStatement(ctx, l, db, baseQuery, filter, ord, false)
	if err != nil {
		return nil, fmt.Errorf("failed to build %s queryAttr query: %w", do.TableName, err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error during close stmt: %s", err)
		}
	}()

	// Execute the query
	rows, err := stmt.QueryxContext(ctx, filterParameters...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute %s attribute query for %s: %w", do.TableName, attr.Name, err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			logrus.Warnf("error during close rows: %s", err)
		}
	}()

	// Collect the results
	lVal := []string{}

	for rows.Next() {
		var raw any
		if err := rows.Scan(&raw); err != nil {
			l.Error("Error scanning row: ", err)
			continue
		}

		if raw == nil {
			continue
		}

		switch v := raw.(type) {
		case string:
			lVal = append(lVal, v)
		case []byte:
			lVal = append(lVal, string(v))
		default:
			lVal = append(lVal, fmt.Sprint(v))
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"error iterating %s attribute rows for %s: %w",
			do.TableName,
			attr.Name,
			err,
		)
	}

	return lVal, nil
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) GetAllIds(ctx context.Context, db Db, filter ETFilter) ([]int64, error) {
	if do.TableName == "" || do.TableKey == "" || do.Prefix == "" {
		panic("database.GetAllIds (" + do.TableName + ") - not allowed")
	}

	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  fmt.Sprintf("database.GetAllIds (%s)", do.TableName),
	})

	ord := NewOrder([]entity.Order{}, do.DefaultOrder)

	baseQuery := sq.Select(do.getAttrStr("id"))
	baseQuery = do.fromBuilder(baseQuery)
	baseQuery = do.groupByBuilder(baseQuery)

	stmt, filterParameters, err := do.BuildStatement(ctx, l, db, baseQuery, filter, ord, false)
	if err != nil {
		return nil, fmt.Errorf("failed to build %s GetAllIds query: %w", do.TableName, err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error during close stmt: %s", err)
		}
	}()

	return performIdScan(ctx, stmt, filterParameters, l)
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) BuildStatement(
	ctx context.Context,
	l *logrus.Entry,
	db Db,
	baseQuery sq.SelectBuilder,
	filter ETFilter,
	ord *Order,
	withCursor bool,
) (Stmt, []any, error) {
	qb, err := do.BuildQuery(baseQuery, filter, ord, withCursor, nil)
	if err != nil {
		return nil, nil, err
	}

	query, params, err := qb.ToSql()
	if err != nil {
		return nil, nil, err
	}

	stmt, err := do.PrepareContext(ctx, l, db, query)

	return stmt, params, err
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) BuildQuery(
	baseQuery sq.SelectBuilder,
	filter ETFilter,
	ord *Order,
	withCursor bool,
	forcedJoin []string,
) (sq.SelectBuilder, error) {
	filter = EnsureFilter(filter)
	qb := do.AddJoins(baseQuery, filter, ord, forcedJoin)
	qb = do.AddFilter(qb, filter)

	if withCursor {
		cursorFields, err := DecodeCursor(filter.GetPaginated().After)
		if err != nil {
			return qb, fmt.Errorf("failed to decode cursor: %w", err)
		}

		qb = do.AddCursor(qb, filter, cursorFields)
	}

	qb = qb.OrderBy(ord.ToSql())

	return qb, nil
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) PrepareContext(ctx context.Context, l *logrus.Entry, db Db, query string) (Stmt, error) {
	stmt, err := db.PreparexContext(ctx, query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			},
		).Error(msg)

		return nil, fmt.Errorf("%s", msg)
	}

	return stmt, nil
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) getSelectBuilder(filter ETFilter, ord *Order) sq.SelectBuilder {
	select0 := []string{}
	if do.TableKey != "" {
		select0 = append(select0, fmt.Sprintf("%s.*", do.TableKey))
	}

	return sq.Select(do.selectColumns(select0, filter, ord)...)
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) getAttrStr(attrName string) string {
	return fmt.Sprintf("%s.%s_%s", do.TableKey, do.Prefix, attrName)
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) selectColumns(s0 []string, filter ETFilter, order *Order) []string {
	if do.ExtraColumnsSelector != nil {
		return append(s0, do.ExtraColumnsSelector(filter, order)...)
	}

	return s0
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) fromBuilder(baseQuery sq.SelectBuilder) sq.SelectBuilder {
	if do.ForceFrom != "" {
		return baseQuery.From(do.ForceFrom)
	} else if do.TableKey != "" {
		return baseQuery.From(do.TableName + " " + do.TableKey)
	}

	return baseQuery.From(do.TableName)
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) groupByBuilder(baseQuery sq.SelectBuilder) sq.SelectBuilder {
	if do.TableKey != "" {
		baseQuery = baseQuery.GroupBy(do.getAttrStr("id"))
	}

	return baseQuery
}

// private TODO: fix tests and unexport
func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) AddJoins(qb sq.SelectBuilder, filter ETFilter, order *Order, forcedJoin []string) sq.SelectBuilder {
	joins := NewJoinResolver(do.JoinDefs, forcedJoin).Build(filter, order)
	for _, join := range joins {
		qb = qb.JoinClause(join)
	}

	return qb
}

// private TODO: fix tests and unexport
func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) AddFilter(qb sq.SelectBuilder, filter ETFilter) sq.SelectBuilder {
	for _, v := range do.FilterProperties {
		if q := v.GetQuery(filter); q != "" {
			qb = qb.Where(q, v.GetParameters(filter)...)
		}
	}

	return qb
}

// private TODO: fix tests and unexport
func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) AddCursor(qb sq.SelectBuilder, filter entity.Filter, cursorFields []Field) sq.SelectBuilder {
	paginated := filter.GetPaginated()
	cursorQuery := CreateCursorQuery("", cursorFields)
	cursorParams, limit := do.getCursorQueryParameters(paginated.First, cursorFields)

	if cursorQuery != "" {
		if do.Aggregated {
			qb = qb.Having(cursorQuery, cursorParams...)
		} else {
			qb = qb.Where(cursorQuery, cursorParams...)
		}
	}

	return qb.Limit(uint64(limit))
}

func (do *DbObject[ET, ETFilter, ETResult, ETAgg]) getCursorQueryParameters(pagFirst *int, cursorFields []Field) ([]any, int) {
	var cursorParameters []any

	p := CreateCursorParameters([]any{}, cursorFields)

	cursorParameters = append(cursorParameters, p...)

	var limit int
	if pagFirst == nil {
		limit = 1000
	} else {
		limit = *pagFirst
	}

	return cursorParameters, limit
}

// Property
const NoUpdate = false

func NewProperty[T any](name string, access func(T) (any, bool)) *Property[T] {
	return &Property[T]{Name: name, Access: access}
}

type Property[T any] struct {
	Name   string
	Access func(T) (any, bool)
}

func (p Property[T]) GetName() string {
	return p.Name
}

func (p Property[T]) GetValue(e T) any {
	val, _ := p.Access(e)
	return val
}

func (p Property[T]) GetUpdateData(e T) (any, bool) {
	return p.Access(e)
}

// FilterProperty
type FilterProperty[T any] struct {
	BuildQuery  func([]any) string
	GetParam    func(T) []any
	BuildParams func(T) []any
}

func (fp FilterProperty[T]) GetParameters(filter T) []any {
	return fp.BuildParams(filter)
}

func (fp FilterProperty[T]) GetQuery(filter T) string {
	return fp.BuildQuery(fp.GetParam(filter))
}

func doNotUseParameters[T any](_ T) []any {
	return []any{}
}

func NewFilterProperty[T any](query string, param func(T) any) *FilterProperty[T] {
	return NewNFilterProperty[T](query, param, 1)
}

func NewNFilterProperty[T any](query string, param func(T) any, nparam int) *FilterProperty[T] {
	return &FilterProperty[T]{
		BuildQuery:  func(filterParam []any) string { return buildFilterQuery(filterParam, query, OP_OR) },
		GetParam:    WrapConvertRetSlice[T, any](param),
		BuildParams: func(filter T) []any { return getNParameters(WrapConvertRetSlice[T, any](param)(filter), nparam) },
	}
}

func getNParameters(params []any, count int) []any {
	var nparams []any

	for _, p := range params {
		for range count {
			nparams = append(nparams, p)
		}
	}

	return nparams
}

func NewStateFilterProperty[T any](
	prefix string,
	param func(T) any,
) *FilterProperty[T] {
	return &FilterProperty[T]{
		BuildQuery:  func(state []any) string { return buildStateFilterQuery(ToStateSlice(state), prefix) },
		GetParam:    WrapConvertRetSlice[T, any](param),
		BuildParams: doNotUseParameters[T],
	}
}

func NewJsonFilterProperty[T any](query string, param func(T) any) *FilterProperty[T] {
	return &FilterProperty[T]{
		BuildQuery: func(json []any) string { return buildJsonFilterQuery(ToJsonSlice(json), query, OP_OR) },
		GetParam:   WrapConvertRetSlice[T, any](param),
		BuildParams: func(filter T) []any {
			return buildJsonQueryParameters([]any{}, WrapConvertRetSlice[T, *entity.Json](param)(filter))
		},
	}
}

func NewCustomFilterProperty[T any](
	buildQuery func([]any) string,
	getParam func(T) any,
) *FilterProperty[T] {
	return &FilterProperty[T]{
		BuildQuery:  buildQuery,
		GetParam:    WrapConvertRetSlice[T, any](getParam),
		BuildParams: doNotUseParameters[T],
	}
}

// Join
type JoinType string

const (
	LeftJoin  JoinType = "LEFT JOIN"
	RightJoin JoinType = "RIGHT JOIN"
	InnerJoin JoinType = "JOIN"
)

func DependentJoin[T any](T, *Order) bool { return false }

type JoinDef[T any] struct {
	Name      string
	Type      JoinType
	Table     string
	On        string
	DependsOn []string
	Condition func(T, *Order) bool
}

type JoinResolver[T any] struct {
	defs       []*JoinDef[T]
	included   map[string]bool
	order      []string
	forcedJoin []string
}

func NewJoinResolver[T any](defs []*JoinDef[T], forcedJoin []string) *JoinResolver[T] {
	r := &JoinResolver[T]{
		defs:       defs,
		included:   map[string]bool{},
		forcedJoin: forcedJoin,
	}

	return r
}

func (jr *JoinResolver[T]) require(name string) {
	if jr.included[name] {
		return
	}

	def, ok := lo.Find(jr.defs, func(jd *JoinDef[T]) bool {
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

func (jr *JoinResolver[T]) Build(filter T, order *Order) []string {
	for _, n := range jr.forcedJoin {
		jr.require(n)
	}

	for _, def := range jr.defs {
		if def.Condition == nil || (def.Condition != nil && def.Condition(filter, order)) {
			jr.require(def.Name)
		}
	}

	var result []string

	// This is little tricky part, but we need to deal with that this way
	// until we have stateful join pattern which is present in issue.go
	// using non-uniq tablename 'IM IssueMatch' which join operation
	// depending on filter pattern with precedence for some members (there
	// is 'if...else if' statement which cannot be replaced by consecutive
	// 'if...' and 'if...' what is a mess and misconception
	uniqTableName := make(map[string]struct{})

	for _, name := range jr.order {
		j, ok := lo.Find(jr.defs, func(jd *JoinDef[T]) bool {
			return jd.Name == name
		})
		if !ok {
			panic("JoinResolver::Build(...) Unknown join: " + name)
		}

		if _, ok := uniqTableName[j.Table]; ok {
			continue
		}

		uniqTableName[j.Table] = struct{}{}

		joinSQL := fmt.Sprintf(
			"%s %s ON %s",
			j.Type,
			j.Table,
			j.On,
		)

		result = append(result, joinSQL)
	}

	return result
}

// DB helpers
func EnsureFilter[T entity.Filter](filter T) T {
	return mustConvert[T](ensurePagination(filter.Ensure()))
}

func ensurePagination(filter entity.Filter) entity.Filter {
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

func AssociateId(db Db, tableName string, tablePrefix string, prefix1 string, id1 int64, prefix2 string, id2 int64) error {
	l := logrus.WithFields(logrus.Fields{
		prefix1: id1,
		prefix2: id2,
		"event": "database.AssociateId (" + prefix1 + ", " + prefix2 + ")",
	})

	qb := sq.
		Insert(tableName).
		Columns(tablePrefix+"_"+prefix1+"_id", tablePrefix+"_"+prefix2+"_id").
		Values(id1, id2)

	query, params, err := qb.ToSql()
	if err != nil {
		l.Error(err)

		return err
	}

	var mysqlErr *mysql.MySQLError

	_, err = db.Exec(query, params...)
	if err != nil {
		if errors.As(err, &mysqlErr) && mysqlErr.Number == database.ErrCodeDuplicateEntry {
			return nil
		}

		l.WithFields(logrus.Fields{
			"error": err,
			"query": query,
			"args":  params,
		}).Error(err)

		return err
	}

	return nil
}

func AssociateIdWithVal(db Db, tableName string, tablePrefix string, prefix1 string, id1 int64, prefix2 string, id2 int64, valName string, val int64) error {
	l := logrus.WithFields(logrus.Fields{
		prefix1: id1,
		prefix2: id2,
		valName: val,
		"event": "database.AssociateIdWithVal (" + prefix1 + ", " + prefix2 + ", " + valName + ")",
	})

	qb := sq.
		Insert(tableName).
		Columns(tablePrefix+"_"+prefix1+"_id", tablePrefix+"_"+prefix2+"_id", tablePrefix+"_"+valName).
		Values(id1, id2, val)

	query, params, err := qb.ToSql()
	if err != nil {
		l.Error(err)

		return err
	}

	var mysqlErr *mysql.MySQLError

	_, err = db.Exec(query, params...)
	if err != nil {
		if errors.As(err, &mysqlErr) && mysqlErr.Number == database.ErrCodeDuplicateEntry {
			return nil
		}

		l.WithFields(logrus.Fields{
			"error": err,
			"query": query,
			"args":  params,
		}).Error(err)

		return err
	}

	return nil
}

func DissociateId(db Db, tableName string, tablePrefix string, prefix1 string, id1 int64, prefix2 string, id2 int64) error {
	l := logrus.WithFields(logrus.Fields{
		prefix1: id1,
		prefix2: id2,
		"event": "database.DissociateId (" + prefix1 + ", " + prefix2 + ")",
	})

	qb := sq.
		Delete(tableName).
		Where(sq.Eq{
			tablePrefix + "_" + prefix1 + "_id": id1,
			tablePrefix + "_" + prefix2 + "_id": id2,
		})

	query, params, err := qb.ToSql()
	if err != nil {
		l.Error(err)

		return err
	}

	_, err = db.Exec(query, params...)
	if err != nil {
		l.WithFields(logrus.Fields{
			"error": err,
			"query": query,
			"args":  params,
		}).Error(err)
	}

	return err
}

func DissociateAllIds(db Db, tableName string, tablePrefix string, prefix string, id int64) error {
	l := logrus.WithFields(logrus.Fields{
		prefix:  id,
		"event": "database.DissociateAllIds (" + prefix + ")",
	})

	qb := sq.
		Delete(tableName).
		Where(sq.Eq{
			tablePrefix + "_" + prefix + "_id": id,
		})

	query, params, err := qb.ToSql()
	if err != nil {
		l.Error(err)

		return err
	}

	_, err = db.Exec(query, params...)
	if err != nil {
		l.WithFields(logrus.Fields{
			"error": err,
			"query": query,
			"args":  params,
		}).Error(err)
	}

	return err
}

// Helpers

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

// WrapConvertRetSlice turns a type-specific accessor into a specified one
func WrapConvertRetSlice[T any, R any](fn func(T) any) func(T) []R {
	return func(v T) []R {
		res := fn(v)

		rv := reflect.ValueOf(res)

		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			panic("function did not return slice or array")
		}

		out := make([]R, rv.Len())

		for i := 0; i < rv.Len(); i++ {
			item, ok := rv.Index(i).Interface().(R)
			if !ok {
				panic(fmt.Sprintf("cannot convert element %d to target type", i))
			}

			out[i] = item
		}

		return out
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

func mustConvert[T any](v any) T {
	res, ok := v.(T)
	if !ok {
		panic(fmt.Sprintf("cannot convert %T to target type", v))
	}

	return res
}
