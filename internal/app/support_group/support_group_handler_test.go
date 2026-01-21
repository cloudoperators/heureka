// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package support_group_test

import (
	"math"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	sg "github.com/cloudoperators/heureka/internal/app/support_group"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/cloudoperators/heureka/internal/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	mock "github.com/stretchr/testify/mock"
)

func TestSupportGroupHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Support Group Service Test Suite")
}

var handlerContext common.HandlerContext
var cfg *util.Config

var _ = BeforeSuite(func() {
	cfg = common.GetTestConfig()
	enableLogs := false
	authz := openfga.NewAuthorizationHandler(cfg, enableLogs)
	handlerContext = common.HandlerContext{
		Cache: cache.NewNoCache(),
		Authz: authz,
	}
})

func getSupportGroupFilter() *entity.SupportGroupFilter {
	return &entity.SupportGroupFilter{
		PaginatedX: entity.PaginatedX{
			First: nil,
			After: nil,
		},
	}
}

var _ = Describe("When listing SupportGroups", Label("app", "ListSupportGroups"), func() {
	var (
		er                  event.EventRegistry
		db                  *mocks.MockDatabase
		supportGroupHandler sg.SupportGroupHandler
		filter              *entity.SupportGroupFilter
		options             *entity.ListOptions
		order               []entity.Order
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getSupportGroupFilter()
		order = []entity.Order{}
		er = event.NewEventRegistry(db, handlerContext.Authz)
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{}, nil)
			db.On("CountSupportGroups", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			supportGroupHandler = sg.NewSupportGroupHandler(handlerContext)
			res, err := supportGroupHandler.ListSupportGroups(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(*res.TotalCount).Should(BeEquivalentTo(int64(1337)), "return correct Totalcount")
		})
	})

	When("the list option does include the PageInfo", func() {
		BeforeEach(func() {
			options.ShowPageInfo = true
		})
		DescribeTable("pagination information is correct", func(pageSize int, dbElements int, resElements int, hasNextPage bool) {
			filter.First = &pageSize
			var supportGroups []entity.SupportGroupResult
			for _, sg := range test.NNewFakeSupportGroupEntities(resElements) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithSupportGroup(order, sg))
				supportGroups = append(supportGroups, entity.SupportGroupResult{
					WithCursor:   entity.WithCursor{Value: cursor},
					SupportGroup: &sg,
				})
			}

			var cursors = lo.Map(supportGroups, func(s entity.SupportGroupResult, _ int) string {
				return s.Value
			})

			var i int64 = 0
			for len(cursors) < dbElements {
				i++
				supportGroup := test.NewFakeSupportGroupEntity()
				c, _ := mariadb.EncodeCursor(mariadb.WithSupportGroup([]entity.Order{}, supportGroup))
				cursors = append(cursors, c)
			}
			db.On("GetSupportGroups", filter, order).Return(supportGroups, nil)
			db.On("GetAllSupportGroupCursors", filter, order).Return(cursors, nil)
			supportGroupHandler = sg.NewSupportGroupHandler(handlerContext)
			res, err := supportGroupHandler.ListSupportGroups(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(*res.PageInfo.HasNextPage).To(BeEquivalentTo(hasNextPage), "correct hasNextPage indicator")
			Expect(len(res.Elements)).To(BeEquivalentTo(resElements))
			Expect(len(res.PageInfo.Pages)).To(BeEquivalentTo(int(math.Ceil(float64(dbElements)/float64(pageSize)))), "correct  number of pages")
		},
			Entry("When  pageSize is 1 and the database was returning 2 elements", 1, 2, 1, true),
			Entry("When  pageSize is 10 and the database was returning 9 elements", 10, 9, 9, false),
			Entry("When  pageSize is 10 and the database was returning 11 elements", 10, 11, 10, true),
		)
	})
})

