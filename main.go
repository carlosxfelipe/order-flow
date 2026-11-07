package main

import (
	"log"
	"os"
	"time"

	"github.com/carlosxfelipe/gota"
	"order-flow-gota/data"
	"order-flow-gota/features/auth"
	"order-flow-gota/features/orders"
)

func main() {
	isProd := os.Getenv("ENV") == "production"

	// In production, disable runtime HTTP TypeScript endpoints (use offline export instead)
	tsURL := "/types.d.ts"
	tsSDKURL := "/sdk.ts"
	if isProd {
		tsURL = "-"
		tsSDKURL = "-"
	}

	// Initializes Gota with built-in OpenAPI 3.1, Scalar UI, and production timeouts
	app := gota.New(gota.Config{
		Title:                 "Order Flow API",
		Version:               "1.0.0",
		Description:           "Order management API with REPR pattern and zero external dependencies.",
		TypeScriptURL:         tsURL,
		TypeScriptSDKURL:      tsSDKURL,
		ScalarTheme:           gota.ThemeFastify,
		ScalarLayout:          "classic",
		DocsCSPEnforce:        true,
		DisallowUnknownFields: true,
		MaxBodyBytes:          10 << 20, // 10MB request body protection
		ReadTimeout:           15 * time.Second,
		WriteTimeout:          15 * time.Second,
		IdleTimeout:           60 * time.Second,
	})

	// Global Middlewares (Observability, Resilience, Security)
	app.Use(gota.RequestID())
	app.Use(gota.Logger())
	app.Use(gota.Recover())
	app.Use(gota.SecureHeaders())
	app.Use(gota.Gzip())
	app.Use(gota.CORS())

	// Protect documentation, schemas, and SDK routes (/docs, /openapi.json, /types.d.ts, /sdk.ts)
	// with HTTP Basic Auth and built-in Rate Limiting to prevent brute-force attacks:
	docsPassword := os.Getenv("DOCS_PASSWORD")
	if docsPassword == "" {
		if isProd {
			log.Fatal("DOCS_PASSWORD environment variable must be set in production")
		}
		docsPassword = "admin123" // Fallback password for local development
	}
	app.UseDocs(gota.BasicAuth("admin", docsPassword, gota.BasicAuthOptions{
		Realm: "Order Flow Documentation",
		RateLimit: &gota.RateLimiterOptions{
			Max:    10,          // Max 10 attempts
			Window: time.Minute, // per minute per client IP
		},
		AllowWeakCredentials: !isProd, // silences the weak password warning only in local development
		RequireTLS:           isProd,  // enforces HTTPS for remote clients in production
	}))

	// Data store pre-seeded with order-flow demo data
	store := data.NewStore()

	// Redirect root "/" directly to Scalar documentation (matches order-flow Program.cs)
	app.Get("/", func(ctx *gota.Context) error {
		return ctx.Redirect("/docs")
	}).Hidden()

	// 1. Public Auth Endpoints (Anonymous: Login, Register)
	auth.RegisterPublicRoutes(app, store)

	// 2. Protected Endpoints (Secured with built-in gota.JWT middleware)
	protectedGroup := app.Group("")
	protectedGroup.Use(gota.JWT(auth.Secret, gota.JWTOptions{
		ClaimsFactory: func() any { return &auth.UserClaims{} },
		RequireExp:    true,
	}))
	auth.RegisterProtectedRoutes(protectedGroup, store)
	orders.RegisterRoutes(protectedGroup, store)

	// Starts server with automatic graceful shutdown on SIGINT / SIGTERM
	if err := app.ListenGraceful(":3000"); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}
