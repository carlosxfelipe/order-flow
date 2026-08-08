using OrderFlow.Domain;

namespace OrderFlow.Features.Orders.GetOrder;

public class Response
{
    public Guid Id { get; set; }
    public string CustomerName { get; set; } = string.Empty;
    public string Status { get; set; } = string.Empty;
    public decimal TotalAmount { get; set; }
    public List<OrderItem> Items { get; set; } = new();
    public string Message { get; set; } = string.Empty;
}