var _ = Describe("When creating SupportGroup", Label("app", "CreateSupportGroup"), func() {
	var (
		er                  event.EventRegistry
		db                  *mocks.MockDatabase
		supportGroupHandler sg.SupportGroupHandler
		supportGroup        entity.SupportGroup
		filter              *entity.SupportGroupFilter
		order               []entity.Order
		p                   openfga.PermissionInput
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		supportGroup = test.NewFakeSupportGroupEntity()
		order = []entity.Order{}
		first := 10
		after := ""
		filter = &entity.SupportGroupFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}

		p = openfga.PermissionInput{
			UserType:   "role",
			UserId:     "0",
			ObjectId:   "",
			ObjectType: "support_group",
			Relation:   "role",
		}
		er = event.NewEventRegistry(db, handlerContext.Authz)
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("creates supportGroup", func() {
		filter.CCRN = []*string{&supportGroup.CCRN}
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("CreateSupportGroup", &supportGroup).Return(&supportGroup, nil)
		db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{}, nil)
		supportGroupHandler = sg.NewSupportGroupHandler(handlerContext)
		newSupportGroup, err := supportGroupHandler.CreateSupportGroup(&supportGroup)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newSupportGroup.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newSupportGroup.CCRN).To(BeEquivalentTo(supportGroup.CCRN))
		})
	})

	Context("when handling a CreateComponentInstanceEvent", func() {
		Context("when new component instance is created", func() {
			It("should add user resource relationship tuple in openfga", func() {
				sgFake := test.NewFakeSupportGroupEntity()
				createEvent := &sg.CreateSupportGroupEvent{
					SupportGroup: &sgFake,
				}

				// Use type assertion to convert a CreateServiceEvent into an Event
				var event event.Event = createEvent
				p.ObjectId = openfga.ObjectIdFromInt(createEvent.SupportGroup.Id)

				// Simulate event
				sg.OnSupportGroupCreateAuthz(db, event, handlerContext.Authz)

				ok, err := handlerContext.Authz.CheckPermission(p)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(ok).To(BeTrue(), "permission should be granted")
			})
		})
	})
})

var _ = Describe("When updating SupportGroup", Label("app", "UpdateSupportGroup"), func() {
	var (
		er                  event.EventRegistry
		db                  *mocks.MockDatabase
		supportGroupHandler sg.SupportGroupHandler
		supportGroup        entity.SupportGroupResult
		filter              *entity.SupportGroupFilter
		order               []entity.Order
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		supportGroup = test.NewFakeSupportGroupResult()
		first := 10
		after := ""
		order = []entity.Order{}
		filter = &entity.SupportGroupFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		er = event.NewEventRegistry(db, handlerContext.Authz)
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("updates supportGroup", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateSupportGroup", supportGroup.SupportGroup).Return(nil)
		supportGroupHandler = sg.NewSupportGroupHandler(handlerContext)
		supportGroup.CCRN = "Team Alone"
		filter.Id = []*int64{&supportGroup.Id}
		db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{supportGroup}, nil)
		updatedSupportGroup, err := supportGroupHandler.UpdateSupportGroup(supportGroup.SupportGroup)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedSupportGroup.CCRN).To(BeEquivalentTo(supportGroup.CCRN))
		})
	})
})

