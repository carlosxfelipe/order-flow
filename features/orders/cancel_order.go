package orders

import (
	"github.com/carlosxfelipe/gota"
	"order-flow/data"
	"order-flow/domain"
)

// CancelOrderRequest defines the path parameter for cancellation.
type CancelOrderRequest struct {
	OrderId string `param:"orderId" doc:"Order ID to cancel" example:"b1a2c3d4-0001-4000-8000-000000000001" validate:"required"`
}

// CancelOrderResponse defines the result of cancelling an order.
type CancelOrderResponse struct {
	Message string `json:"message" doc:"Status message" example:"Order cancelled successfully."`
}

// CancelOrderEndpoint handles order cancellation.
type CancelOrderEndpoint struct {
	Store *data.Store
}

// Handle cancels an order unless it was already shipped.
func (e *CancelOrderEndpoint) Handle(ctx *gota.Context, req CancelOrderRequest) (CancelOrderResponse, error) {
	order, ok := e.Store.GetOrder(req.OrderId)
	if !ok {
		return CancelOrderResponse{}, gota.NotFound("Order not found")
	}
	if err := authorizeOrderAccess(ctx, order); err != nil {
		return CancelOrderResponse{}, err
	}

	if order.Status == domain.StatusShipped {
		return CancelOrderResponse{}, gota.BadRequest("Shipped orders cannot be cancelled.")
	}

	order.Status = domain.StatusCancelled
	e.Store.SaveOrder(order)

	return CancelOrderResponse{
		Message: "Order cancelled successfully.",
	}, nil
}
