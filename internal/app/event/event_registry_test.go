// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.wdf.sap.corp/cc/heureka/internal/database"
	"github.wdf.sap.corp/cc/heureka/internal/mocks"
)

func TestEventRegistry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Event Registry Test Suite")
}

type TestEvent struct {
	name string
}

func NewTestEvent(name string) Event {
	return &TestEvent{name: name}
}

func (e *TestEvent) Name() EventName {
	return EventName(e.name)
}

var _ = Describe("EventRegistry", Label("app", "event", "EventRegistry"), func() {
	var (
		er     EventRegistry
		db     *mocks.MockDatabase
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = NewEventRegistry(db)
		ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		cancel()
	})

	It("should register and handle events", func() {
		var eventHandled int32
		handler := func(db database.Database, e Event) {
			atomic.AddInt32(&eventHandled, 1)
		}

		er.Run(ctx)
		er.RegisterEventHandler("test_event", EventHandlerFunc(handler))
		er.PushEvent(NewTestEvent("test_event"))

		time.Sleep(25 * time.Millisecond) // Allow some time for the event to be processed

		Expect(atomic.LoadInt32(&eventHandled)).To(Equal(int32(1)))
	})

	It("should handle multiple events", func() {
		var eventHandled int32
		handler := func(db database.Database, e Event) {
			atomic.AddInt32(&eventHandled, 1)
		}
		er.Run(ctx)
		er.RegisterEventHandler("test_event", EventHandlerFunc(handler))
		er.PushEvent(NewTestEvent("test_event"))
		er.PushEvent(NewTestEvent("test_event"))

		time.Sleep(25 * time.Millisecond) // Allow some time for the events to be processed

		Expect(atomic.LoadInt32(&eventHandled)).To(Equal(int32(2)))
	})

	It("should handle multiple times when multiple handlers are registered", func() {
		var eventHandled int32
		handler := func(db database.Database, e Event) {
			atomic.AddInt32(&eventHandled, 1)
		}

		er.Run(ctx)
		er.RegisterEventHandler("test_event", EventHandlerFunc(handler))
		er.RegisterEventHandler("test_event", EventHandlerFunc(handler))
		er.RegisterEventHandler("test_event", EventHandlerFunc(handler))
		er.PushEvent(NewTestEvent("test_event"))
		er.PushEvent(NewTestEvent("test_event"))

		time.Sleep(25 * time.Millisecond) // Allow some time for the events to be processed

		Expect(atomic.LoadInt32(&eventHandled)).To(Equal(int32(6)))
	})

	It("should stop processing on context cancel", func() {
		var eventHandled int32
		handler := func(db database.Database, e Event) {
			atomic.AddInt32(&eventHandled, 1)
		}

		er.Run(ctx)
		er.RegisterEventHandler("test_event", EventHandlerFunc(handler))
		er.PushEvent(NewTestEvent("test_event"))

		time.Sleep(10 * time.Millisecond) // Allow some time for the event to be processed
		cancel()                          // Cancel the context to stop processing
		time.Sleep(10 * time.Millisecond) // Allow some time for the cancellation to take effect

		Expect(atomic.LoadInt32(&eventHandled)).To(Equal(int32(1)))
	})
})
