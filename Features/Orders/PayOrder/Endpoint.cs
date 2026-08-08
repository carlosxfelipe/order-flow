using FastEndpoints;
using OrderFlow.Data;
using OrderFlow.Domain;

namespace OrderFlow.Features.Orders.PayOrder;

public class PayOrderEndpoint : Endpoint<Request, Response>
{
    public override void Configure()
    {
        Post("/api/orders/{orderId}/pay");
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
            AddError("Apenas pedidos no status inicial podem ser pagos.");
            ThrowIfAnyErrors();
            return;
        }

        if (!order.Items.Any())
        {
            AddError("Não é possível pagar um pedido sem itens.");
            ThrowIfAnyErrors();
            return;
        }

        order.Status = OrderStatus.Paid;

        await Send.OkAsync(new Response
        {
            Message = "Pedido pago com sucesso! Refrescando com uma Coca-Cola bem gelada!"
        }, cancellation: ct);
    }
}
