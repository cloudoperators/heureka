// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb_test

import (
	"database/sql"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

type TestStruct struct {
	Id        sql.NullInt64  `cursor:"id"`
	Name      sql.NullString `cursor:"name"`
	Timestamp sql.NullTime   `cursor:"timestamp"`
}

type ToCompositeStruct struct {
	Additional sql.NullString `cursor:"additional"`
}

type NotToSet struct {
	NotToSet sql.NullString `cursor:"not_to_set"`
}

type CompositeStructPointer struct {
	*TestStruct
	*ToCompositeStruct
	*NotToSet
}

type CompositeStructNoPointer struct {
	TestStruct
	ToCompositeStruct
	NotToSet
}

// Example execution:
// go test --fuzztime 30s github.com/cloudoperators/heureka/internal/database/mariadb -fuzz ^FuzzMarshalCursor$ -run ^$
func FuzzMarshalCursor(f *testing.F) {
	f.Add(int64(1), "test", "2023-10-01T00:00:00Z", "additional")
	f.Add(int64(0), "", "invalid-timestamp", "")
	f.Add(int64(1), "test", "1700-10-01T00:00:00Z", "2023-10-01T00:00:00Z")
	f.Fuzz(func(t *testing.T, id int64, name string, nameValid bool, timestamp string, timestampValid bool, additional string, additionalValid bool) {
		ts := TestStruct{
			Id:        sql.NullInt64{Int64: id, Valid: true},
			Name:      sql.NullString{String: name, Valid: true},
			Timestamp: sql.NullTime{Time: parseTime(timestamp), Valid: true},
		}

		tcs := ToCompositeStruct{
			Additional: sql.NullString{String: additional, Valid: true},
		}

		cs := CompositeStructPointer{
			TestStruct:        &ts,
			ToCompositeStruct: &tcs,
		}

		cso := CompositeStructNoPointer{
			TestStruct:        ts,
			ToCompositeStruct: tcs,
		}

		cursor := mariadb.MarshalCursor(cs)
		if cursor == "" {
			t.Errorf("Expected non-empty cursor string")
		}
		cursor = mariadb.MarshalCursor(ts)
		if cursor == "" {
			t.Errorf("Expected non-empty cursor string")
		}

		cursor = mariadb.MarshalCursor(tcs)
		if cursor == "" {
			t.Errorf("Expected non-empty cursor string")
		}

		cursor = mariadb.MarshalCursor(cso)
		if cursor == "" {
			t.Errorf("Expected non-empty cursor string")
		}
	})
}

// Execute with timelimit to avoid infinite loop
// Example:
//
//	go test --fuzztime 30s github.com/cloudoperators/heureka/internal/database/mariadb -fuzz ^FuzzUnmarshalCursor$ -run ^$
func FuzzUnmarshalCursor(f *testing.F) {
	f.Add("id=1, name=test, timestamp=2023-10-01T00:00:00Z, additional=additional")
	f.Add("id=, name=, timestamp=, additional=")
	f.Add("black=1, love=isreal")
	f.Add("my, id=1, name=test, john, timestamp=2023-10-01T00:00:00Z, additional=additional, memo")
	f.Fuzz(func(t *testing.T, cursorStr string) {
		cs := CompositeStructPointer{
			TestStruct:        &TestStruct{},
			ToCompositeStruct: &ToCompositeStruct{},
		}
		err := mariadb.UnmarshalCursor(cursorStr, &cs)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func parseTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return t
}

var _ = Describe("MarshalCursor", func() {
	When("Using a composite struct with pointers", func() {
		It("should marshal composite struct with valid fields", func() {
			cs := CompositeStructPointer{
				TestStruct: &TestStruct{
					Id:        sql.NullInt64{Int64: 1, Valid: true},
					Name:      sql.NullString{String: "test", Valid: true},
					Timestamp: sql.NullTime{Time: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC), Valid: true},
				},
				ToCompositeStruct: &ToCompositeStruct{
					Additional: sql.NullString{String: "additional", Valid: true},
				},
				NotToSet: &NotToSet{
					NotToSet: sql.NullString{String: "", Valid: false},
				},
			}
			cursor := mariadb.MarshalCursor(cs)
			Expect(cursor).To(Equal("id=1, name=test, timestamp=2023-10-01T00:00:00Z, additional=additional"))
		})

		It("should skip nil fields in composite struct", func() {
			cs := CompositeStructPointer{
				TestStruct: &TestStruct{
					Id:        sql.NullInt64{Int64: 1, Valid: true},
					Name:      sql.NullString{String: "test", Valid: true},
					Timestamp: sql.NullTime{Time: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC), Valid: true},
				},
				ToCompositeStruct: nil,
				NotToSet:          nil,
			}
			cursor := mariadb.MarshalCursor(cs)
			Expect(cursor).To(Equal("id=1, name=test, timestamp=2023-10-01T00:00:00Z"))
		})
	})
	When("Using a composite struct without pointers", func() {
		It("should marshal composite struct with valid fields", func() {
			cs := CompositeStructNoPointer{
				TestStruct: TestStruct{
					Id:        sql.NullInt64{Int64: 1, Valid: true},
					Name:      sql.NullString{String: "test", Valid: true},
					Timestamp: sql.NullTime{Time: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC), Valid: true},
				},
				ToCompositeStruct: ToCompositeStruct{
					Additional: sql.NullString{String: "additional", Valid: true},
				},
				NotToSet: NotToSet{
					NotToSet: sql.NullString{String: "", Valid: false},
				},
			}
			cursor := mariadb.MarshalCursor(cs)
			Expect(cursor).To(Equal("id=1, name=test, timestamp=2023-10-01T00:00:00Z, additional=additional"))
		})

		It("should skip invalid fields in composite struct", func() {
			cs := CompositeStructNoPointer{
				TestStruct: TestStruct{
					Id:        sql.NullInt64{Int64: 0, Valid: false},
					Name:      sql.NullString{String: "", Valid: false},
					Timestamp: sql.NullTime{Time: time.Time{}, Valid: false},
				},
				ToCompositeStruct: ToCompositeStruct{
					Additional: sql.NullString{String: "", Valid: false},
				},
				NotToSet: NotToSet{
					NotToSet: sql.NullString{String: "", Valid: false},
				},
			}
			cursor := mariadb.MarshalCursor(cs)
			Expect(cursor).To(Equal(""))
		})
	})
	When("Using a normal struct", func() {
		It("should marshal struct with valid fields", func() {
			ts := TestStruct{
				Id:        sql.NullInt64{Int64: 1, Valid: true},
				Name:      sql.NullString{String: "test", Valid: true},
				Timestamp: sql.NullTime{Time: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC), Valid: true},
			}
			cursor := mariadb.MarshalCursor(ts)
			Expect(cursor).To(Equal("id=1, name=test, timestamp=2023-10-01T00:00:00Z"))
		})

		It("should skip fields with invalid values", func() {
			ts := TestStruct{
				Id:        sql.NullInt64{Int64: 0, Valid: false},
				Name:      sql.NullString{String: "", Valid: false},
				Timestamp: sql.NullTime{Time: time.Time{}, Valid: false},
			}
			cursor := mariadb.MarshalCursor(ts)
			Expect(cursor).To(Equal(""))
		})

		It("should handle mixed valid and invalid fields", func() {
			ts := TestStruct{
				Id:        sql.NullInt64{Int64: 1, Valid: true},
				Name:      sql.NullString{String: "", Valid: false},
				Timestamp: sql.NullTime{Time: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC), Valid: true},
			}
			cursor := mariadb.MarshalCursor(ts)
			Expect(cursor).To(Equal("id=1, timestamp=2023-10-01T00:00:00Z"))
		})
	})
})

