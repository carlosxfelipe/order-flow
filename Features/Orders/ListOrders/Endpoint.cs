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

    }

    public override async Task HandleAsync(CancellationToken ct)
    {
        var query = Db.Orders.Include(o => o.Items).AsQueryable();

        if (!User.IsInRole("Admin"))
        {
            var userId = Convert.ToInt32(User.Claims.FirstOrDefault(c => c.Type == "UserId")?.Value ?? "0");
            query = query.Where(o => o.UserId == userId);
        }

        var dbOrders = await query.ToListAsync(ct);

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
