using FastEndpoints;
using OrderFlow.Data;
using OrderFlow.Domain;

namespace OrderFlow.Features.Orders.RemoveItem;

public class RemoveItemEndpoint : Endpoint<Request, Response>
{
    public override void Configure()
    {
        Delete("/api/orders/{orderId}/items/{itemId}");
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
            AddError("Não é possível remover itens. O pedido não está mais em estado de criação.");
            ThrowIfAnyErrors();
            return;
        }

        var item = order.Items.FirstOrDefault(i => i.Id == req.ItemId);
        if (item == null)
        {
            await Send.NotFoundAsync(cancellation: ct);
            return;
        }

        order.Items.Remove(item);

        await Send.OkAsync(new Response
        {
            Message = "Item removido com sucesso!"
        }, cancellation: ct);
    }
}
