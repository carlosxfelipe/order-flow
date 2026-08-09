using FastEndpoints;
using Microsoft.EntityFrameworkCore;
using OrderFlow.Data;
using OrderFlow.Domain;

namespace OrderFlow.Features.Orders.AddItem;

public class AddItemEndpoint : Endpoint<Request, Response>
{
    public AppDbContext Db { get; set; } = null!;

    public override void Configure()
    {
        Post("/api/orders/{orderId}/items");

    }

    public override async Task HandleAsync(Request req, CancellationToken ct)
    {
        var order = await Db.Orders.Include(o => o.Items).FirstOrDefaultAsync(o => o.Id == req.OrderId, ct);
        if (order == null)
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

        var item = new OrderItem
        {
            ProductName = string.IsNullOrWhiteSpace(req.ProductName) ? "Coca-Cola Lata 350ml" : req.ProductName,
            Price = req.Price <= 0 ? 5.50m : req.Price,
            Quantity = req.Quantity <= 0 ? 1 : req.Quantity
        };

        order.Items.Add(item);
        Db.OrderItems.Add(item); // EXPLICITLY mark as Added
        await Db.SaveChangesAsync(ct);

        await Send.OkAsync(new Response
        {
            Message = "Item adicionado ao pedido com sucesso!",
            ItemId = item.Id
        }, cancellation: ct);
    }
}
