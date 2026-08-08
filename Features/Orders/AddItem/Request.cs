namespace OrderFlow.Features.Orders.AddItem;

public class Request
{
    public Guid OrderId { get; set; }
    public string ProductName { get; set; } = string.Empty;
    public decimal Price { get; set; }
    public int Quantity { get; set; }
}
