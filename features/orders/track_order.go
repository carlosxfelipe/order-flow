package orders

import (
	"github.com/carlosxfelipe/gota"
	"order-flow-gota/data"
)

// TrackOrderRequest defines the order ID to track via Server-Sent Events.
type TrackOrderRequest struct {
	OrderId string `param:"orderId" doc:"Order ID to track" example:"b1a2c3d4-0001-4000-8000-000000000001" validate:"required"`
}

// TrackOrderEndpoint streams real-time status updates for an order using SSE.
type TrackOrderEndpoint struct {
	Store *data.Store
}

// Handle emits status events for the requested order.
func (e *TrackOrderEndpoint) Handle(ctx *gota.Context, req TrackOrderRequest) (gota.SSEStream, error) {
	order, ok := e.Store.GetOrder(req.OrderId)
	if !ok {
		return gota.SSEStream{}, gota.NotFound("Order not found")
	}
	if err := authorizeOrderAccess(ctx, order); err != nil {
		return gota.SSEStream{}, err
	}

	events := make(chan gota.SSEvent, 2)
	events <- gota.SSEvent{
		Event: "order_status",
		Data: gota.Map{
			"orderId": order.ID,
			"status":  string(order.Status),
		},
	}
	events <- gota.SSEvent{
		Event: "order_summary",
		Data: gota.Map{
			"orderId": order.ID,
			"items":   len(order.Items),
			"total":   order.TotalAmount(),
		},
	}
	close(events)

	return gota.SSEStream{Events: events}, nil
}