type UnMarshalTestStruct struct {
	Id        sql.NullInt64  `json:"some_id" cursor:"id" meaningless:"true"`
	Name      sql.NullString `cursor:"name"`
	Severity  sql.NullString `cursor:"severity" cursor_enum:"cirtical,high,medium,low"`
	Timestamp sql.NullTime   `cursor:"timestamp"`
}

var _ = Describe("UnmarshalCursor", func() {
	When("Using a composite struct with pointers", func() {
		It("should unmarshal cursor string into sub-structs with valid fields", func() {
			cursorStr := "id=1, name=test, timestamp=2023-10-01T00:00:00Z, additional=something, not_to_set=ignore_me"
			cs := CompositeStructPointer{
				TestStruct:        &TestStruct{},
				ToCompositeStruct: &ToCompositeStruct{},
				NotToSet:          nil,
			}
			err := mariadb.UnmarshalCursor(cursorStr, &cs)
			Expect(err).To(BeNil())
			Expect(cs.TestStruct.Id.Valid).To(BeTrue())
			Expect(cs.TestStruct.Id.Int64).To(Equal(int64(1)))
			Expect(cs.TestStruct.Name.Valid).To(BeTrue())
			Expect(cs.TestStruct.Name.String).To(Equal("test"))
			Expect(cs.TestStruct.Timestamp.Valid).To(BeTrue())
			Expect(cs.TestStruct.Timestamp.Time).To(Equal(time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)))
			Expect(cs.ToCompositeStruct.Additional.Valid).To(BeTrue())
			Expect(cs.ToCompositeStruct.Additional.String).To(Equal("something"))
			Expect(cs.NotToSet).To(BeNil(), "should not set fields that are null in the given composite ")
		})
	})

	When("Using a composite struct without pointers", func() {
		It("should unmarshal cursor string into sub-structs with valid fields", func() {
			cursorStr := "id=1, name=test, timestamp=2023-10-01T00:00:00Z, additional=something, not_to_set=not_ignore_me"
			cs := CompositeStructNoPointer{}
			err := mariadb.UnmarshalCursor(cursorStr, &cs)
			Expect(err).To(BeNil())
			Expect(cs.TestStruct.Id.Valid).To(BeTrue())
			Expect(cs.TestStruct.Id.Int64).To(Equal(int64(1)))
			Expect(cs.TestStruct.Name.Valid).To(BeTrue())
			Expect(cs.TestStruct.Name.String).To(Equal("test"))
			Expect(cs.TestStruct.Timestamp.Valid).To(BeTrue())
			Expect(cs.TestStruct.Timestamp.Time).To(Equal(time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)))
			Expect(cs.ToCompositeStruct.Additional.Valid).To(BeTrue())
			Expect(cs.ToCompositeStruct.Additional.String).To(Equal("something"))
			Expect(cs.NotToSet.NotToSet.Valid).To(BeTrue())
			Expect(cs.NotToSet.NotToSet.String).To(Equal("not_ignore_me"))
		})
	})

	When("Using a normal struct", func() {
		It("should unmarshal cursor string into struct with valid fields", func() {
			cursorStr := "id=1, name=test, timestamp=2023-10-01T00:00:00Z"
			var ts UnMarshalTestStruct
			err := mariadb.UnmarshalCursor(cursorStr, &ts)
			Expect(err).To(BeNil())
			Expect(ts.Id.Valid).To(BeTrue())
			Expect(ts.Id.Int64).To(Equal(int64(1)))
			Expect(ts.Name.Valid).To(BeTrue())
			Expect(ts.Name.String).To(Equal("test"))
			Expect(ts.Timestamp.Valid).To(BeTrue())
			Expect(ts.Timestamp.Time).To(Equal(time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)))
		})

		It("should handle cursor string with invalid fields", func() {
			cursorStr := "id=, name=, timestamp="
			var ts UnMarshalTestStruct
			err := mariadb.UnmarshalCursor(cursorStr, &ts)
			Expect(err).To(BeNil())
			Expect(ts.Id.Valid).To(BeFalse())
			Expect(ts.Name.Valid).To(BeFalse())
			Expect(ts.Timestamp.Valid).To(BeFalse())
		})

		It("should handle mixed valid and invalid fields", func() {
			cursorStr := "id=1, name=, timestamp=2023-10-01T00:00:00Z"
			var ts UnMarshalTestStruct
			err := mariadb.UnmarshalCursor(cursorStr, &ts)
			Expect(err).To(BeNil())
			Expect(ts.Id.Valid).To(BeTrue())
			Expect(ts.Id.Int64).To(Equal(int64(1)))
			Expect(ts.Name.Valid).To(BeFalse())
			Expect(ts.Timestamp.Valid).To(BeTrue())
			Expect(ts.Timestamp.Time).To(Equal(time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)))
		})

		It("should ignore unkown fields", func() {
			cursorStr := "id=1, name=test, invalid_name=2023-10-01T00:00:00Z"
			var ts UnMarshalTestStruct
			err := mariadb.UnmarshalCursor(cursorStr, &ts)
			Expect(err).To(BeNil())
			Expect(ts.Id.Valid).To(BeTrue())
			Expect(ts.Id.Int64).To(Equal(int64(1)))
			Expect(ts.Name.Valid).To(BeTrue())
			Expect(ts.Name.String).To(Equal("test"))
		})

		It("should mark unset fields as invalid", func() {
			cursorStr := "id=1"
			var ts UnMarshalTestStruct
			err := mariadb.UnmarshalCursor(cursorStr, &ts)
			Expect(err).To(BeNil())
			Expect(ts.Id.Valid).To(BeTrue())
			Expect(ts.Id.Int64).To(Equal(int64(1)))
			Expect(ts.Name.Valid).To(BeFalse())
			Expect(ts.Timestamp.Valid).To(BeFalse())
		})

		It("should unmarshal cursor string into struct with valid fields and enum", func() {
			cursorStr := "id=1, name=test, severity=high, timestamp=2023-10-01T00:00:00Z"
			var ts UnMarshalTestStruct
			err := mariadb.UnmarshalCursor(cursorStr, &ts)
			Expect(err).To(BeNil())
			By("setting the field to high correct value and valid to true", func() {
				Expect(ts.Severity.String).To(Equal("high"))
				Expect(ts.Timestamp.Valid).To(BeTrue())
			})
		})

		It("should handle cursor string with invalid enum", func() {
			cursorStr := "id=1, name=test, severity=invalid, timestamp=2023-10-01T00:00:00Z"
			var ts UnMarshalTestStruct
			err := mariadb.UnmarshalCursor(cursorStr, &ts)
			Expect(err).To(BeNil())
			By("setting the field to invalid", func() {
				Expect(ts.Severity.Valid).To(BeFalse())
			})
		})
	})
})
