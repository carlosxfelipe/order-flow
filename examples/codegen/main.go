package main

import (
	"log"

	"github.com/carlosxfelipe/gota"
	"order-flow/data"
	"order-flow/features/auth"
	"order-flow/features/orders"
)

// Example demonstrating offline TypeScript SDK code generation (e.g. for CI/CD or build scripts)
// for the Order Flow API.
//
// Running this script writes the full client SDK and TypeScript definitions to disk
// without needing to expose HTTP endpoints (/sdk.ts or /types.d.ts) in production.
//
// Usage:
//
//	go run ./examples/codegen/main.go
func main() {
	// 1. Initialize the Order Flow app with HTTP TypeScript routes disabled
	app := gota.New(gota.Config{
		Title:            "Order Flow API",
		Version:          "1.0.0",
		Description:      "Order management API with REPR pattern and zero external dependencies.",
		TypeScriptURL:    "-",
		TypeScriptSDKURL: "-",
	})

	// 2. Register Order Flow models and endpoints
	store := data.NewStore()

	// Public Auth Endpoints (Anonymous: Login, Register)
	auth.RegisterPublicRoutes(app, store)

	// Protected Endpoints (Secured with built-in gota.JWT middleware)
	protectedGroup := app.Group("")
	protectedGroup.Use(gota.JWT(auth.Secret, gota.JWTOptions{
		ClaimsFactory: func() any { return &auth.UserClaims{} },
		RequireExp:    true,
	}))
	auth.RegisterProtectedRoutes(protectedGroup, store)
	orders.RegisterRoutes(protectedGroup, store)

	// 3. Export TypeScript SDK and Definitions directly to disk
	sdkTarget := "./dist/sdk.ts"
	if err := app.ExportTypeScriptSDK(sdkTarget); err != nil {
		log.Fatalf("Failed to export TypeScript SDK: %v", err)
	}
	log.Printf("✅ Successfully exported TypeScript SDK to %s", sdkTarget)

	dtsTarget := "./dist/types.d.ts"
	if err := app.ExportTypeScript(dtsTarget); err != nil {
		log.Fatalf("Failed to export TypeScript definitions: %v", err)
	}
	log.Printf("✅ Successfully exported TypeScript definitions to %s", dtsTarget)
}
