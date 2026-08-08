namespace OrderFlow.Features.Orders.RemoveItem;

public class Request
{
    public Guid OrderId { get; set; }
    public Guid ItemId { get; set; }
}
