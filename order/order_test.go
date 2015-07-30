package order_test

import "github.com/marcusolsson/cqrs-example/order"

import "testing"

func TestPlaceOrder(t *testing.T) {
	repo := order.NewRepository(
		order.NewEventStore(),
	)

	handler := order.NewCommandHandler(repo)
	handler.Handle(order.Place{OrderID: "ABC123", Lines: []order.Line{{}}})

	o := repo.Load("ABC123")

	if o.ID != "ABC123" {
		t.Errorf("expected: %v, got: %v", "ABC123", o.ID)
	}

	if o.Status != order.StatusPlaced {
		t.Errorf("expected: %v, got: %v", order.StatusPlaced, o.Status)
	}
}

func TestActivateOrder(t *testing.T) {
	repo := order.NewRepository(
		order.NewEventStore(),
	)

	handler := order.NewCommandHandler(repo)
	handler.Handle(order.Place{OrderID: "ABC123", Lines: []order.Line{{}}})
	handler.Handle(order.Activate{OrderID: "ABC123"})

	o := repo.Load("ABC123")

	if o.ID != "ABC123" {
		t.Errorf("expected: %v, got: %v", "ABC123", o.ID)
	}

	if o.Status != order.StatusActivated {
		t.Errorf("expected: %v, got: %v", order.StatusActivated, o.Status)
	}
}
