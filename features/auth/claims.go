package auth

import (
	"time"

	"github.com/carlosxfelipe/gota"
	"order-flow-gota/domain"
)

// Secret is the HMAC signing key used for JWT tokens in this showcase (at least 32 bytes for HS256).
var Secret = []byte("GotaOrderFlowSecretKey2026!@#_SuperSecret")

// UserClaims represents the authenticated user payload inside the JWT.
type UserClaims struct {
	UserID   int    `json:"UserId"`
	Username string `json:"Username"`
	Role     string `json:"role"`
	Exp      int64  `json:"exp"`
}

// GenerateToken creates and signs a JWT Bearer token for the given user.
func GenerateToken(user domain.User) (string, error) {
	claims := UserClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		// Exp:      time.Now().Add(24 * time.Hour).Unix(),
		Exp: time.Now().AddDate(10, 0, 0).Unix(), // 10 years for showcase convenience
	}
	return gota.SignJWT(Secret, claims)
}

// CurrentUser retrieves the authenticated user claims from the Gota Context.
func CurrentUser(ctx *gota.Context) (*UserClaims, bool) {
	return gota.JWTClaims[*UserClaims](ctx)
}
