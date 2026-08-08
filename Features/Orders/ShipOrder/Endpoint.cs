using FastEndpoints;
using Microsoft.EntityFrameworkCore;
using OrderFlow.Data;
using OrderFlow.Domain;

namespace OrderFlow.Features.Orders.ShipOrder;

public class ShipOrderEndpoint : Endpoint<Request, Response>
{
    public AppDbContext Db { get; set; } = null!;

    public override void Configure()
    {
        Post("/api/orders/{orderId}/ship");

    }

    public override async Task HandleAsync(Request req, CancellationToken ct)
    {
        var order = await Db.Orders.FirstOrDefaultAsync(o => o.Id == req.OrderId, ct);
        if (order == null)
        {
            await Send.NotFoundAsync(cancellation: ct);
            return;
        }

        if (order.Status != OrderStatus.Paid)
        {
            AddError("Apenas pedidos pagos podem ser enviados.");
            ThrowIfAnyErrors();
            return;
        }

        order.Status = OrderStatus.Shipped;
        await Db.SaveChangesAsync(ct);

        await Send.OkAsync(new Response
        {
            Message = "Pedido enviado! Suas Coca-Colas estão a caminho!"
        }, cancellation: ct);
    }
}
