# Order Flow API

A simple order management API built using .NET 10, FastEndpoints (REPR pattern), and Entity Framework Core with SQLite.

## How to Clone

Clone the repository to your local machine:

```bash
git clone https://github.com/carlosxfelipe/order-flow.git
cd order-flow
```

## Setup and Run

The project uses a local SQLite database. The Entity Framework Core migrations are configured to run automatically upon application startup, creating the database file (`orderflow.db`) if it does not exist.

To start the application, run the following command in the project root:

```bash
dotnet run
```

## API Documentation

Once the application is running, it will automatically redirect the root path to the Scalar API documentation interface. 

You can access the documentation at:
http://localhost:5189/scalar
(or https://localhost:7186/scalar)

Through this interface, you can view request schemas, response models, and test all the routes directly from your browser.

## Testing with Bruno

You can also use [Bruno](https://www.usebruno.com/) to test the API routes locally. This project includes a `Bruno` collection folder at the root directory.

### Why Bruno instead of Postman?

If you are coming from Postman, Bruno offers a few key advantages for developer workflows:
* **Local First:** Bruno stores collections directly in your file system as plain text files (`.bru`), rather than forcing cloud synchronization.
* **Version Control Friendly:** Because collections are plain text files within your project folder, they can be easily committed to Git and versioned along with your source code.
* **No Account Required:** You don't need to sign up for an account, log in, or rely on a cloud service just to test your local APIs.
* **Privacy & Security:** Your API endpoints, payloads, and environment variables never leave your machine.

## Authentication

This project uses JWT (JSON Web Tokens) for security. 
When the database is initialized, a default admin user is seeded so you can test the secured endpoints.

**Default Credentials:**
- **Username:** `admin`
- **Password:** `admin123`

You can use these credentials in the `/api/auth/login` endpoint (via Bruno or Scalar UI) to obtain the Bearer Token.

## Available Routes

### Orders
* `POST /api/orders` - Creates a new empty order.
* `GET /api/orders` - Lists a summary of all existing orders.
* `GET /api/orders/{id}` - Retrieves the full details of a specific order, including its items.

### Items
* `POST /api/orders/{orderId}/items` - Adds a new item to an existing order (defaults to Coca-Cola if not specified).
* `DELETE /api/orders/{orderId}/items/{itemId}` - Removes a specific item from an order.

### Actions (Business Logic)
* `POST /api/orders/{orderId}/pay` - Marks an order as paid.
* `POST /api/orders/{orderId}/cancel` - Cancels an order (if not already shipped).
* `POST /api/orders/{orderId}/ship` - Ships an order (requires the order to be paid first).
