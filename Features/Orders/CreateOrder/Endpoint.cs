using FastEndpoints;
using OrderFlow.Data;
using OrderFlow.Domain;

namespace OrderFlow.Features.Orders.CreateOrder;

public class CreateOrderEndpoint : Endpoint<Request, Response>
{
    public AppDbContext Db { get; set; } = null!;

    public override void Configure()
    {
        Post("/api/orders");

    }

    public override async Task HandleAsync(Request req, CancellationToken ct)
    {
        var order = new OrderFlow.Domain.Order
        {
            Id = Guid.NewGuid(),
            UserId = Convert.ToInt32(User.Claims.FirstOrDefault(c => c.Type == "UserId")?.Value ?? "0"),
            CustomerName = req.CustomerName,
            Status = OrderStatus.Created
        };

        await Db.Orders.AddAsync(order, ct);
        await Db.SaveChangesAsync(ct);

        await Send.OkAsync(new Response
        {
            OrderId = order.Id,
            Message = "Pedido criado com sucesso!"
        }, cancellation: ct);
    }
}
