// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"sync"

	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/cloudoperators/heureka/internal/util"
)

type EventHandler interface {
	HandleEvent(database.Database, Event, openfga.Authorization)
}

type EventHandlerFunc func(database.Database, Event, openfga.Authorization)

func (f EventHandlerFunc) HandleEvent(db database.Database, e Event, authz openfga.Authorization) {
	f(db, e, authz)
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
	authz       openfga.Authorization
	cfg         *util.Config
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
	// Try to push the event without locking first
	select {
	case er.ch <- event:
		// Event successfully pushed to the channel
		return
	default:
		// Channel might be full, acquire lock and try again
		er.mu.Lock()
		defer er.mu.Unlock()

		// Try again with lock held
		select {
		case er.ch <- event:
			// Event successfully pushed to the channel
		default:
			// Channel is definitely full, increase its size
			newCap := cap(er.ch) * 2
			if newCap < 1024 {
				newCap = 1024 // Set a reasonable minimum for large batches
			}

			newCh := make(chan Event, newCap)

			// Push the current event to the new channel
			newCh <- event

			// Replace the channel
			oldCh := er.ch
			er.ch = newCh

			// Drain the old channel in a separate goroutine
			go func() {
				for e := range oldCh {
					er.ch <- e // Forward all events to the new channel
				}
			}()
		}
	}
}

func NewEventRegistry(db database.Database, authz openfga.Authorization) EventRegistry {
	initialBufferSize := 1024 // Start with a larger buffer
	workerCount := 4
	er := &eventRegistry{
		handlers:    make(map[EventName][]EventHandler),
		ch:          make(chan Event, initialBufferSize),
		db:          db,
		workerCount: workerCount,
		authz:       authz,
	}

	return er
}

func (er *eventRegistry) Run(ctx context.Context) {
	// Start workers
	for i := 0; i < er.workerCount; i++ {
		er.wg.Add(1)
		go er.worker(ctx)
	}

	// Wait for context cancellation
	go func() {
		<-ctx.Done() // Block until context is canceled
		er.mu.Lock()
		close(er.ch) // Close the channel
		er.mu.Unlock()
		er.wg.Wait() // Wait for all workers to finish
	}()
}

func (er *eventRegistry) worker(ctx context.Context) {
	defer er.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return // Exit when context is canceled
		case event, ok := <-er.ch:
			if !ok {
				return // Channel closed
			}

			er.processEvent(event)
		}
	}
}

func (er *eventRegistry) processEvent(event Event) {
	er.mu.Lock()
	handlers := er.handlers[event.Name()]
	er.mu.Unlock()

	for _, handler := range handlers {
		handler.HandleEvent(er.db, event, er.authz)
	}
}