var _ = Describe("When deleting SupportGroup", Label("app", "DeleteSupportGroup"), func() {
	var (
		er                  event.EventRegistry
		db                  *mocks.MockDatabase
		supportGroupHandler sg.SupportGroupHandler
		id                  int64
		filter              *entity.SupportGroupFilter
		order               []entity.Order
		listOptions         *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		listOptions = entity.NewListOptions()
		id = 1
		first := 10
		after := ""
		order = []entity.Order{}
		filter = &entity.SupportGroupFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		er = event.NewEventRegistry(db, handlerContext.Authz)
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("deletes supportGroup", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("DeleteSupportGroup", id, mock.Anything).Return(nil)
		supportGroupHandler = sg.NewSupportGroupHandler(handlerContext)
		db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{}, nil)
		err := supportGroupHandler.DeleteSupportGroup(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		supportGroups, err := supportGroupHandler.ListSupportGroups(filter, listOptions)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(supportGroups.Elements).To(BeEmpty(), "no error should be thrown")
	})

	Context("when handling a DeleteSupportGroupEvent", func() {
		Context("when new support group is deleted", func() {
			It("should delete tuples related to that support group in openfga", func() {
				// Test OnSupportGroupDeleteAuthz against all possible relations
				sgFake := test.NewFakeSupportGroupEntity()
				deleteEvent := &sg.DeleteSupportGroupEvent{
					SupportGroupID: sgFake.Id,
				}
				objectId := openfga.ObjectIdFromInt(deleteEvent.SupportGroupID)
				userId := openfga.UserIdFromInt(deleteEvent.SupportGroupID)

				relations := []openfga.RelationInput{
					{ // user - support_group
						UserType:   openfga.TypeUser,
						UserId:     openfga.IDUser,
						ObjectId:   objectId,
						ObjectType: openfga.TypeSupportGroup,
						Relation:   openfga.RelMember,
					},
					{ // support_group - support_group
						UserType:   openfga.TypeSupportGroup,
						UserId:     userId,
						ObjectId:   objectId,
						ObjectType: openfga.TypeSupportGroup,
						Relation:   openfga.RelSupportGroup,
					},
					{ // role - support_group
						UserType:   openfga.TypeRole,
						UserId:     openfga.IDRole,
						ObjectId:   objectId,
						ObjectType: openfga.TypeSupportGroup,
						Relation:   openfga.RelRole,
					},
					{ // support_group - service
						UserType:   openfga.TypeSupportGroup,
						UserId:     userId,
						ObjectId:   openfga.IDService,
						ObjectType: openfga.TypeService,
						Relation:   openfga.RelSupportGroup,
					},
				}

				handlerContext.Authz.AddRelationBulk(relations)

				// get the number of relations before deletion
				relationsBefore, err := handlerContext.Authz.ListRelations(relations)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(relationsBefore).To(HaveLen(len(relations)), "all relations should exist before deletion")

				// check that relations were created
				for _, r := range relations {
					ok, err := handlerContext.Authz.CheckPermission(openfga.PermissionInput{
						UserType:   r.UserType,
						UserId:     r.UserId,
						ObjectType: r.ObjectType,
						ObjectId:   r.ObjectId,
						Relation:   r.Relation,
					})
					Expect(err).To(BeNil(), "no error should be thrown")
					Expect(ok).To(BeTrue(), "permission should be granted")
				}

				var event event.Event = deleteEvent
				// Simulate event
				sg.OnSupportGroupDeleteAuthz(db, event, handlerContext.Authz)

				// get the number of relations after deletion
				relationsAfter, err := handlerContext.Authz.ListRelations(relations)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(relationsAfter) < len(relationsBefore)).To(BeTrue(), "less relations after deletion")
				Expect(relationsAfter).To(BeEmpty(), "no relations should exist after deletion")

				// verify that relations were deleted
				for _, r := range relations {
					ok, err := handlerContext.Authz.CheckPermission(openfga.PermissionInput{
						UserType:   r.UserType,
						UserId:     r.UserId,
						ObjectType: r.ObjectType,
						ObjectId:   r.ObjectId,
						Relation:   r.Relation,
					})
					Expect(err).To(BeNil(), "no error should be thrown")
					Expect(ok).To(BeFalse(), "permission should NOT be granted")
				}
			})
		})
	})
})

