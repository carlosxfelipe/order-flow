package orders

import (
	"github.com/carlosxfelipe/gota"
	"order-flow/data"
	"order-flow/domain"
	"order-flow/features/auth"
)

// CreateOrderRequest defines the input payload to create a new order.
type CreateOrderRequest struct {
	CustomerName    string                  `json:"customerName" doc:"Customer name" example:"Carlos Felipe" validate:"required,min=3" msg:"required:Customer name is required|min:Customer name must be at least 3 characters"`
	ShippingAddress *domain.ShippingAddress `json:"shippingAddress,omitempty" doc:"Optional delivery address with nested validation"`
}

// CreateOrderResponse defines the result of creating a new order.
type CreateOrderResponse struct {
	OrderId string `json:"orderId" doc:"Created order ID" example:"b1a2c3d4-0001-4000-8000-000000000001"`
	Message string `json:"message" doc:"Confirmation message" example:"Order created successfully!"`
}

// CreateOrderEndpoint handles order creation (matches FastEndpoints CreateOrderEndpoint).
type CreateOrderEndpoint struct {
	Store *data.Store
}

// Handle creates the order in initial 'Created' state.
func (e *CreateOrderEndpoint) Handle(ctx *gota.Context, req CreateOrderRequest) (CreateOrderResponse, error) {
	userID := 0
	if user, ok := auth.CurrentUser(ctx); ok {
		userID = user.UserID
	}

	order := domain.Order{
		ID:              domain.NewUUID(),
		UserID:          userID,
		CustomerName:    req.CustomerName,
		Status:          domain.StatusCreated,
		Items:           []domain.OrderItem{},
		ShippingAddress: req.ShippingAddress,
	}

	e.Store.SaveOrder(order)

	return CreateOrderResponse{
		OrderId: order.ID,
		Message: "Order created successfully!",
	}, nil
}
