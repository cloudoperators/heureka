# Order

## API

The query contains an additional `orderBy` argument:

```graphql
IssueMatches(filter: IssueMatchFilter, first: Int, after: String, orderBy: [IssueMatchOrderBy]): IssueMatchConnection
```

The OrderBy input is defined for each model:

```graphql
input IssueMatchOrderBy {
    by: IssueMatchOrderByField
    direction: OrderDirection
}
```

The `By` fields define the allowed order options:

```graphql
enum IssueMatchOrderByField {
    primaryName
    targetRemediationDate
    componentInstanceCcrn
}
```

The `OrderDirections` are defined in the `common.graphqls`:
```graphql
enum OrderDirection {
    asc
    desc
}
```

The generated order models are converted to the entity order model in `api/graph/model/models.go`:

```go
func (imo *IssueMatchOrderBy) ToOrderEntity() entity.Order {
	var order entity.Order
	switch *imo.By {
	case IssueMatchOrderByFieldPrimaryName:
		order.By = entity.IssuePrimaryName
	case IssueMatchOrderByFieldComponentInstanceCcrn:
		order.By = entity.ComponentInstanceCcrn
	case IssueMatchOrderByFieldTargetRemediationDate:
		order.By = entity.IssueMatchTargetRemediationDate
	}
	order.Direction = imo.Direction.ToOrderDirectionEntity()
	return order
}
```

## Entity

```go
type Order struct {
	By        DbColumnName
	Direction OrderDirection
}
```

The `By` field is the database column name and is defined as a constant:

```go
var DbColumnNameMap = map[DbColumnName]string{
	ComponentInstanceCcrn:           "componentinstance_ccrn",
	IssuePrimaryName:                "issue_primary_name",
	IssueMatchId:                    "issuematch_id",
	IssueMatchRating:                "issuematch_rating",
	IssueMatchTargetRemediationDate: "issuematch_target_remediation_date",
	SupportGroupName:                "supportgroup_name",
}
```


## Database

The `GetIssueMatches()` function has an additional order argument:

```go
func (s *SqlDatabase) GetIssueMatches(filter *entity.IssueMatchFilter, order []entity.Order) ([]entity.IssueMatchResult, error) {
    ...
}
```

The order string is created by in `entity/order.go`:

```go
func CreateOrderString(order []Order) string {
	orderStr := ""
	for i, o := range order {
		if i > 0 {
			orderStr = fmt.Sprintf("%s, %s %s", orderStr, o.By, o.Direction)
		} else {
			orderStr = fmt.Sprintf("%s %s %s", orderStr, o.By, o.Direction)
		}
	}
	return orderStr
}
```


