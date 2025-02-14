package baseResolver

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"k8s.io/utils/pointer"
)

func ScannerRunTagFilterValues(app app.Heureka, ctx context.Context) ([]*string, error) {
	tags, err := app.GetScannerRunTags()

	if err != nil {
		return nil, err
	}

	res := make([]*string, len(tags))
	for _, tag := range tags {
		res = append(res, pointer.String(tag))
	}

	return res, nil
}

func ScannerRuns(app app.Heureka, ctx context.Context, filter *model.ScannerRunFilter, first *int, after *string) ([]*model.ScannerRunConnection, error) {
	panic("not implemented")
	return nil, nil
}
