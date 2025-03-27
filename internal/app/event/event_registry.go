// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"sync"

	"github.com/cloudoperators/heureka/internal/database"
)

type EventHandler interface {
	HandleEvent(database.Database, Event)
}

type EventHandlerFunc func(database.Database, Event)

func (f EventHandlerFunc) HandleEvent(db database.Database, e Event) {
	f(db, e)
}

type EventRegistry interface {
	RegisterEventHandler(EventName, EventHandler)
	PushEvent(Event)
	Run(ctx context.Context)
}

type eventRegistry struct {
	handlers    map[EventName][]EventHandler
	db          database.Database
	ch          chan Event
	wg          sync.WaitGroup
	mu          sync.Mutex
	workerCount int
}

func (er *eventRegistry) RegisterEventHandler(event EventName, handler EventHandler) {
	er.mu.Lock()
	defer er.mu.Unlock()

	if er.handlers == nil {
		er.handlers = make(map[EventName][]EventHandler)
	}
	er.handlers[event] = append(er.handlers[event], handler)
}

func (er *eventRegistry) PushEvent(event Event) {
	er.mu.Lock()
	defer er.mu.Unlock()

	select {
	case er.ch <- event:
		// Event successfully pushed to the channel
	default:
		// Channel is full, create a new channel with twice as large buffer
		newCh := make(chan Event, cap(er.ch)*2)
		go func(oldCh, newCh chan Event) {
			for e := range oldCh {
				newCh <- e
			}
			close(newCh)
		}(er.ch, newCh)
		er.ch = newCh
		er.reinitWorkers()
		er.ch <- event
	}
}

func (er *eventRegistry) reinitWorkers() {
	for i := 0; i < er.workerCount; i++ {
		er.wg.Add(1)
		go er.worker()
	}
}

func NewEventRegistry(db database.Database) EventRegistry {
	bufferSize := 500
	workerCount := 4
	er := &eventRegistry{
		handlers:    make(map[EventName][]EventHandler),
		ch:          make(chan Event, bufferSize),
		db:          db,
		workerCount: workerCount,
	}

	for i := 0; i < workerCount; i++ {
		er.wg.Add(1)
		go er.worker()
	}

	return er
}

func (er *eventRegistry) Run(ctx context.Context) {
	go func() {
		<-ctx.Done() // Block until context is canceled
		close(er.ch) // Close the channel
		er.wg.Wait() // Wait for all workers to finish
	}()
}

func (er *eventRegistry) worker() {
	defer er.wg.Done()
	for event := range er.ch { // Infinite loop to listen for events
		for _, handler := range er.handlers[event.Name()] {
			handler.HandleEvent(er.db, event)
		}
	}
}
