package auth

import (
	"crypto/subtle"

	"github.com/carlosxfelipe/gota"
	"order-flow/data"
)

// LoginRequest defines credentials for authentication.
type LoginRequest struct {
	Username string `json:"username" doc:"Username" example:"carlos" validate:"required" msg:"Username is required"`
	Password string `json:"password" doc:"User password" example:"carlos123" validate:"required" msg:"Password is required"`
}

// LoginResponse returns the generated JWT Bearer token.
type LoginResponse struct {
	Token string `json:"token" doc:"JWT Bearer token" example:"eyJhbGciOi..."`
}

// LoginEndpoint handles authentication (matches FastEndpoints Features/Auth/Login).
type LoginEndpoint struct {
	Store *data.Store
}

// Handle processes the login request and returns a signed JWT token.
func (e *LoginEndpoint) Handle(ctx *gota.Context, req LoginRequest) (LoginResponse, error) {
	user, ok := e.Store.GetUserByUsername(req.Username)
	if !ok || subtle.ConstantTimeCompare([]byte(user.Password), []byte(req.Password)) != 1 {
		return LoginResponse{}, gota.BadRequest("Invalid username or password.")
	}

	token, err := GenerateToken(user)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		Token: token,
	}, nil
}
