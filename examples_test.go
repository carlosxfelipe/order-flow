package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/carlosxfelipe/gota"
	"order-flow-gota/data"
	"order-flow-gota/domain"
	"order-flow-gota/features/auth"
	"order-flow-gota/features/orders"
)

func setupTestApp() (*gota.App, *data.Store) {
	app := gota.New(gota.Config{
		Title:                 "Order Flow API Test",
		Version:               "1.0.0",
		DocsCSPEnforce:        true,
		DisallowUnknownFields: true,
	})
	app.Use(gota.RequestID())
	app.Use(gota.SecureHeaders())
	app.Use(gota.Gzip())

	store := data.NewStore()

	// Redirect root "/" directly to Scalar documentation
	app.Get("/", func(ctx *gota.Context) error {
		return ctx.Redirect("/docs")
	}).Hidden()

	// Auth (public)
	auth.RegisterPublicRoutes(app, store)

	// Protected routes (Auth + Orders)
	protectedGroup := app.Group("")
	protectedGroup.Use(gota.JWT(auth.Secret, gota.JWTOptions{
		ClaimsFactory: func() any { return &auth.UserClaims{} },
		RequireExp:    true,
	}))
	auth.RegisterProtectedRoutes(protectedGroup, store)
	orders.RegisterRoutes(protectedGroup, store)

	return app, store
}

func doRequest(app *gota.App, method, path, token string, body any) *httptest.ResponseRecorder {
	var bodyReader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	return rec
}

