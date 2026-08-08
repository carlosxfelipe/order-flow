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
