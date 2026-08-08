using FastEndpoints;
using OrderFlow.Data;
using OrderFlow.Domain;

namespace OrderFlow.Features.Orders.AddItem;

public class AddItemEndpoint : Endpoint<Request, Response>
{
    public override void Configure()
    {
        Post("/api/orders/{orderId}/items");
        AllowAnonymous();
    }

    public override async Task HandleAsync(Request req, CancellationToken ct)
    {
        if (!Database.Orders.TryGetValue(req.OrderId, out var order))
        {
            await Send.NotFoundAsync(cancellation: ct);
            return;
        }

        if (order.Status != OrderStatus.Created)
        {
            AddError("Não é possível adicionar itens. O pedido não está no estado inicial.");
            ThrowIfAnyErrors();
            return;
        }

        // Using Coca-Cola as requested in examples/defaults
        var item = new OrderItem
        {
            ProductName = string.IsNullOrWhiteSpace(req.ProductName) ? "Coca-Cola Lata 350ml" : req.ProductName,
            Price = req.Price <= 0 ? 5.50m : req.Price,
            Quantity = req.Quantity <= 0 ? 1 : req.Quantity
        };

        order.Items.Add(item);

        await Send.OkAsync(new Response
        {
            Message = "Item adicionado ao pedido com sucesso!",
            ItemId = item.Id
        }, cancellation: ct);
    }
}