var _ = Describe("When modifying relationship of Service and SupportGroup", Label("app", "ServiceSupportGroupRelationship"), func() {
	var (
		er                  event.EventRegistry
		db                  *mocks.MockDatabase
		supportGroupHandler sg.SupportGroupHandler
		service             entity.Service
		supportGroup        entity.SupportGroupResult
		filter              *entity.SupportGroupFilter
		order               []entity.Order
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		service = test.NewFakeServiceEntity()
		supportGroup = test.NewFakeSupportGroupResult()
		order = []entity.Order{}
		first := 10
		after := ""
		filter = &entity.SupportGroupFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
			Id: []*int64{&supportGroup.Id},
		}
		er = event.NewEventRegistry(db, handlerContext.Authz)
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("adds service to supportGroup", func() {
		db.On("AddServiceToSupportGroup", supportGroup.Id, service.Id).Return(nil)
		db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{supportGroup}, nil)
		supportGroupHandler = sg.NewSupportGroupHandler(handlerContext)
		supportGroup, err := supportGroupHandler.AddServiceToSupportGroup(supportGroup.Id, service.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(supportGroup).NotTo(BeNil(), "supportGroup should be returned")
	})

	Context("when handling an AddServiceToSupportGroupEvent", func() {
		It("should add the service-supportGroup relation tuple in openfga", func() {
			sgFake := test.NewFakeSupportGroupResult()
			serviceFake := test.NewFakeServiceEntity()
			addEvent := &sg.AddServiceToSupportGroupEvent{
				SupportGroupID: sgFake.Id,
				ServiceID:      serviceFake.Id,
			}
			supportGroupId := openfga.UserIdFromInt(addEvent.SupportGroupID)
			serviceId := openfga.ObjectIdFromInt(addEvent.ServiceID)

			rel := openfga.RelationInput{
				UserType:   openfga.TypeSupportGroup,
				UserId:     supportGroupId,
				ObjectType: openfga.TypeService,
				ObjectId:   serviceId,
				Relation:   openfga.RelSupportGroup,
			}

			var event event.Event = addEvent
			sg.OnAddServiceToSupportGroup(db, event, handlerContext.Authz)

			remaining, err := handlerContext.Authz.ListRelations([]openfga.RelationInput{rel})
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(remaining).NotTo(BeEmpty(), "relation should exist after addition")
		})
	})

	It("removes service from supportGroup", func() {
		db.On("RemoveServiceFromSupportGroup", supportGroup.Id, service.Id).Return(nil)
		db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{supportGroup}, nil)
		supportGroupHandler = sg.NewSupportGroupHandler(handlerContext)
		supportGroup, err := supportGroupHandler.RemoveServiceFromSupportGroup(supportGroup.Id, service.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(supportGroup).NotTo(BeNil(), "supportGroup should be returned")
	})

	Context("when handling a RemoveServiceFromSupportGroupEvent", func() {
		It("should remove the service-supportGroup relation tuple in openfga", func() {
			sgFake := test.NewFakeSupportGroupResult()
			serviceFake := test.NewFakeServiceEntity()
			removeEvent := &sg.RemoveServiceFromSupportGroupEvent{
				SupportGroupID: sgFake.Id,
				ServiceID:      serviceFake.Id,
			}
			supportGroupId := openfga.UserIdFromInt(removeEvent.SupportGroupID)
			serviceId := openfga.ObjectIdFromInt(removeEvent.ServiceID)

			rel := openfga.RelationInput{
				UserType:   openfga.TypeSupportGroup,
				UserId:     supportGroupId,
				ObjectType: openfga.TypeService,
				ObjectId:   serviceId,
				Relation:   openfga.TypeSupportGroup,
			}

			handlerContext.Authz.AddRelationBulk([]openfga.RelationInput{rel})

			var event event.Event = removeEvent
			sg.OnRemoveServiceFromSupportGroup(db, event, handlerContext.Authz)

			remaining, err := handlerContext.Authz.ListRelations([]openfga.RelationInput{rel})
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(remaining).To(BeEmpty(), "relation should not exist after removal")
		})
	})
})

