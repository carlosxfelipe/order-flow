package auth

import (
	"github.com/carlosxfelipe/gota"
	"order-flow/data"
)

// RegisterRequest defines input for account registration.
type RegisterRequest struct {
	Username string `json:"username" doc:"Username" example:"novousuario" validate:"required,min=3" msg:"required:Username is required|min:Username must be at least 3 characters long"`
	Password string `json:"password" doc:"User password" example:"senha123" validate:"required,min=6,nowarn" msg:"required:Password is required|min:Password must be at least 6 characters long"`
}

// RegisterResponse returns the created user info along with an initial JWT token.
type RegisterResponse struct {
	ID       int    `json:"id" doc:"User ID" example:"5"`
	Username string `json:"username" doc:"Username" example:"novousuario"`
	Role     string `json:"role" doc:"User role" example:"Customer"`
	Token    string `json:"token" doc:"JWT Bearer token" example:"eyJhbGciOi..."`
}

// RegisterEndpoint handles user account registration.
type RegisterEndpoint struct {
	Store *data.Store
}

// Handle creates a new user account and returns a signed JWT token.
func (e *RegisterEndpoint) Handle(ctx *gota.Context, req RegisterRequest) (RegisterResponse, error) {
	// The initial role is always assigned by the server; clients must never choose it.
	const defaultRole = "Customer"

	user, ok := e.Store.CreateUser(req.Username, req.Password, defaultRole)
	if !ok {
		return RegisterResponse{}, gota.BadRequest("Username is already taken.")
	}

	token, err := GenerateToken(user)
	if err != nil {
		return RegisterResponse{}, err
	}

	return RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		Token:    token,
	}, nil
}
