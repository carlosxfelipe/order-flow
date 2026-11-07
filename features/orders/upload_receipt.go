package orders

import (
	"mime/multipart"

	"github.com/carlosxfelipe/gota"
	"order-flow-gota/data"
)

// UploadReceiptRequest defines the multipart/form-data payload for attaching a payment receipt.
type UploadReceiptRequest struct {
	OrderId string                `param:"orderId" doc:"Order ID" example:"b1a2c3d4-0001-4000-8000-000000000001" validate:"required"`
	Receipt *multipart.FileHeader `form:"receipt" doc:"Payment receipt file (PDF, PNG, JPG)" validate:"required"`
	Notes   string                `form:"notes" doc:"Optional notes about the receipt" example:"Paid at store counter"`
}

// UploadReceiptResponse describes the uploaded receipt confirmation.
type UploadReceiptResponse struct {
	OrderId  string `json:"orderId" doc:"Order ID" example:"b1a2c3d4-0001-4000-8000-000000000001"`
	Filename string `json:"filename" doc:"Uploaded file name" example:"receipt.pdf"`
	Size     int64  `json:"size" doc:"Uploaded file size in bytes" example:"10240"`
	Notes    string `json:"notes,omitempty" doc:"Notes provided with upload" example:"Paid at store counter"`
	Message  string `json:"message" doc:"Confirmation message" example:"Receipt attached successfully!"`
}

// UploadReceiptEndpoint handles attaching payment receipt files to an order.
type UploadReceiptEndpoint struct {
	Store *data.Store
}

// Handle binds the multipart form file, updates the order record, and returns the confirmation.
func (e *UploadReceiptEndpoint) Handle(ctx *gota.Context, req UploadReceiptRequest) (UploadReceiptResponse, error) {
	order, ok := e.Store.GetOrder(req.OrderId)
	if !ok {
		return UploadReceiptResponse{}, gota.NotFound("Order not found")
	}
	if err := authorizeOrderAccess(ctx, order); err != nil {
		return UploadReceiptResponse{}, err
	}

	safeFilename := gota.SanitizeFilename(req.Receipt.Filename)
	if safeFilename == "" {
		return UploadReceiptResponse{}, gota.BadRequest("Invalid receipt filename")
	}

	order.ReceiptFilename = safeFilename
	e.Store.SaveOrder(order)

	return UploadReceiptResponse{
		OrderId:  order.ID,
		Filename: safeFilename,
		Size:     req.Receipt.Size,
		Notes:    req.Notes,
		Message:  "Receipt attached successfully!",
	}, nil
}
