# Order Flow API 💧

Order management API built with [Gota Framework](https://github.com/carlosxfelipe/gota) using the **REPR Design Pattern (Request-Endpoint-Response)**, native **OpenAPI 3.1 + Scalar UI**, and **zero external dependencies**.

---

## 🏃 Running the Application

```bash
go run main.go
```

Then visit:

- **Scalar UI**: [http://localhost:3000/docs](http://localhost:3000/docs) (or open [http://localhost:3000](http://localhost:3000) which redirects to `/docs`)
- **OpenAPI 3.1 Spec**: [http://localhost:3000/openapi.json](http://localhost:3000/openapi.json)
- **TypeScript Client SDK**: [http://localhost:3000/sdk.ts](http://localhost:3000/sdk.ts)
- **TypeScript Type Definitions**: [http://localhost:3000/types.d.ts](http://localhost:3000/types.d.ts)

> 🔒 **Documentation Authentication**: The `/docs`, `/openapi.json`, `/sdk.ts`, and `/types.d.ts` routes are protected with HTTP Basic Auth via `app.UseDocs(gota.BasicAuth("admin", "admin123"))`. Use user: `admin` and password: `admin123`.

## 🖥️ Zero-Build Frontend Application

A full-featured, zero-build client application using **Vue 3** and **Tailwind CSS** (via CDN) is available in [`frontend/`](frontend):

- **Quick Launch**: Open [`frontend/index.html`](frontend/index.html) directly in any browser (or serve with `python3 -m http.server` / `npx serve`).
- **Features Included**:
  - 1-click authentication switcher for seed users (`admin`, `carlos`, `maria`, `luis`).
  - Scoped order management (Customers see own orders; Admins view all).
  - Full order lifecycle: create order, add/remove line items, pay (Pix, Card, Boleto), ship, and cancel.
  - Multipart receipt upload (`POST /api/orders/:orderId/receipt`).
  - Live Server-Sent Events (SSE) stream tracking (`GET /api/orders/:orderId/track`).
  - Interactive JSON response inspector & 422 validation error feedback.


## 📦 Offline TypeScript SDK Codegen

To generate the full TypeScript Client SDK (`./dist/sdk.ts`) and TypeScript definitions (`./dist/types.d.ts`) offline (e.g., for CI/CD pipelines or frontend builds without running the HTTP server):

```bash
go run ./examples/codegen/main.go
```

## 🧪 Running Tests

```bash
go test ./... -v
```

## 🔄 Updating Gota Dependency

By default, `go.mod` uses a local `replace` directive pointing to `../gota` for side-by-side local development:

```go
replace github.com/carlosxfelipe/gota => ../gota
```

To switch to (or update to) the latest version of Gota directly from GitHub (`main` branch):

```bash
go mod edit -dropreplace=github.com/carlosxfelipe/gota -droprequire=github.com/carlosxfelipe/gota
go get github.com/carlosxfelipe/gota@main
go mod tidy
```

---

## 👥 Seed Users & Credentials

The showcase is pre-seeded with users for testing role-based workflows:

| Username | Password    | Role       | Permissions                                  |
| -------- | ----------- | ---------- | -------------------------------------------- |
| `admin`  | `admin123`  | `Admin`    | Full access (views all orders in the system) |
| `carlos` | `carlos123` | `Customer` | Scoped to own orders (seed `Created` order)  |
| `maria`  | `maria123`  | `Customer` | Scoped to own orders (seed `Paid` order)     |
| `luis`   | `luis123`   | `Customer` | Scoped to own orders (seed `Shipped` order)  |

---

## 🛣️ Showcase API Routes (Order Flow)

| Method   | Endpoint                             | Auth   | Description                                                                                                 |
| -------- | ------------------------------------ | ------ | ----------------------------------------------------------------------------------------------------------- |
| `POST`   | `/api/auth/register`                 | Public | Registers a new `Customer` account (role is always server-assigned) and returns an initial JWT Bearer token |
| `POST`   | `/api/auth/login`                    | Public | Authenticates credentials and returns a signed JWT Bearer token                                             |
| `GET`    | `/api/auth/me`                       | Bearer | Retrieves the profile of the currently authenticated user                                                   |
| `DELETE` | `/api/auth/account`                  | Bearer | Permanently deletes the authenticated user's account                                                        |
| `GET`    | `/api/orders`                        | Bearer | Lists orders (scoped by Customer, or full list for Admin)                                                   |
| `POST`   | `/api/orders`                        | Bearer | Creates a new order in `Created` state                                                                      |
| `GET`    | `/api/orders/:orderId`               | Bearer | Retrieves order details, items, and calculated `totalAmount`                                                |
| `POST`   | `/api/orders/:orderId/items`         | Bearer | Adds a line item to an order still in `Created` state                                                       |
| `DELETE` | `/api/orders/:orderId/items/:itemId` | Bearer | Removes a line item from an open `Created` order                                                            |
| `POST`   | `/api/orders/:orderId/pay`           | Bearer | Pays order with items and transitions status to `Paid`                                                      |
| `POST`   | `/api/orders/:orderId/ship`          | Bearer | Ships paid order and transitions status to `Shipped`                                                        |
| `POST`   | `/api/orders/:orderId/cancel`        | Bearer | Cancels order unless it was already `Shipped`                                                               |
| `POST`   | `/api/orders/:orderId/receipt`       | Bearer | Uploads a payment receipt or invoice document (`multipart/form-data`)                                       |
| `GET`    | `/api/orders/:orderId/track`         | Bearer | Streams real-time order updates via Server-Sent Events (`SSEStream`)                                        |

---

## 🟠 Testing with Bruno

A pre-configured [Bruno](https://www.usebruno.com/) collection is ready to use in [`bruno/`](bruno):

### Desktop App (GUI):

1. Open the `bruno` folder in the Bruno app.
2. Run **`10. Login (Customer)`**, **`9. Login (Admin)`**, or create a new user with **`11. Register User`** — the response script automatically stores `jwt_token`.
3. Test your profile with **`12. Get Current User (Me)`**, delete your account with **`13. Delete Account`**, or execute any of the order lifecycle requests (**`Create Order`**, **`Add Item`**, **`Pay Order`**, **`Upload Receipt`**, **`Track Order (SSE)`**, **`Ship Order`**, etc.).

### Command Line (Bruno CLI):

With the Gota server running (`go run main.go`), execute the entire collection automatically via terminal:

```bash
cd bruno
npx -y @usebruno/cli run --env Local
```

---

## 📄 License

MIT © [Carlos Felipe](https://github.com/carlosxfelipe)
