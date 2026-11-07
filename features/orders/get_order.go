package orders

import (
	"github.com/carlosxfelipe/gota"
	"order-flow-gota/data"
	"order-flow-gota/domain"
)

// GetOrderRequest defines the path parameter to retrieve an order.
type GetOrderRequest struct {
	OrderId string `param:"orderId" doc:"Unique order ID" example:"b1a2c3d4-0001-4000-8000-000000000001"`
}

// GetOrderResponse returns complete order details.
type GetOrderResponse struct {
	Id           string             `json:"id" doc:"Order ID" example:"b1a2c3d4-0001-4000-8000-000000000001"`
	CustomerName string             `json:"customerName" doc:"Customer name" example:"Carlos Felipe"`
	Status       string             `json:"status" doc:"Order status" example:"Created"`
	TotalAmount  float64            `json:"totalAmount" doc:"Accumulated total amount of items" example:"36.00"`
	Items        []domain.OrderItem `json:"items" doc:"List of order items"`
	Message      string             `json:"message" doc:"Status message" example:"Order retrieved successfully!"`
}

// GetOrderEndpoint handles order retrieval.
type GetOrderEndpoint struct {
	Store *data.Store
}

// Handle retrieves the order by ID.
func (e *GetOrderEndpoint) Handle(ctx *gota.Context, req GetOrderRequest) (GetOrderResponse, error) {
	order, ok := e.Store.GetOrder(req.OrderId)
	if !ok {
		return GetOrderResponse{}, gota.NotFound("Order not found")
	}
	if err := authorizeOrderAccess(ctx, order); err != nil {
		return GetOrderResponse{}, err
	}

	return GetOrderResponse{
		Id:           order.ID,
		CustomerName: order.CustomerName,
		Status:       string(order.Status),
		TotalAmount:  order.TotalAmount(),
		Items:        order.Items,
		Message:      "Order retrieved successfully!",
	}, nil
}
