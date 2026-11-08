package auth

import (
	"github.com/carlosxfelipe/gota"
	"order-flow/data"
	"order-flow/domain"
)

// MeEndpoint retrieves the profile of the currently authenticated user.
type MeEndpoint struct {
	Store *data.Store
}

// Handle extracts the claims from context and returns user information.
func (e *MeEndpoint) Handle(ctx *gota.Context, _ gota.Empty) (domain.User, error) {
	claims, ok := CurrentUser(ctx)
	if !ok {
		return domain.User{}, gota.Unauthorized("Unauthorized.")
	}

	user, ok := e.Store.GetUserByUsername(claims.Username)
	if !ok {
		return domain.User{}, gota.NotFound("User not found.")
	}

	return user, nil
}