var _ = Describe("When modifying relationship of User and SupportGroup", Label("app", "UserSupportGroupRelationship"), func() {
	var (
		er                  event.EventRegistry
		db                  *mocks.MockDatabase
		supportGroupHandler sg.SupportGroupHandler
		user                entity.User
		supportGroup        entity.SupportGroupResult
		filter              *entity.SupportGroupFilter
		order               []entity.Order
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		user = test.NewFakeUserEntity()
		supportGroup = test.NewFakeSupportGroupResult()
		first := 10
		after := ""
		order = []entity.Order{}
		filter = &entity.SupportGroupFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
			Id: []*int64{&supportGroup.Id},
		}
		er = event.NewEventRegistry(db, handlerContext.Authz)
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("adds user to supportGroup", func() {
		db.On("AddUserToSupportGroup", supportGroup.Id, user.Id).Return(nil)
		db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{supportGroup}, nil)
		supportGroupHandler = sg.NewSupportGroupHandler(handlerContext)
		supportGroup, err := supportGroupHandler.AddUserToSupportGroup(supportGroup.Id, user.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(supportGroup).NotTo(BeNil(), "supportGroup should be returned")
	})

	Context("when handling an AddUserToSupportGroupEvent", func() {
		It("should add the user-supportGroup relation tuple in openfga", func() {
			sgFake := test.NewFakeSupportGroupResult()
			userFake := test.NewFakeUserEntity()
			addEvent := &sg.AddUserToSupportGroupEvent{
				SupportGroupID: sgFake.Id,
				UserID:         userFake.Id,
			}
			supportGroupId := openfga.ObjectIdFromInt(addEvent.SupportGroupID)
			userId := openfga.UserIdFromInt(addEvent.UserID)

			rel := openfga.RelationInput{
				UserType:   openfga.TypeUser,
				UserId:     userId,
				ObjectType: openfga.TypeSupportGroup,
				ObjectId:   supportGroupId,
				Relation:   openfga.RelMember,
			}

			var event event.Event = addEvent
			sg.OnAddUserToSupportGroup(db, event, handlerContext.Authz)

			remaining, err := handlerContext.Authz.ListRelations([]openfga.RelationInput{rel})
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(remaining).NotTo(BeEmpty(), "relation should exist after addition")
		})
	})

	It("removes user from supportGroup", func() {
		db.On("RemoveUserFromSupportGroup", supportGroup.Id, user.Id).Return(nil)
		db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{supportGroup}, nil)
		supportGroupHandler = sg.NewSupportGroupHandler(handlerContext)
		supportGroup, err := supportGroupHandler.RemoveUserFromSupportGroup(supportGroup.Id, user.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(supportGroup).NotTo(BeNil(), "supportGroup should be returned")
	})

	Context("when handling a RemoveUserFromSupportGroupEvent", func() {
		It("should remove the user-supportGroup relation tuple in openfga", func() {
			sgFake := test.NewFakeSupportGroupResult()
			userFake := test.NewFakeUserEntity()
			removeEvent := &sg.RemoveUserFromSupportGroupEvent{
				SupportGroupID: sgFake.Id,
				UserID:         userFake.Id,
			}
			supportGroupId := openfga.ObjectIdFromInt(removeEvent.SupportGroupID)
			userId := openfga.UserIdFromInt(removeEvent.UserID)

			rel := openfga.RelationInput{
				UserType:   openfga.TypeUser,
				UserId:     userId,
				ObjectType: openfga.TypeSupportGroup,
				ObjectId:   supportGroupId,
				Relation:   openfga.RelMember,
			}
			// Bulk add instead of single add
			handlerContext.Authz.AddRelationBulk([]openfga.RelationInput{rel})

			var event event.Event = removeEvent
			sg.OnRemoveUserFromSupportGroup(db, event, handlerContext.Authz)

			remaining, err := handlerContext.Authz.ListRelations([]openfga.RelationInput{rel})
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(remaining).To(BeEmpty(), "relation should not exist after removal")
		})
	})
})

var _ = Describe("When listing supportGroupCcrns", Label("app", "ListSupportGroupCcrns"), func() {
	var (
		er                  event.EventRegistry
		db                  *mocks.MockDatabase
		supportGroupHandler sg.SupportGroupHandler
		filter              *entity.SupportGroupFilter
		options             *entity.ListOptions
		ccrn                string
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getSupportGroupFilter()
		ccrn = "src"
		er = event.NewEventRegistry(db, handlerContext.Authz)
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	When("no filters are used", func() {

		BeforeEach(func() {
			db.On("GetSupportGroupCcrns", filter).Return([]string{}, nil)
		})

		It("it return the results", func() {
			supportGroupHandler = sg.NewSupportGroupHandler(handlerContext)
			res, err := supportGroupHandler.ListSupportGroupCcrns(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(BeEmpty(), "return correct result")
		})
	})
	When("specific supportGroupCcrns filter is applied", func() {
		BeforeEach(func() {
			filter = &entity.SupportGroupFilter{
				CCRN: []*string{&ccrn},
			}

			db.On("GetSupportGroupCcrns", filter).Return([]string{ccrn}, nil)
		})
		It("returns filtered userGroups according to the service type", func() {
			supportGroupHandler = sg.NewSupportGroupHandler(handlerContext)
			res, err := supportGroupHandler.ListSupportGroupCcrns(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(ConsistOf(ccrn), "should only consist of supportGroup")
		})
	})
})
