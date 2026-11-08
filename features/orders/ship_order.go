package orders

import (
	"github.com/carlosxfelipe/gota"
	"order-flow/data"
	"order-flow/domain"
)

// ShipOrderRequest defines the path parameter for shipping.
type ShipOrderRequest struct {
	OrderId string `param:"orderId" doc:"Order ID to ship" example:"b1a2c3d4-0002-4000-8000-000000000002" validate:"required"`
}

// ShipOrderResponse defines the result of shipping an order.
type ShipOrderResponse struct {
	Message string `json:"message" doc:"Status message" example:"Order shipped! Your Coca-Colas are on the way!"`
}

// ShipOrderEndpoint handles order shipping.
type ShipOrderEndpoint struct {
	Store *data.Store
}

// Handle transitions order to 'Shipped'.
func (e *ShipOrderEndpoint) Handle(ctx *gota.Context, req ShipOrderRequest) (ShipOrderResponse, error) {
	order, ok := e.Store.GetOrder(req.OrderId)
	if !ok {
		return ShipOrderResponse{}, gota.NotFound("Order not found")
	}
	if err := authorizeOrderAccess(ctx, order); err != nil {
		return ShipOrderResponse{}, err
	}

	if order.Status != domain.StatusPaid {
		return ShipOrderResponse{}, gota.BadRequest("Only paid orders can be shipped.")
	}

	order.Status = domain.StatusShipped
	e.Store.SaveOrder(order)

	return ShipOrderResponse{
		Message: "Order shipped! Your Coca-Colas are on the way!",
	}, nil
}