func TestOrderFlowEndToEnd(t *testing.T) {
	app, _ := setupTestApp()

	// 1. Login Carlos (Customer)
	loginRes := doRequest(app, http.MethodPost, "/api/auth/login", "", map[string]string{
		"username": "carlos",
		"password": "carlos123",
	})
	if loginRes.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d: %s", loginRes.Code, loginRes.Body.String())
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginRes.Body.Bytes(), &loginBody); err != nil || loginBody.Token == "" {
		t.Fatalf("failed to decode login token: %v", err)
	}
	carlosToken := loginBody.Token
	t.Log("Carlos Token:", carlosToken)

	// 2. Login Admin
	adminLoginRes := doRequest(app, http.MethodPost, "/api/auth/login", "", map[string]string{
		"username": "admin",
		"password": "admin123",
	})
	if adminLoginRes.Code != http.StatusOK {
		t.Fatalf("expected admin login 200, got %d", adminLoginRes.Code)
	}

	// 3. Unauthorized access check
	noAuthRes := doRequest(app, http.MethodGet, "/api/orders", "", nil)
	if noAuthRes.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", noAuthRes.Code)
	}

	// 4. Create Order
	createRes := doRequest(app, http.MethodPost, "/api/orders", carlosToken, map[string]string{
		"CustomerName": "Carlos",
	})
	if createRes.Code != http.StatusOK {
		t.Fatalf("expected create order 200, got %d: %s", createRes.Code, createRes.Body.String())
	}
	var createBody struct {
		OrderId string `json:"orderId"`
	}
	if err := json.Unmarshal(createRes.Body.Bytes(), &createBody); err != nil || createBody.OrderId == "" {
		t.Fatalf("failed to extract orderId: %v", err)
	}
	orderID := createBody.OrderId

	// 5. Add Item
	addItemRes := doRequest(app, http.MethodPost, "/api/orders/"+orderID+"/items", carlosToken, map[string]any{
		"ProductName": "Coca-Cola Can 350ml",
		"Price":       5.50,
		"Quantity":    2,
	})
	if addItemRes.Code != http.StatusOK {
		t.Fatalf("expected add item 200, got %d: %s", addItemRes.Code, addItemRes.Body.String())
	}
	var addItemBody struct {
		ItemId string `json:"itemId"`
	}
	if err := json.Unmarshal(addItemRes.Body.Bytes(), &addItemBody); err != nil || addItemBody.ItemId == "" {
		t.Fatalf("failed to extract itemId: %v", err)
	}
	firstItemID := addItemBody.ItemId

	// Add second item so we can remove one and still have items
	doRequest(app, http.MethodPost, "/api/orders/"+orderID+"/items", carlosToken, map[string]any{
		"ProductName": "Burger",
		"Price":       25.00,
		"Quantity":    1,
	})

	// 6. List Orders
	listRes := doRequest(app, http.MethodGet, "/api/orders", carlosToken, nil)
	if listRes.Code != http.StatusOK {
		t.Fatalf("expected list orders 200, got %d", listRes.Code)
	}

	// 7. Get Order
	getRes := doRequest(app, http.MethodGet, "/api/orders/"+orderID, carlosToken, nil)
	if getRes.Code != http.StatusOK {
		t.Fatalf("expected get order 200, got %d: %s", getRes.Code, getRes.Body.String())
	}
	var getBody struct {
		ID     string             `json:"id"`
		Status domain.OrderStatus `json:"status"`
		Items  []domain.OrderItem `json:"items"`
	}
	if err := json.Unmarshal(getRes.Body.Bytes(), &getBody); err != nil || len(getBody.Items) != 2 {
		t.Fatalf("expected 2 items in order, got %d", len(getBody.Items))
	}

	// 8. Remove Item
	removeRes := doRequest(app, http.MethodDelete, "/api/orders/"+orderID+"/items/"+firstItemID, carlosToken, nil)
	if removeRes.Code != http.StatusOK {
		t.Fatalf("expected remove item 200, got %d: %s", removeRes.Code, removeRes.Body.String())
	}

	// 9. Pay Order
	payRes := doRequest(app, http.MethodPost, "/api/orders/"+orderID+"/pay", carlosToken, nil)
	if payRes.Code != http.StatusOK {
		t.Fatalf("expected pay order 200, got %d: %s", payRes.Code, payRes.Body.String())
	}

	// 10. Ship Order
	shipRes := doRequest(app, http.MethodPost, "/api/orders/"+orderID+"/ship", carlosToken, nil)
	if shipRes.Code != http.StatusOK {
		t.Fatalf("expected ship order 200, got %d: %s", shipRes.Code, shipRes.Body.String())
	}

	// 11. Cannot cancel shipped order
	cancelShippedRes := doRequest(app, http.MethodPost, "/api/orders/"+orderID+"/cancel", carlosToken, nil)
	if cancelShippedRes.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when cancelling shipped order, got %d", cancelShippedRes.Code)
	}

	// 12. Create another order and cancel it
	cancelOrderRes := doRequest(app, http.MethodPost, "/api/orders", carlosToken, map[string]string{
		"CustomerName": "Carlos 2",
	})
	var cancelOrderBody struct {
		OrderId string `json:"orderId"`
	}
	json.Unmarshal(cancelOrderRes.Body.Bytes(), &cancelOrderBody)
	cancelSuccessRes := doRequest(app, http.MethodPost, "/api/orders/"+cancelOrderBody.OrderId+"/cancel", carlosToken, nil)
	if cancelSuccessRes.Code != http.StatusOK {
		t.Fatalf("expected 200 when cancelling open order, got %d", cancelSuccessRes.Code)
	}
}

