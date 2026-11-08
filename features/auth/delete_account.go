package auth

import (
	"github.com/carlosxfelipe/gota"
	"order-flow/data"
)

// DeleteAccountResponse confirms account deletion.
type DeleteAccountResponse struct {
	Message string `json:"message" doc:"Confirmation message" example:"Account deleted successfully."`
}

// DeleteAccountEndpoint handles self-service account deletion for the authenticated user.
type DeleteAccountEndpoint struct {
	Store *data.Store
}

// Handle deletes the authenticated user's account based on token claims.
func (e *DeleteAccountEndpoint) Handle(ctx *gota.Context, _ gota.Empty) (DeleteAccountResponse, error) {
	claims, ok := CurrentUser(ctx)
	if !ok {
		return DeleteAccountResponse{}, gota.Unauthorized("Unauthorized.")
	}

	if !e.Store.DeleteUser(claims.Username) {
		return DeleteAccountResponse{}, gota.NotFound("User not found or already deleted.")
	}

	return DeleteAccountResponse{
		Message: "Account deleted successfully.",
	}, nil
}
