package auth

import (
	"net/http"

	"github.com/carlosxfelipe/gota"
	"order-flow/data"
)

// RegisterPublicRoutes registers anonymous authentication endpoints (Login, Register).
func RegisterPublicRoutes(target gota.RouterTarget, store *data.Store) {
	gota.HandleEndpoint(target, http.MethodPost, "/api/auth/login", &LoginEndpoint{Store: store}).
		Summary("Authenticates a user").
		Description("Validates user credentials and returns a JWT Bearer token.").
		Tags("Auth")

	gota.HandleEndpoint(target, http.MethodPost, "/api/auth/register", &RegisterEndpoint{Store: store}).
		Summary("Register new account").
		Description("Creates a new user account and returns an initial JWT Bearer token.").
		Tags("Auth")
}

// RegisterProtectedRoutes registers endpoints that require JWT authentication (Me, Delete Account).
func RegisterProtectedRoutes(target gota.RouterTarget, store *data.Store) {
	gota.HandleEndpoint(target, http.MethodGet, "/api/auth/me", &MeEndpoint{Store: store}).
		Summary("Get current user profile").
		Description("Returns the profile of the currently authenticated user from token claims.").
		Tags("Auth")

	gota.HandleEndpoint(target, http.MethodDelete, "/api/auth/account", &DeleteAccountEndpoint{Store: store}).
		Summary("Delete user account").
		Description("Permanently removes the authenticated user's account from the system.").
		Tags("Auth")
}
