package orders

import (
	"github.com/carlosxfelipe/gota"
	"order-flow-gota/data"
	"order-flow-gota/domain"
)

// RemoveItemRequest defines the path parameters to delete an item.
type RemoveItemRequest struct {
	OrderId string `param:"orderId" doc:"Order ID" example:"b1a2c3d4-0001-4000-8000-000000000001" validate:"required"`
	ItemId  string `param:"itemId" doc:"Item ID to remove" example:"item_101" validate:"required"`
}

// RemoveItemResponse defines the result of removing an item.
type RemoveItemResponse struct {
	Message string `json:"message" doc:"Status message" example:"Item removed successfully!"`
}

// RemoveItemEndpoint handles item removal.
type RemoveItemEndpoint struct {
	Store *data.Store
}

// Handle removes an item from an open order.
func (e *RemoveItemEndpoint) Handle(ctx *gota.Context, req RemoveItemRequest) (RemoveItemResponse, error) {
	order, ok := e.Store.GetOrder(req.OrderId)
	if !ok {
		return RemoveItemResponse{}, gota.NotFound("Order not found")
	}
	if err := authorizeOrderAccess(ctx, order); err != nil {
		return RemoveItemResponse{}, err
	}

	if order.Status != domain.StatusCreated {
		return RemoveItemResponse{}, gota.BadRequest("Cannot remove items. Order is no longer in created state.")
	}

	foundIdx := -1
	for idx, it := range order.Items {
		if it.ID == req.ItemId {
			foundIdx = idx
			break
		}
	}

	if foundIdx == -1 {
		return RemoveItemResponse{}, gota.NotFound("Item not found in order")
	}

	order.Items = append(order.Items[:foundIdx], order.Items[foundIdx+1:]...)
	e.Store.SaveOrder(order)

	return RemoveItemResponse{
		Message: "Item removed successfully!",
	}, nil
}
