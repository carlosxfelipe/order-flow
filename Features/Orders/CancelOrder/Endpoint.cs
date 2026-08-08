using FastEndpoints;
using OrderFlow.Data;
using OrderFlow.Domain;

namespace OrderFlow.Features.Orders.CancelOrder;

public class CancelOrderEndpoint : Endpoint<Request, Response>
{
    public override void Configure()
    {
        Post("/api/orders/{orderId}/cancel");
        AllowAnonymous();
    }

    public override async Task HandleAsync(Request req, CancellationToken ct)
    {
        if (!Database.Orders.TryGetValue(req.OrderId, out var order))
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

        await Send.OkAsync(new Response
        {
            Message = "Pedido cancelado com sucesso."
        }, cancellation: ct);
    }
}
