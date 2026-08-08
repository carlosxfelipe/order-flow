using FastEndpoints;
using Microsoft.EntityFrameworkCore;
using OrderFlow.Data;
using OrderFlow.Domain;

namespace OrderFlow.Features.Orders.CancelOrder;

public class CancelOrderEndpoint : Endpoint<Request, Response>
{
    public AppDbContext Db { get; set; } = null!;

    public override void Configure()
    {
        Post("/api/orders/{orderId}/cancel");

    }

    public override async Task HandleAsync(Request req, CancellationToken ct)
    {
        var order = await Db.Orders.FirstOrDefaultAsync(o => o.Id == req.OrderId, ct);
        if (order == null)
        {
            await Send.NotFoundAsync(cancellation: ct);
            return;
        }

        if (order.Status == OrderStatus.Shipped)
        {
            AddError("Pedidos já enviados não podem ser cancelados.");
            ThrowIfAnyErrors();
            return;
        }

        order.Status = OrderStatus.Cancelled;
        await Db.SaveChangesAsync(ct);

        await Send.OkAsync(new Response
        {
            Message = "Pedido cancelado com sucesso."
        }, cancellation: ct);
    }
}
