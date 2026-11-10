package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/carlosxfelipe/gota"
	"order-flow/data"
	"order-flow/features/auth"
	"order-flow/features/orders"
)

// Offline TypeScript SDK code generation (e.g. for CI/CD or build scripts)
// for the Order Flow API.
//
// Running this script writes the full client SDK and TypeScript definitions to disk
// without needing to expose HTTP endpoints (/sdk.ts or /types.d.ts) in production.
//
// Usage:
//
//	go run ./codegen/main.go
//	go run ./codegen/main.go --ts-sdk=./dist/sdk.ts --ts-types=./dist/types.d.ts
//	go run ./codegen/main.go --stdout
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

	// 3. Process CLI flags if passed (e.g. for CI/CD or build scripts)
	if handled, err := app.RunCodegen(os.Args[1:]); handled {
		if err != nil {
			log.Fatalf("Codegen failed: %v", err)
		}
		return
	}

	// 4. Default behavior: export directly to <rootDir>/dist
	rootDir := findProjectRoot()
	distDir := filepath.Join(rootDir, "dist")

	sdkTarget := filepath.Join(distDir, "sdk.ts")
	if err := app.ExportTypeScriptSDK(sdkTarget); err != nil {
		log.Fatalf("Failed to export TypeScript SDK: %v", err)
	}
	log.Printf("✅ Successfully exported TypeScript SDK to %s", sdkTarget)

	dtsTarget := filepath.Join(distDir, "types.d.ts")
	if err := app.ExportTypeScript(dtsTarget); err != nil {
		log.Fatalf("Failed to export TypeScript definitions: %v", err)
	}
	log.Printf("✅ Successfully exported TypeScript definitions to %s", dtsTarget)
}

// findProjectRoot locates the repository root by searching upward for go.mod,
// ensuring files are placed in <root>/dist regardless of the working directory.
func findProjectRoot() string {
	// 1. Try walking up from the current working directory
	if wd, err := os.Getwd(); err == nil {
		for dir := wd; dir != "/" && dir != "."; dir = filepath.Dir(dir) {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return dir
			}
		}
	}

	// 2. Fallback: walk up from the source file location
	if _, filename, _, ok := runtime.Caller(0); ok {
		for dir := filepath.Dir(filename); dir != "/" && dir != "."; dir = filepath.Dir(dir) {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return dir
			}
		}
	}

	return "."
}
