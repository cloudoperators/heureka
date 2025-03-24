package event

import (
	"context"
	"github.com/cloudoperators/heureka/internal/event"
)

// EventRegistry is the central point for managing handlers for all kind of events
type EventRegistry interface {
	RegisterEventHandler(event.EventName, event.EventHandler)
	PushEvent(event.Event)
	Run(ctx context.Context)
	// todo: add shutdown
	// Shutdown stops the event registry
}