func TestAuthLifecycle(t *testing.T) {
	app, _ := setupTestApp()

	// 1. Register new user
	regRes := doRequest(app, http.MethodPost, "/api/auth/register", "", map[string]string{
		"username": "tester",
		"password": "secretpassword",
	})
	if regRes.Code != http.StatusOK {
		t.Fatalf("expected register 200, got %d: %s", regRes.Code, regRes.Body.String())
	}

	var regBody struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
		Token    string `json:"token"`
	}
	if err := json.Unmarshal(regRes.Body.Bytes(), &regBody); err != nil {
		t.Fatalf("failed to parse register response: %v", err)
	}
	if regBody.Username != "tester" || regBody.Token == "" {
		t.Fatalf("unexpected register output: %+v", regBody)
	}

	// 2. Reject duplicate username
	dupRes := doRequest(app, http.MethodPost, "/api/auth/register", "", map[string]string{
		"username": "tester",
		"password": "secretpassword2",
	})
	if dupRes.Code != http.StatusBadRequest {
		t.Fatalf("expected duplicate register 400, got %d", dupRes.Code)
	}

	// 3. Get profile (/api/auth/me) with token
	meRes := doRequest(app, http.MethodGet, "/api/auth/me", regBody.Token, nil)
	if meRes.Code != http.StatusOK {
		t.Fatalf("expected me 200, got %d: %s", meRes.Code, meRes.Body.String())
	}
	var me domain.User
	if err := json.Unmarshal(meRes.Body.Bytes(), &me); err != nil {
		t.Fatalf("failed to parse me response: %v", err)
	}
	if me.Username != "tester" || me.Role != "Customer" {
		t.Fatalf("unexpected me output: %+v", me)
	}

	// 4. Delete account (/api/auth/account)
	delRes := doRequest(app, http.MethodDelete, "/api/auth/account", regBody.Token, nil)
	if delRes.Code != http.StatusOK {
		t.Fatalf("expected delete account 200, got %d: %s", delRes.Code, delRes.Body.String())
	}

	// 5. Subsequent /api/auth/me should fail with 404
	meAfterDel := doRequest(app, http.MethodGet, "/api/auth/me", regBody.Token, nil)
	if meAfterDel.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after deletion, got %d", meAfterDel.Code)
	}

	// 6. Login with deleted user should fail
	loginRes := doRequest(app, http.MethodPost, "/api/auth/login", "", map[string]string{
		"username": "tester",
		"password": "secretpassword",
	})
	if loginRes.Code != http.StatusBadRequest {
		t.Fatalf("expected login failure 400 after deletion, got %d", loginRes.Code)
	}
}

func TestDocsAndRedirectRoutes(t *testing.T) {
	app, _ := setupTestApp()

	// 1. Root redirect
	resRoot := doRequest(app, http.MethodGet, "/", "", nil)
	if resRoot.Code != http.StatusFound && resRoot.Code != http.StatusMovedPermanently {
		t.Fatalf("expected redirect for /, got status %d", resRoot.Code)
	}
	if resRoot.Header().Get("Location") != "/docs" {
		t.Errorf("expected Location /docs, got %s", resRoot.Header().Get("Location"))
	}

	// 2. Scalar UI HTML
	resDocs := doRequest(app, http.MethodGet, "/docs", "", nil)
	if resDocs.Code != http.StatusOK {
		t.Fatalf("expected 200 for /docs, got %d", resDocs.Code)
	}
	if !bytes.Contains(resDocs.Body.Bytes(), []byte("@scalar/api-reference")) {
		t.Errorf("expected Scalar script in /docs output")
	}

	// 3. OpenAPI 3.1 JSON spec
	resOpenAPI := doRequest(app, http.MethodGet, "/openapi.json", "", nil)
	if resOpenAPI.Code != http.StatusOK {
		t.Fatalf("expected 200 for /openapi.json, got %d", resOpenAPI.Code)
	}
	var spec map[string]any
	if err := json.Unmarshal(resOpenAPI.Body.Bytes(), &spec); err != nil {
		t.Fatalf("failed to decode openapi.json: %v", err)
	}
	if spec["openapi"] != "3.1.0" {
		t.Errorf("expected openapi 3.1.0, got %v", spec["openapi"])
	}
	paths, ok := spec["paths"].(map[string]any)
	if !ok || len(paths) == 0 {
		t.Fatal("expected non-empty paths in OpenAPI spec")
	}
	expectedEndpoints := []string{
		"/api/auth/login",
		"/api/auth/register",
		"/api/auth/me",
		"/api/auth/account",
		"/api/orders",
		"/api/orders/{orderId}",
		"/api/orders/{orderId}/items",
		"/api/orders/{orderId}/items/{itemId}",
		"/api/orders/{orderId}/pay",
		"/api/orders/{orderId}/ship",
		"/api/orders/{orderId}/cancel",
		"/api/orders/{orderId}/receipt",
		"/api/orders/{orderId}/track",
	}
	for _, ep := range expectedEndpoints {
		if _, exists := paths[ep]; !exists {
			t.Errorf("expected endpoint %s in OpenAPI spec", ep)
		}
	}

	// Verify automatic BearerAuth security scheme & protected route security in OpenAPI 3.1
	components, ok := spec["components"].(map[string]any)
	if !ok {
		t.Fatal("expected components in OpenAPI spec")
	}
	secSchemes, ok := components["securitySchemes"].(map[string]any)
	if !ok || secSchemes["BearerAuth"] == nil {
		t.Errorf("expected automatic BearerAuth in components.securitySchemes, got %v", secSchemes)
	}
	ordersPath, _ := paths["/api/orders"].(map[string]any)
	ordersGet, _ := ordersPath["get"].(map[string]any)
	if ordersGet["security"] == nil {
		t.Errorf("expected protected route GET /api/orders to have security requirement in OpenAPI spec")
	}
	loginPath, _ := paths["/api/auth/login"].(map[string]any)
	loginPost, _ := loginPath["post"].(map[string]any)
	if loginPost["security"] != nil {
		t.Errorf("expected public route POST /api/auth/login not to have security requirement, got %v", loginPost["security"])
	}
}

