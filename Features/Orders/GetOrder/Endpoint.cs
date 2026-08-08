using FastEndpoints;
using OrderFlow.Data;

namespace OrderFlow.Features.Orders.GetOrder;

public class GetOrderEndpoint : Endpoint<Request, Response>
{
    public override void Configure()
    {
        Get("/api/orders/{id}");
        AllowAnonymous();
    }

    public override async Task HandleAsync(Request req, CancellationToken ct)
    {
        if (Database.Orders.TryGetValue(req.Id, out var order))
        {
            await Send.OkAsync(new Response
            {
                Id = order.Id,
                CustomerName = order.CustomerName,
                Status = order.Status.ToString(),
                TotalAmount = order.TotalAmount,
                Items = order.Items,
                Message = "Pedido recuperado com sucesso!"
            }, cancellation: ct);
            return;
        }

        await Send.NotFoundAsync(cancellation: ct);
    }
}
