# Order Flow Frontend 📦⚡

Zero-build single-page application built with **Vue 3** and **Tailwind CSS** via CDN, designed to consume the **Order Flow API (Gota Framework)** without requiring Node.js, npm, bundlers, or any compilation steps.

---

## 🚀 Quick Start

1. Start the Order Flow backend server:
   ```bash
   cd ..
   go run main.go
   ```

2. Open `index.html` in your web browser:
   - Double-click `index.html`, or
   - Use any static HTTP file server:
     ```bash
     python3 -m http.server 8080
     ```
     Then navigate to [http://localhost:8080](http://localhost:8080).

---

## ✨ Features

- **Zero-Build Architecture**: Uses Vue 3 (`global.prod.js`) and Tailwind CSS directly via CDN.
- **Full Gota Client SDK Runtime**: Implements `createGotaClient()` matching the auto-generated TypeScript SDK in `../dist/sdk.ts`.
- **Role-Based Authentication**:
  - One-click login shortcuts for all 4 seed accounts (`admin`, `carlos`, `maria`, `luis`).
  - Modal form for custom credentials and new user registration (`POST /api/auth/register`).
  - Scoped order visibility: Customers only see their own orders, while Admins see all orders across the entire system.
  - Profile inspection (`GET /api/auth/me`) and permanent account deletion (`DELETE /api/auth/account`).
- **Complete Order Lifecycle Management**:
  - **Create Order** (`POST /api/orders`) with optional nested delivery address (street, city, 8-digit CEP).
  - **Order Directory** (`GET /api/orders`) with status filtering (`All`, `Created`, `Paid`, `Shipped`, `Cancelled`).
  - **Add Line Items** (`POST /api/orders/:orderId/items`) and **Remove Line Items** (`DELETE /api/orders/:orderId/items/:itemId`) while order is in `Created` status.
  - **Pay Order** (`POST /api/orders/:orderId/pay`) with payment method (`pix`, `credit_card`, `boleto`) and currency selection.
  - **Ship Order** (`POST /api/orders/:orderId/ship`) transitioning `Paid` orders to `Shipped`.
  - **Cancel Order** (`POST /api/orders/:orderId/cancel`).
  - **Multipart Receipt Upload** (`POST /api/orders/:orderId/receipt`) for attaching payment vouchers or invoices.
  - **Real-Time Live Tracking via SSE** (`GET /api/orders/:orderId/track`) streaming server-sent events with live status and summary chunks.
- **Rich Validation & Error Feedback**: Displays field-specific 422 validation errors with field indicators.
- **Dark & Light Mode**: Built with Tailwind CSS colors that automatically adapt to system appearance preferences.
- **Interactive JSON Inspector**: Inspect raw server responses directly from the interface.
