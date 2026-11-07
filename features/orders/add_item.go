package orders

import (
	"github.com/carlosxfelipe/gota"
	"order-flow-gota/data"
	"order-flow-gota/domain"
)

// AddItemRequest defines the route parameters and body to add an item to an order.
type AddItemRequest struct {
	OrderId     string  `param:"orderId" doc:"Order ID" example:"b1a2c3d4-0001-4000-8000-000000000001" validate:"required" msg:"Order ID is required"`
	ProductName string  `json:"productName" doc:"Item name" example:"Coca-Cola Can 350ml" validate:"required" msg:"Product name is required"`
	Price       float64 `json:"price" doc:"Unit price" example:"5.50" validate:"required,min=0.01,nowarn" msg:"required:Price is required|min:Price must be at least 0.01"`
	Quantity    int     `json:"quantity" doc:"Quantity" example:"2" validate:"required,min=1,nowarn" msg:"required:Quantity is required|min:Quantity must be at least 1"`
}

// AddItemResponse defines the result of adding an item.
type AddItemResponse struct {
	Message string `json:"message" doc:"Status message" example:"Item added to order successfully!"`
	ItemId  string `json:"itemId" doc:"Added item ID" example:"item_1"`
}

// AddItemEndpoint handles adding items to an open order.
type AddItemEndpoint struct {
	Store *data.Store
}

// Handle adds an item if the order is still in 'Created' state.
func (e *AddItemEndpoint) Handle(ctx *gota.Context, req AddItemRequest) (AddItemResponse, error) {
	order, ok := e.Store.GetOrder(req.OrderId)
	if !ok {
		return AddItemResponse{}, gota.NotFound("Order not found")
	}
	if err := authorizeOrderAccess(ctx, order); err != nil {
		return AddItemResponse{}, err
	}

	if order.Status != domain.StatusCreated {
		return AddItemResponse{}, gota.BadRequest("Cannot add items. Order is not in initial state.")
	}

	item := domain.OrderItem{
		ID:          domain.NewUUID(),
		ProductName: req.ProductName,
		Price:       req.Price,
		Quantity:    req.Quantity,
	}

	order.Items = append(order.Items, item)
	e.Store.SaveOrder(order)

	return AddItemResponse{
		Message: "Item added to order successfully!",
		ItemId:  item.ID,
	}, nil
}
