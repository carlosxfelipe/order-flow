using FastEndpoints;
using Microsoft.EntityFrameworkCore;
using OrderFlow.Data;
using OrderFlow.Domain;

namespace OrderFlow.Features.Orders.PayOrder;

public class PayOrderEndpoint : Endpoint<Request, Response>
{
    public AppDbContext Db { get; set; } = null!;

    public override void Configure()
    {
        Post("/api/orders/{orderId}/pay");

    }

    public override async Task HandleAsync(Request req, CancellationToken ct)
    {
        var order = await Db.Orders.Include(o => o.Items).FirstOrDefaultAsync(o => o.Id == req.OrderId, ct);
        if (order == null)
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
        await Db.SaveChangesAsync(ct);

        await Send.OkAsync(new Response
        {
            Message = "Pedido pago com sucesso! Refrescando com uma Coca-Cola bem gelada!"
        }, cancellation: ct);
    }
}
