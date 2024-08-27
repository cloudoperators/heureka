package app

func (h *HeurekaApp) MatchOnComponentVersionAddedToComponentInstance(componentVersionId int64, componentInstanceId int64) error {
	//searvices, err := h.database.GetServices(&entity.ServiceFilter{
	//	Paginated:           entity.MaxPaginated(),
	//	ComponentInstanceId: []*int64{&componentInstanceId},
	//})
	//
	////@todo add proper error handling
	//if err != nil {
	//	return err
	//
	//
	//serviceIds := lo.Map(searvices, func(service entity.Service, _ int) *int64 { return &service.Id })
	//serviceNames := lo.Map(searvices, func(service entity.Service, _ int) *string { return &service.Name })
	//
	//issues, err := h.database.GetIssues(&entity.IssueFilter{
	//	Paginated:          entity.MaxPaginated(),
	//	ComponentVersionId: []*int64{&componentVersionId},
	//	ServiceName:        serviceNames, //@todo check whatever we need this at this point ? Do we have a service assignment by default?
	//})
	////@todo add proper error handling
	//if err != nil {
	//	return err
	//}
	//
	//for _, issue := range issues {
	//	h.database.GetIssueVariants(&entity.IssueVariantFilter{)
	//	}
	//	return nil
	//}
	//
	//func(h *HeurekaApp) MatchOnComponentVersionAddedToIssue(componentVersionId
	//int64, issueId
	//int64) error{
	return nil
}
