// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"strings"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type dummyEntity struct {
	Id int64
	A  string
}

func (de *dummyEntity) GetId() int64 {
	return de.Id
}

func (de *dummyEntity) SetId(id int64) {
	de.Id = id
}

type dummyEntityFilter struct {
	entity.Paginated
	Id []*int64
	A  []*string
	B  []*int
}

var _ = Describe("DbObject", Label("database", "DbObject"), func() {
	const (
		dummyId                      = 7
		dummyString                  = "ttt"
		anyVal                       = false
		noPagination                 = false
		pagination                   = true
		dummyCursorFieldVal          = 9
		defaultFirstEntriesForCursor = 1000
	)

	dummyCursorField := mariadb.Field{Name: entity.OrderByField(4), Value: dummyCursorFieldVal, Order: entity.OrderDirectionDesc}

	When("Insert query is called", func() {
		Context("Two properties are there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity]{
				TableName: "DummyTable",
				Properties: []*mariadb.Property{
					mariadb.NewProperty("id", mariadb.WrapAccess(func(de *dummyEntity) (int64, bool) { return de.Id, anyVal })),
					mariadb.NewProperty("a", mariadb.WrapAccess(func(de *dummyEntity) (string, bool) { return de.A, anyVal })),
				},
			}
			It("can build insert query and query parameters", func() {
				de := dummyEntity{Id: dummyId, A: dummyString}
				query, params, err := testObject.InsertQuery(&de)
				Expect(err).To(BeNil())
				Expect(query).To(BeEquivalentTo("INSERT INTO DummyTable (id,a) VALUES (?,?)"))
				Expect(params).To(HaveLen(2))
				Expect(params[0]).To(BeEquivalentTo(dummyId))
				Expect(params[1]).To(BeEquivalentTo(dummyString))
			})
		})
	})
	When("Filter query is called", func() {
		Context("Two filters are there in the filter object", func() {
			testObject := mariadb.DbObject[*dummyEntity]{
				TableName: "DummyTable",
				FilterProperties: []*mariadb.FilterProperty{
					mariadb.NewFilterProperty("DT.dummytable_id = ?", mariadb.WrapRetSlice(func(filter *dummyEntityFilter) []*int64 { return filter.Id })),
					mariadb.NewFilterProperty("DT.dummytable_a = ?", mariadb.WrapRetSlice(func(filter *dummyEntityFilter) []*string { return filter.A })),
				},
			}
			It("builds filter for one filter item", func() {
				def := dummyEntityFilter{A: []*string{lo.ToPtr(dummyString)}}
				By("returning correct filter query string", func() {
					query := testObject.GetFilterQuery(&def)
					noWSQuery := strings.ReplaceAll(query, " ", "")
					Expect(noWSQuery).To(BeEquivalentTo("((DT.dummytable_a=?))"))
				})
				By("returning correct parameter", func() {
					params := testObject.GetFilterParameters(&def, noPagination, nil)
					Expect(params).To(HaveLen(1))
					Expect(params[0]).To(HaveValue(Equal(dummyString)))
				})
				By("returning correct parameter with cursor parameters", func() {
					params := testObject.GetFilterParameters(&def, pagination, []mariadb.Field{dummyCursorField})
					Expect(params).To(HaveLen(3))
					Expect(params[0]).To(HaveValue(Equal(dummyString)))
					Expect(params[1]).To(HaveValue(Equal(dummyCursorFieldVal)))
					Expect(params[2]).To(HaveValue(Equal(defaultFirstEntriesForCursor)))
				})
			})

			It("builds filter for two filter items", func() {
				id := int64(dummyId)
				def := dummyEntityFilter{Id: []*int64{&id}, A: []*string{lo.ToPtr(dummyString)}}
				By("returning correct filter query string", func() {
					query := testObject.GetFilterQuery(&def)
					noWSQuery := strings.ReplaceAll(query, " ", "")
					Expect(noWSQuery).To(BeEquivalentTo("((DT.dummytable_id=?)AND(DT.dummytable_a=?))"))
				})
				By("returning correct parameters", func() {
					params := testObject.GetFilterParameters(&def, noPagination, nil)
					Expect(params).To(HaveLen(2))
					Expect(params[0]).To(HaveValue(Equal(int64(dummyId))))
					Expect(params[1]).To(HaveValue(Equal(dummyString)))
				})
			})

			It("builds empty filter query string for no filter", func() {
				def := dummyEntityFilter{}
				query := testObject.GetFilterQuery(&def)
				Expect(query).To(BeEmpty())
			})
		})
	})
	When("Joins query is build", func() {
		Context("no joins are there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity]{
				TableName: "DummyTable",
			}
			It("gets empty string as there is no join defined", func() {
				def := dummyEntityFilter{A: []*string{lo.ToPtr(dummyString)}}
				joins := testObject.GetJoins(&def, nil)
				Expect(joins).To(BeEmpty())
			})
		})
		Context("Two basic joins are there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity]{
				TableName: "DummyTable",
				JoinDefs: []*mariadb.JoinDef{
					{
						Name:      "X",
						Type:      mariadb.LeftJoin,
						Table:     "Xabc X",
						On:        "DT.dummytable_id = X.xabc_dummytable_id",
						Condition: mariadb.WrapJoinCondition(func(f *dummyEntityFilter, _ []entity.Order) bool { return len(f.A) > 0 }),
					},
					{
						Name:      "Y",
						Type:      mariadb.RightJoin,
						Table:     "Yabc Y",
						On:        "DT.dummytable_id = Y.yabc_dummytable_id",
						Condition: mariadb.WrapJoinCondition(func(f *dummyEntityFilter, _ []entity.Order) bool { return len(f.B) > 0 }),
					},
				},
			}
			It("gets empty string when there is empty filter used", func() {
				def := dummyEntityFilter{}
				joins := testObject.GetJoins(&def, nil)
				Expect(joins).To(BeEmpty())
			})
			It("gets one join with X when A mapping condition is met (there is at least one element to filter)", func() {
				def := dummyEntityFilter{A: []*string{lo.ToPtr(dummyString)}}
				joins := testObject.GetJoins(&def, nil)
				Expect(joins).To(BeEquivalentTo("LEFT JOIN Xabc X ON DT.dummytable_id = X.xabc_dummytable_id"))
			})
			It("gets one join with Y when B mapping condition is met (there is at least one element to filter)", func() {
				def := dummyEntityFilter{B: []*int{lo.ToPtr(10)}}
				joins := testObject.GetJoins(&def, nil)
				Expect(joins).To(BeEquivalentTo("RIGHT JOIN Yabc Y ON DT.dummytable_id = Y.yabc_dummytable_id"))
			})
			It("gets two joins with X and Y when A and B mapping conditions are met (there are at least one element for each filter)", func() {
				def := dummyEntityFilter{A: []*string{lo.ToPtr(dummyString)}, B: []*int{lo.ToPtr(10)}}
				joins := testObject.GetJoins(&def, nil)
				lines := strings.Split(strings.TrimSpace(joins), "\n")
				Expect(lines).To(HaveLen(2))
				Expect(lines).To(ConsistOf(
					MatchRegexp("^LEFT JOIN Xabc X ON DT.dummytable_id = X.xabc_dummytable_id$"),
					MatchRegexp("^RIGHT JOIN Yabc Y ON DT.dummytable_id = Y.yabc_dummytable_id$"),
				))
			})
		})
		Context("Join with dependency is there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity]{
				TableName: "DummyTable",
				JoinDefs: []*mariadb.JoinDef{
					{
						Name:      "X",
						Type:      mariadb.LeftJoin,
						Table:     "Xabc X",
						On:        "DT.dummytable_id = X.xabc_dummytable_id",
						Condition: mariadb.DependentJoin,
					},
					{
						Name:      "Y",
						Type:      mariadb.RightJoin,
						Table:     "Yabc Y",
						On:        "X.xabc_yabc_id = Y.yabc_id",
						DependsOn: []string{"X"},
						Condition: mariadb.WrapJoinCondition(func(f *dummyEntityFilter, _ []entity.Order) bool { return len(f.B) > 0 }),
					},
				},
			}
			It("gets two joins with X (and dependent Y) when B mapping condition is met (there is at least one element to filter)", func() {
				def := dummyEntityFilter{B: []*int{lo.ToPtr(10)}}
				joins := testObject.GetJoins(&def, nil)
				lines := strings.Split(strings.TrimSpace(joins), "\n")
				Expect(lines).To(HaveLen(2))
				Expect(lines[0]).To(BeEquivalentTo("LEFT JOIN Xabc X ON DT.dummytable_id = X.xabc_dummytable_id"))
				Expect(lines[1]).To(BeEquivalentTo("RIGHT JOIN Yabc Y ON X.xabc_yabc_id = Y.yabc_id"))
			})
		})
		Context("Two joins with the same dependency are there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity]{
				TableName: "DummyTable",
				JoinDefs: []*mariadb.JoinDef{
					{
						Name:      "X",
						Type:      mariadb.LeftJoin,
						Table:     "Xabc X",
						On:        "DT.dummytable_id = X.xabc_dummytable_id",
						Condition: mariadb.DependentJoin,
					},
					{
						Name:      "Y",
						Type:      mariadb.RightJoin,
						Table:     "Yabc Y",
						On:        "X.xabc_yabc_id = Y.yabc_id",
						DependsOn: []string{"X"},
						Condition: mariadb.WrapJoinCondition(func(f *dummyEntityFilter, _ []entity.Order) bool { return len(f.B) > 0 }),
					},
					{
						Name:      "Z",
						Type:      mariadb.RightJoin,
						Table:     "Zabc Z",
						On:        "X.xabc_zabc_id = Z.zabc_id",
						DependsOn: []string{"X"},
						Condition: mariadb.WrapJoinCondition(func(f *dummyEntityFilter, _ []entity.Order) bool { return len(f.A) > 0 }),
					},
				},
			}
			It("gets two joins and one dependent join when A and B mapping conditions are met (there is at least one element for each filter)", func() {
				def := dummyEntityFilter{A: []*string{lo.ToPtr(dummyString)}, B: []*int{lo.ToPtr(10)}}
				joins := testObject.GetJoins(&def, nil)
				lines := strings.Split(strings.TrimSpace(joins), "\n")
				Expect(lines).To(HaveLen(3))
				Expect(lines[0]).To(BeEquivalentTo("LEFT JOIN Xabc X ON DT.dummytable_id = X.xabc_dummytable_id"))
				Expect(lines[1:]).To(ConsistOf(
					MatchRegexp("^RIGHT JOIN Yabc Y ON X.xabc_yabc_id = Y.yabc_id$"),
					MatchRegexp("^RIGHT JOIN Zabc Z ON X.xabc_zabc_id = Z.zabc_id$"),
				))
			})
		})
		Context("Two joins on the same table are there in the object", func() {
			testObject := mariadb.DbObject[*dummyEntity]{
				TableName: "DummyTable",
				JoinDefs: []*mariadb.JoinDef{
					{
						Name:      "X_left",
						Type:      mariadb.LeftJoin,
						Table:     "Xabc X",
						On:        "DT.dummytable_id = X.xabc_dummytable_id",
						Condition: mariadb.WrapJoinCondition(func(f *dummyEntityFilter, _ []entity.Order) bool { return len(f.B) > 0 }),
					},
					{
						Name:      "X_right",
						Type:      mariadb.RightJoin,
						Table:     "Xabc X",
						On:        "DT.dummytable_id = X.xabc_dummytable_id",
						Condition: mariadb.WrapJoinCondition(func(f *dummyEntityFilter, _ []entity.Order) bool { return len(f.B) > 0 }),
					},
				},
			}
			It("gets one join even when both join defs have the same condition but only first one will be included to prevent join table name duplication", func() {
				def := dummyEntityFilter{B: []*int{lo.ToPtr(10)}}
				joins := testObject.GetJoins(&def, nil)
				Expect(joins).To(BeEquivalentTo("LEFT JOIN Xabc X ON DT.dummytable_id = X.xabc_dummytable_id"))
			})
		})
	})
})