func TestDocsProtectedWithBasicAuth(t *testing.T) {
	app, _ := setupTestApp()
	app.UseDocs(gota.BasicAuth("admin", "admin123", gota.BasicAuthOptions{
		AllowWeakCredentials: true,
	}))

	// 1. Without credentials -> 401 Unauthorized
	reqDocs := httptest.NewRequest(http.MethodGet, "/docs", nil)
	recDocs := httptest.NewRecorder()
	app.ServeHTTP(recDocs, reqDocs)
	if recDocs.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for /docs without credentials, got %d", recDocs.Code)
	}

	reqOpenAPI := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	recOpenAPI := httptest.NewRecorder()
	app.ServeHTTP(recOpenAPI, reqOpenAPI)
	if recOpenAPI.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for /openapi.json without credentials, got %d", recOpenAPI.Code)
	}

	// 2. With valid Basic Auth credentials -> 200 OK
	reqAuthedDocs := httptest.NewRequest(http.MethodGet, "/docs", nil)
	reqAuthedDocs.SetBasicAuth("admin", "admin123")
	recAuthedDocs := httptest.NewRecorder()
	app.ServeHTTP(recAuthedDocs, reqAuthedDocs)
	if recAuthedDocs.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /docs with Basic Auth, got %d", recAuthedDocs.Code)
	}

	reqAuthedOpenAPI := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	reqAuthedOpenAPI.SetBasicAuth("admin", "admin123")
	recAuthedOpenAPI := httptest.NewRecorder()
	app.ServeHTTP(recAuthedOpenAPI, reqAuthedOpenAPI)
	if recAuthedOpenAPI.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /openapi.json with Basic Auth, got %d", recAuthedOpenAPI.Code)
	}
}

