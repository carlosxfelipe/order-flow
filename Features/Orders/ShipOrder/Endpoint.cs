using FastEndpoints;
using OrderFlow.Data;
using OrderFlow.Domain;

namespace OrderFlow.Features.Orders.ShipOrder;

public class ShipOrderEndpoint : Endpoint<Request, Response>
{
    public override void Configure()
    {
        Post("/api/orders/{orderId}/ship");
        AllowAnonymous();
    }

    public override async Task HandleAsync(Request req, CancellationToken ct)
    {
        if (!Database.Orders.TryGetValue(req.OrderId, out var order))
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

        await Send.OkAsync(new Response
        {
            Message = "Pedido enviado! Suas Coca-Colas estão a caminho!"
        }, cancellation: ct);
    }
}
