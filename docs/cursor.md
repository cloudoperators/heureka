# Cursor

The cursor is a list of `Field`, where `Field` is defined as:

```go
type Field struct {
	Name  DbColumnName
	Value any
	Order OrderDirection
}
```

This list will be encoded as a base64 string.

Each entity defines its own cursor values. For example `IssueMatch` allows the following cursor fields:

- IssueMatchId
- IssueMatchTargetRemediationDate
- IssueMatchRating
- ComponentInstanceCCRN
- IssuePrimaryName

```go
func WithIssueMatch(order []Order, im IssueMatch) NewCursor {

	return func(cursors *cursors) error {
		cursors.fields = append(cursors.fields, Field{Name: IssueMatchId, Value: im.Id, Order: OrderDirectionAsc})
		cursors.fields = append(cursors.fields, Field{Name: IssueMatchTargetRemediationDate, Value: im.TargetRemediationDate, Order: OrderDirectionAsc})
		cursors.fields = append(cursors.fields, Field{Name: IssueMatchRating, Value: im.Severity.Value, Order: OrderDirectionAsc})

		if im.ComponentInstance != nil {
			cursors.fields = append(cursors.fields, Field{Name: ComponentInstanceCcrn, Value: im.ComponentInstance.CCRN, Order: OrderDirectionAsc})
		}
		if im.Issue != nil {
			cursors.fields = append(cursors.fields, Field{Name: IssuePrimaryName, Value: im.Issue.PrimaryName, Order: OrderDirectionAsc})
		}

		m := CreateOrderMap(order)
		for _, f := range cursors.fields {
			if orderDirection, ok := m[f.Name]; ok {
				f.Order = orderDirection
			}
		}

		return nil
	}
}
```

The cursor is returned by the database layer an can be encoded such as:

```go
    cursor, _ := entity.EncodeCursor(entity.WithIssueMatch(order, im))
```

A order list can be passed to override the default ordering.

## Cursor Query

The cursor points to the starting point in a list of database rows. All elements *after* the cursor are returned.
Depending on the ordering, the query looks like:

```sql
WHERE id < cursor_id

Or:

Where id > cursor_id

```

If the cursor contains two fields, the query needs to check for the second field, if the first field is equal:

```sql
WHERE (id = cursor_id AND primaryName > cursor_primaryName) OR (id > cursor_id)

```

Similarly, for three fields:
```sql
WHERE 
    (id = cursor_id AND primaryName = cursor_primaryName AND trd > cursor_trd) OR
    (id = cursor_id AND primaryName > cursor_primaryName) OR 
    (id > cursor_id)
```
