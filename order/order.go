package order

import (
	"errors"
	"log"
)

var (
	errAlreadyPlaced  = errors.New("order has already been placed")
	errEmptyOrderLine = errors.New("empty order line")
	errOrderNotFound  = errors.New("order was not found")
)

// Status represents the order status.
type Status int

// All possible order statuses.
const (
	StatusPlaced Status = iota
	StatusActivated
)

// Order is the aggregate root.
type Order struct {
	ID     string
	Status Status

	uncommitted []Event
}

// Place places the order by assigning order lines if not already placed.
func (o *Order) Place(orderLines []Line) error {
	if o.ID == "" {
		return errAlreadyPlaced
	}

	if len(orderLines) == 0 {
		return errEmptyOrderLine
	}

	apply(o, Placed{OrderID: o.ID}, true)

	return nil
}

// Activate activates the order.
func (o *Order) Activate() {
	if o.Status == StatusPlaced {
		apply(o, Activated{OrderID: o.ID}, true)
	}
}

// Event is the interface for all domain events.
type Event interface {
	ID() string
}

// Placed represents the event when an order was placed.
type Placed struct {
	OrderID string
}

// ID returns the identifier of the aggregate root, i.e. the order.
func (e Placed) ID() string {
	return e.OrderID
}

// Activated represents the event when an order was activated.
type Activated struct {
	OrderID string
}

// ID returns the identifier of the order (aggregate root).
func (e Activated) ID() string {
	return e.OrderID
}

// Line represents an order line.
type Line struct {
}

// Place represents a command for placing an order.
type Place struct {
	OrderID string
	Lines   []Line
}

// Activate represents a command for activating an order.
type Activate struct {
	OrderID string
}

// loadFromHistory builds a order from a series of events.
func loadFromHistory(events []Event) Order {
	var o Order
	for _, e := range events {
		apply(&o, e, false)
	}
	return o
}

// apply updates meta data of the order and stores the new event after it has been handled.
func apply(o *Order, e Event, isNew bool) {
	o.ID = e.ID()

	handle(o, e)

	if isNew {
		o.uncommitted = append(o.uncommitted, e)
	}
}

// handle updates the state of the order for every events.
func handle(o *Order, e Event) {
	switch e.(type) {
	case Activated:
		o.Status = StatusActivated
	case Placed:
		o.Status = StatusPlaced
	}
}

// EventStore defines the operations of a event store.
type EventStore interface {
	Save(id string, events []Event)
	Load(id string) ([]Event, error)
}

type eventStore struct {
	events []Event
}

func (s *eventStore) Save(id string, events []Event) {
	s.events = append(s.events, events...)
}

func (s *eventStore) Load(id string) ([]Event, error) {
	var result []Event
	for _, e := range s.events {
		if e.ID() == id {
			result = append(result, e)
		}
	}

	if len(result) == 0 {
		return nil, errOrderNotFound
	}

	return result, nil
}

// NewEventStore returns a new instance of the default event store.
func NewEventStore() EventStore {
	return &eventStore{}
}

// Repository ...
type Repository interface {
	Save(Order)
	Load(string) Order
}

type defaultRepository struct {
	Store EventStore
}

// Save ...
func (r *defaultRepository) Save(order Order) {
	if len(order.uncommitted) > 0 {
		r.Store.Save(order.ID, order.uncommitted)
	}
}

// Load ...
func (r *defaultRepository) Load(id string) Order {
	events, err := r.Store.Load(id)
	if err != nil {
		return Order{}
	}

	return loadFromHistory(events)
}

// NewRepository returns a new instance of the default repository.
func NewRepository(store EventStore) Repository {
	return &defaultRepository{
		Store: store,
	}
}

// CommandHandler defines an interface for handling order commands.
type CommandHandler interface {
	Handle(c interface{})
}

type commandHandler struct {
	Repository Repository
}

func (h *commandHandler) Handle(c interface{}) {
	switch cmd := c.(type) {
	case Place:
		order := Order{
			ID: cmd.OrderID,
		}
		if err := order.Place(cmd.Lines); err != nil {
			log.Println(err)
		}
		h.Repository.Save(order)
	case Activate:
		order := h.Repository.Load(cmd.OrderID)
		order.Activate()
		h.Repository.Save(order)
	}
}

// NewCommandHandler returns a new instance of the default command handler.
func NewCommandHandler(r Repository) CommandHandler {
	return &commandHandler{
		Repository: r,
	}
}
