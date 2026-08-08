using FastEndpoints;
using Microsoft.EntityFrameworkCore;
using OrderFlow.Data;

namespace OrderFlow.Features.Orders.ListOrders;

public class ListOrdersEndpoint : EndpointWithoutRequest<Response>
{
    public AppDbContext Db { get; set; } = null!;

    public override void Configure()
    {
        Get("/api/orders");
        AllowAnonymous();
    }

    public override async Task HandleAsync(CancellationToken ct)
    {
        var dbOrders = await Db.Orders.Include(o => o.Items).ToListAsync(ct);

        var orders = dbOrders.Select(o => new OrderSummary
        {
            Id = o.Id,
            CustomerName = o.CustomerName,
            Status = o.Status.ToString(),
            TotalAmount = o.TotalAmount
        }).ToList();

        await Send.OkAsync(new Response
        {
            Orders = orders
        }, cancellation: ct);
    }
}
