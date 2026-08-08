using FastEndpoints;
using Microsoft.EntityFrameworkCore;
using OrderFlow.Data;
using OrderFlow.Domain;

namespace OrderFlow.Features.Orders.RemoveItem;

public class RemoveItemEndpoint : Endpoint<Request, Response>
{
    public AppDbContext Db { get; set; } = null!;

    public override void Configure()
    {
        Delete("/api/orders/{orderId}/items/{itemId}");

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
            AddError("Não é possível remover itens. O pedido não está mais em estado de criação.");
            ThrowIfAnyErrors();
            return;
        }

        var item = order.Items.FirstOrDefault(i => i.Id == req.ItemId);
        if (item == null)
        {
            await Send.NotFoundAsync(cancellation: ct);
            return;
        }

        order.Items.Remove(item);
        Db.OrderItems.Remove(item);
        await Db.SaveChangesAsync(ct);

        await Send.OkAsync(new Response
        {
            Message = "Item removido com sucesso!"
        }, cancellation: ct);
    }
}
