using FastEndpoints;
using Microsoft.EntityFrameworkCore;
using OrderFlow.Data;

namespace OrderFlow.Features.Orders.GetOrder;

public class GetOrderEndpoint : Endpoint<Request, Response>
{
    public AppDbContext Db { get; set; } = null!;

    public override void Configure()
    {
        Get("/api/orders/{id}");

    }

    public override async Task HandleAsync(Request req, CancellationToken ct)
    {
        var order = await Db.Orders
            .Include(o => o.Items)
            .FirstOrDefaultAsync(o => o.Id == req.Id, ct);

        if (order != null)
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
