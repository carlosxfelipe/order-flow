package orders

import (
	"net/http"

	"github.com/carlosxfelipe/gota"
	"order-flow/data"
	"order-flow/domain"
	"order-flow/features/auth"
)

// authorizeOrderAccess ensures a Customer can only access or mutate their own orders,
// while Admins have full access to all orders.
func authorizeOrderAccess(ctx *gota.Context, order domain.Order) error {
	user, ok := auth.CurrentUser(ctx)
	if !ok {
		return nil
	}
	if user.Role == "Admin" || order.UserID == 0 || order.UserID == user.UserID {
		return nil
	}
	return gota.Forbidden("You do not have permission to access this order.")
}

// RegisterRoutes registers all Order REPR endpoints onto the router target (App or RouteGroup).
func RegisterRoutes(target gota.RouterTarget, store *data.Store) {
	// 1. List Orders
	gota.HandleEndpoint(target, http.MethodGet, "/api/orders", &ListOrdersEndpoint{Store: store}).
		Summary("List orders").
		Description("Lists orders. Customers only see their own orders; Admins see all orders.").
		Tags("Orders")

	// 2. Create Order
	gota.HandleEndpoint(target, http.MethodPost, "/api/orders", &CreateOrderEndpoint{Store: store}).
		Summary("Create order").
		Description("Creates a new order in initial 'Created' state.").
		Tags("Orders")

	// 3. Get Order by ID
	gota.HandleEndpoint(target, http.MethodGet, "/api/orders/:orderId", &GetOrderEndpoint{Store: store}).
		Summary("Get order by ID").
		Description("Retrieves complete order details including all items and total amount.").
		Tags("Orders")

	// 4. Add Item to Order
	gota.HandleEndpoint(target, http.MethodPost, "/api/orders/:orderId/items", &AddItemEndpoint{Store: store}).
		Summary("Add item to order").
		Description("Adds a new line item to an order still in 'Created' state.").
		Tags("Orders")

	// 5. Remove Item from Order
	gota.HandleEndpoint(target, http.MethodDelete, "/api/orders/:orderId/items/:itemId", &RemoveItemEndpoint{Store: store}).
		Summary("Remove item from order").
		Description("Removes a line item from an order still in 'Created' state.").
		Tags("Orders")

	// 6. Pay Order
	gota.HandleEndpoint(target, http.MethodPost, "/api/orders/:orderId/pay", &PayOrderEndpoint{Store: store}).
		Summary("Pay order").
		Description("Transitions an order with items from 'Created' to 'Paid' state.").
		Tags("Orders")

	// 7. Ship Order
	gota.HandleEndpoint(target, http.MethodPost, "/api/orders/:orderId/ship", &ShipOrderEndpoint{Store: store}).
		Summary("Ship order").
		Description("Transitions a 'Paid' order to 'Shipped' state.").
		Tags("Orders")

	// 8. Cancel Order
	gota.HandleEndpoint(target, http.MethodPost, "/api/orders/:orderId/cancel", &CancelOrderEndpoint{Store: store}).
		Summary("Cancel order").
		Description("Cancels an order unless it was already shipped.").
		Tags("Orders")

	// 9. Upload Order Receipt (Multipart Form File Upload)
	gota.HandleEndpoint(target, http.MethodPost, "/api/orders/:orderId/receipt", &UploadReceiptEndpoint{Store: store}).
		Summary("Upload order receipt").
		Description("Attaches a payment receipt or invoice document (multipart/form-data) to the order.").
		Tags("Orders")

	// 10. Track Order (Server-Sent Events Streaming)
	gota.HandleEndpoint(target, http.MethodGet, "/api/orders/:orderId/track", &TrackOrderEndpoint{Store: store}).
		Summary("Track order status stream").
		Description("Streams real-time Server-Sent Events (SSE) representing order processing stages.").
		Tags("Orders")
}
