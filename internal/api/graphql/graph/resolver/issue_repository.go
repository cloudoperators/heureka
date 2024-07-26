package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.49

import (
	"context"

	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/graph"
	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/graph/baseResolver"
	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/graph/model"
)

// IssueVariants is the resolver for the issueVariants field.
func (r *issueRepositoryResolver) IssueVariants(ctx context.Context, obj *model.IssueRepository, filter *model.IssueVariantFilter, first *int, after *string) (*model.IssueVariantConnection, error) {
	return baseResolver.IssueVariantBaseResolver(r.App, ctx, filter, first, after, &model.NodeParent{
		Parent:     obj,
		ParentName: model.IssueRepositoryNodeName,
	})
}

// Services is the resolver for the services field.
func (r *issueRepositoryResolver) Services(ctx context.Context, obj *model.IssueRepository, filter *model.ServiceFilter, first *int, after *string) (*model.ServiceConnection, error) {
	return baseResolver.ServiceBaseResolver(r.App, ctx, filter, first, after,
		&model.NodeParent{
			Parent:     obj,
			ParentName: model.IssueRepositoryNodeName,
		})
}

// IssueRepository returns graph.IssueRepositoryResolver implementation.
func (r *Resolver) IssueRepository() graph.IssueRepositoryResolver {
	return &issueRepositoryResolver{r}
}

type issueRepositoryResolver struct{ *Resolver }