func TestOrderRichValidationAndObservability(t *testing.T) {
	app, _ := setupTestApp()

	// 1. Check Request-ID and Security Headers on responses
	res := doRequest(app, http.MethodPost, "/api/auth/login", "", map[string]string{
		"username": "carlos",
		"password": "carlos123",
	})
	if res.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header in response")
	}
	if res.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected X-Content-Type-Options nosniff, got %q", res.Header().Get("X-Content-Type-Options"))
	}
	if res.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("expected X-Frame-Options DENY, got %q", res.Header().Get("X-Frame-Options"))
	}
	if res.Header().Get("Referrer-Policy") != "no-referrer" {
		t.Errorf("expected Referrer-Policy no-referrer, got %q", res.Header().Get("Referrer-Policy"))
	}

	// Gzip compression for clients accepting gzip
	reqGzip := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	reqGzip.Header.Set("Accept-Encoding", "gzip")
	recGzip := httptest.NewRecorder()
	app.ServeHTTP(recGzip, reqGzip)
	if recGzip.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected Content-Encoding gzip, got %q", recGzip.Header().Get("Content-Encoding"))
	}

	var loginBody struct {
		Token string `json:"token"`
	}
	json.Unmarshal(res.Body.Bytes(), &loginBody)
	token := loginBody.Token

	// 2. Registration must reject unknown/unmapped fields like "role" (mass-assignment protection)
	escalateRes := doRequest(app, http.MethodPost, "/api/auth/register", "", map[string]string{
		"username": "hacker123",
		"password": "password123",
		"role":     "Admin",
	})
	if escalateRes.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for unknown field role, got %d: %s", escalateRes.Code, escalateRes.Body.String())
	}

	// 3. Validate nested struct & len=8 on ZipCode (invalid CEP length)
	badZipRes := doRequest(app, http.MethodPost, "/api/orders", token, map[string]any{
		"customerName": "Carlos Felipe",
		"shippingAddress": map[string]string{
			"street":  "Av. Paulista, 1000",
			"city":    "São Paulo",
			"zipCode": "123", // invalid: length must be exactly 8
		},
	})
	if badZipRes.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for invalid zipCode len, got %d: %s", badZipRes.Code, badZipRes.Body.String())
	}

	// 4. Valid nested struct with len=8 succeeds
	validOrderRes := doRequest(app, http.MethodPost, "/api/orders", token, map[string]any{
		"customerName": "Carlos Felipe",
		"shippingAddress": map[string]string{
			"street":  "Av. Paulista, 1000",
			"city":    "São Paulo",
			"zipCode": "01310100", // valid: exactly 8 digits
		},
	})
	if validOrderRes.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid order with nested address, got %d: %s", validOrderRes.Code, validOrderRes.Body.String())
	}
	var createdOrder struct {
		OrderId string `json:"orderId"`
	}
	json.Unmarshal(validOrderRes.Body.Bytes(), &createdOrder)

	// Add item to allow payment
	doRequest(app, http.MethodPost, "/api/orders/"+createdOrder.OrderId+"/items", token, map[string]any{
		"productName": "Item A",
		"price":       10.0,
		"quantity":    1,
	})

	// 5. Validate oneof on payment method (crypto is invalid)
	badPayRes := doRequest(app, http.MethodPost, "/api/orders/"+createdOrder.OrderId+"/pay", token, map[string]string{
		"method": "crypto",
	})
	if badPayRes.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for invalid payment method, got %d: %s", badPayRes.Code, badPayRes.Body.String())
	}

	// 6. Validate oneof on currency (REAIS is not in BRL,USD,EUR)
	badCurrRes := doRequest(app, http.MethodPost, "/api/orders/"+createdOrder.OrderId+"/pay", token, map[string]string{
		"method":   "pix",
		"currency": "REAIS", // invalid: must be oneof BRL,USD,EUR
	})
	if badCurrRes.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for invalid currency, got %d: %s", badCurrRes.Code, badCurrRes.Body.String())
	}

	// 7. Valid payment with oneof & len=3 passes
	validPayRes := doRequest(app, http.MethodPost, "/api/orders/"+createdOrder.OrderId+"/pay", token, map[string]string{
		"method":   "pix",
		"currency": "BRL",
	})
	if validPayRes.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid payment, got %d: %s", validPayRes.Code, validPayRes.Body.String())
	}
}

func TestOrderMultipartReceiptUpload(t *testing.T) {
	app, _ := setupTestApp()

	// 1. Login Carlos
	loginRes := doRequest(app, http.MethodPost, "/api/auth/login", "", map[string]string{
		"username": "carlos",
		"password": "carlos123",
	})
	var loginBody struct {
		Token string `json:"token"`
	}
	json.Unmarshal(loginRes.Body.Bytes(), &loginBody)

	// 2. Prepare multipart request with fake receipt PDF
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fileWriter, err := writer.CreateFormFile("receipt", "comprovante_pix.pdf")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	fakeContent := []byte("%PDF-1.4 mock receipt file content")
	fileWriter.Write(fakeContent)

	_ = writer.WriteField("notes", "Payment confirmed via Pix QR code")
	_ = writer.Close()

	orderID := "b1a2c3d4-0001-4000-8000-000000000001"
	req := httptest.NewRequest(http.MethodPost, "/api/orders/"+orderID+"/receipt", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+loginBody.Token)

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for receipt upload, got %d: %s", rec.Code, rec.Body.String())
	}

	var res struct {
		OrderId  string `json:"orderId"`
		Filename string `json:"filename"`
		Size     int64  `json:"size"`
		Notes    string `json:"notes"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse receipt response: %v", err)
	}

	if res.Filename != "comprovante_pix.pdf" || res.Size != int64(len(fakeContent)) || res.OrderId != orderID || res.Notes != "Payment confirmed via Pix QR code" {
		t.Errorf("unexpected upload response payload: %+v", res)
	}
}

func TestOrderSSETracking(t *testing.T) {
	app, _ := setupTestApp()

	orderID := "b1a2c3d4-0001-4000-8000-000000000001"
	req := httptest.NewRequest(http.MethodGet, "/api/orders/"+orderID+"/track", nil)
	// Authenticate Carlos
	loginRes := doRequest(app, http.MethodPost, "/api/auth/login", "", map[string]string{
		"username": "carlos",
		"password": "carlos123",
	})
	var loginBody struct {
		Token string `json:"token"`
	}
	json.Unmarshal(loginRes.Body.Bytes(), &loginBody)
	req.Header.Set("Authorization", "Bearer "+loginBody.Token)

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for order tracking SSE, got %d", rec.Code)
	}

	if rec.Header().Get("Content-Type") != "text/event-stream" {
		t.Fatalf("expected Content-Type text/event-stream, got %s", rec.Header().Get("Content-Type"))
	}

	bodyStr := rec.Body.String()
	if !strings.Contains(bodyStr, "event: order_status\n") {
		t.Errorf("expected order_status event in SSE stream: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "event: order_summary\n") {
		t.Errorf("expected order_summary event in SSE stream: %s", bodyStr)
	}
}

func TestVerifyAllRoutesIndividually(t *testing.T) {
	app, _ := setupTestApp()

	// 1. Root redirect (GET /)
	t.Run("01_GET_/", func(t *testing.T) {
		res := doRequest(app, http.MethodGet, "/", "", nil)
		if res.Code != http.StatusFound && res.Code != http.StatusMovedPermanently {
			t.Fatalf("expected 302, got %d", res.Code)
		}
		if res.Header().Get("Location") != "/docs" {
			t.Errorf("expected Location /docs, got %s", res.Header().Get("Location"))
		}
	})

	// 2. Scalar Documentation (GET /docs)
	t.Run("02_GET_/docs", func(t *testing.T) {
		res := doRequest(app, http.MethodGet, "/docs", "", nil)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.Code)
		}
		if !strings.Contains(res.Body.String(), "@scalar/api-reference") {
			t.Errorf("expected scalar content")
		}
	})

	// 3. OpenAPI 3.1 Spec (GET /openapi.json)
	t.Run("03_GET_/openapi.json", func(t *testing.T) {
		res := doRequest(app, http.MethodGet, "/openapi.json", "", nil)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.Code)
		}
		var spec map[string]any
		if err := json.Unmarshal(res.Body.Bytes(), &spec); err != nil {
			t.Fatalf("invalid json spec: %v", err)
		}
	})

	// 4. Register Account (POST /api/auth/register)
	var authToken string
	t.Run("04_POST_/api/auth/register", func(t *testing.T) {
		res := doRequest(app, http.MethodPost, "/api/auth/register", "", map[string]string{
			"username": "all_routes_tester",
			"password": "strongpassword123",
		})
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
		}
		var body struct {
			Token string `json:"token"`
		}
		json.Unmarshal(res.Body.Bytes(), &body)
		if body.Token == "" {
			t.Fatal("empty token returned")
		}
		authToken = body.Token
	})

	// 5. Login (POST /api/auth/login)
	t.Run("05_POST_/api/auth/login", func(t *testing.T) {
		res := doRequest(app, http.MethodPost, "/api/auth/login", "", map[string]string{
			"username": "all_routes_tester",
			"password": "strongpassword123",
		})
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
		}
	})

	// 6. Current User Profile (GET /api/auth/me)
	t.Run("06_GET_/api/auth/me", func(t *testing.T) {
		res := doRequest(app, http.MethodGet, "/api/auth/me", authToken, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
		}
		var user domain.User
		json.Unmarshal(res.Body.Bytes(), &user)
		if user.Username != "all_routes_tester" {
			t.Errorf("expected all_routes_tester, got %s", user.Username)
		}
	})

	// 7. List Orders (GET /api/orders)
	t.Run("07_GET_/api/orders", func(t *testing.T) {
		res := doRequest(app, http.MethodGet, "/api/orders", authToken, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
		}
	})

	// 8. Create Order (POST /api/orders)
	var orderID string
	t.Run("08_POST_/api/orders", func(t *testing.T) {
		res := doRequest(app, http.MethodPost, "/api/orders", authToken, map[string]any{
			"customerName": "All Routes Tester",
			"shippingAddress": map[string]string{
				"street":  "Rua das Palmeiras, 10",
				"city":    "Porto Alegre",
				"zipCode": "90000000",
			},
		})
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
		}
		var body struct {
			OrderId string `json:"orderId"`
		}
		json.Unmarshal(res.Body.Bytes(), &body)
		if body.OrderId == "" {
			t.Fatal("empty orderId returned")
		}
		orderID = body.OrderId
	})

	// 9. Get Order Details (GET /api/orders/:orderId)
	t.Run("09_GET_/api/orders/:orderId", func(t *testing.T) {
		res := doRequest(app, http.MethodGet, "/api/orders/"+orderID, authToken, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
		}
	})

	// 10. Add Item to Order (POST /api/orders/:orderId/items)
	var itemID string
	t.Run("10_POST_/api/orders/:orderId/items", func(t *testing.T) {
		res := doRequest(app, http.MethodPost, "/api/orders/"+orderID+"/items", authToken, map[string]any{
			"productName": "Espresso Coffee",
			"price":       4.00,
			"quantity":    2,
		})
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
		}
		var body struct {
			ItemId string `json:"itemId"`
		}
		json.Unmarshal(res.Body.Bytes(), &body)
		itemID = body.ItemId

		// Add a second item so we can remove one and still pay
		doRequest(app, http.MethodPost, "/api/orders/"+orderID+"/items", authToken, map[string]any{
			"productName": "Croissant",
			"price":       7.50,
			"quantity":    1,
		})
	})

	// 11. Remove Item from Order (DELETE /api/orders/:orderId/items/:itemId)
	t.Run("11_DELETE_/api/orders/:orderId/items/:itemId", func(t *testing.T) {
		res := doRequest(app, http.MethodDelete, "/api/orders/"+orderID+"/items/"+itemID, authToken, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
		}
	})

	// 12. Upload Order Receipt (POST /api/orders/:orderId/receipt)
	t.Run("12_POST_/api/orders/:orderId/receipt", func(t *testing.T) {
		var body bytes.Buffer
		w := multipart.NewWriter(&body)
		part, err := w.CreateFormFile("receipt", "recibo_cafe.pdf")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		part.Write([]byte("%PDF-1.4 mock coffee receipt"))
		w.WriteField("notes", "Paid via Pix at checkout")
		w.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/orders/"+orderID+"/receipt", &body)
		req.Header.Set("Content-Type", w.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+authToken)
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 13. Track Order SSE Stream (GET /api/orders/:orderId/track)
	t.Run("13_GET_/api/orders/:orderId/track", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/orders/"+orderID+"/track", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if rec.Header().Get("Content-Type") != "text/event-stream" {
			t.Errorf("expected text/event-stream, got %s", rec.Header().Get("Content-Type"))
		}
	})

	// 14. Pay Order (POST /api/orders/:orderId/pay)
	t.Run("14_POST_/api/orders/:orderId/pay", func(t *testing.T) {
		res := doRequest(app, http.MethodPost, "/api/orders/"+orderID+"/pay", authToken, map[string]string{
			"method":   "pix",
			"currency": "BRL",
		})
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
		}
	})

	// 15. Ship Order (POST /api/orders/:orderId/ship)
	t.Run("15_POST_/api/orders/:orderId/ship", func(t *testing.T) {
		res := doRequest(app, http.MethodPost, "/api/orders/"+orderID+"/ship", authToken, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
		}
	})

	// 16. Cancel Order (POST /api/orders/:orderId/cancel)
	t.Run("16_POST_/api/orders/:orderId/cancel", func(t *testing.T) {
		cRes := doRequest(app, http.MethodPost, "/api/orders", authToken, map[string]any{
			"customerName": "Order To Cancel",
		})
		var cBody struct {
			OrderId string `json:"orderId"`
		}
		json.Unmarshal(cRes.Body.Bytes(), &cBody)

		res := doRequest(app, http.MethodPost, "/api/orders/"+cBody.OrderId+"/cancel", authToken, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
		}
	})

	// 17. Delete Account (DELETE /api/auth/account)
	t.Run("17_DELETE_/api/auth/account", func(t *testing.T) {
		res := doRequest(app, http.MethodDelete, "/api/auth/account", authToken, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
		}
	})
}

func TestSecurityOrderOwnershipAndModelAudit(t *testing.T) {
	// 1. Audit domain models for accidental sensitive field exposure
	for _, m := range []any{domain.User{}, domain.Order{}, domain.OrderItem{}, domain.ShippingAddress{}} {
		if warnings := gota.AuditModel(m); len(warnings) > 0 {
			t.Fatalf("expected 0 sensitive field warnings on domain model %T, got: %v", m, warnings)
		}
	}

	app, _ := setupTestApp()

	// 2. Login Carlos (Customer, UserID=2) and Admin (UserID=1)
	carlosLogin := doRequest(app, http.MethodPost, "/api/auth/login", "", map[string]string{
		"username": "carlos",
		"password": "carlos123",
	})
	var carlosBody struct {
		Token string `json:"token"`
	}
	json.Unmarshal(carlosLogin.Body.Bytes(), &carlosBody)

	adminLogin := doRequest(app, http.MethodPost, "/api/auth/login", "", map[string]string{
		"username": "admin",
		"password": "admin123",
	})
	var adminBody struct {
		Token string `json:"token"`
	}
	json.Unmarshal(adminLogin.Body.Bytes(), &adminBody)

	// Maria's order (UserID=3)
	mariasOrderID := "b1a2c3d4-0002-4000-8000-000000000002"

	// Carlos cannot read or cancel Maria's order (403 Forbidden)
	getOtherRes := doRequest(app, http.MethodGet, "/api/orders/"+mariasOrderID, carlosBody.Token, nil)
	if getOtherRes.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when Customer accesses another user's order, got %d", getOtherRes.Code)
	}
	cancelOtherRes := doRequest(app, http.MethodPost, "/api/orders/"+mariasOrderID+"/cancel", carlosBody.Token, nil)
	if cancelOtherRes.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when Customer cancels another user's order, got %d", cancelOtherRes.Code)
	}

	// Admin can read Maria's order (200 OK)
	adminGetRes := doRequest(app, http.MethodGet, "/api/orders/"+mariasOrderID, adminBody.Token, nil)
	if adminGetRes.Code != http.StatusOK {
		t.Fatalf("expected 200 OK when Admin accesses any order, got %d", adminGetRes.Code)
	}

	// 3. Multipart filename sanitization strips path traversal
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fileWriter, _ := writer.CreateFormFile("receipt", "../../etc/comprovante_pix.pdf")
	fileWriter.Write([]byte("%PDF-1.4 receipt"))
	writer.Close()

	carlosOrderID := "b1a2c3d4-0001-4000-8000-000000000001"
	req := httptest.NewRequest(http.MethodPost, "/api/orders/"+carlosOrderID+"/receipt", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+carlosBody.Token)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for sanitized receipt upload, got %d: %s", rec.Code, rec.Body.String())
	}
	var uploadRes struct {
		Filename string `json:"filename"`
	}
	json.Unmarshal(rec.Body.Bytes(), &uploadRes)
	if uploadRes.Filename != "comprovante_pix.pdf" {
		t.Fatalf("expected sanitized filename comprovante_pix.pdf, got %q", uploadRes.Filename)
	}
}

func BenchmarkGetOrders(b *testing.B) {
	app, _ := setupTestApp()
	loginRes := doRequest(app, http.MethodPost, "/api/auth/login", "", map[string]string{
		"username": "carlos",
		"password": "carlos123",
	})
	var loginBody struct {
		Token string `json:"token"`
	}
	json.Unmarshal(loginRes.Body.Bytes(), &loginBody)

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	req.Header.Set("Authorization", "Bearer "+loginBody.Token)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, req)
	}
}
