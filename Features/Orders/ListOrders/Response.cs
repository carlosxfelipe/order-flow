namespace OrderFlow.Features.Orders.ListOrders;

public class Response
{
    public List<OrderSummary> Orders { get; set; } = new();
}

public class OrderSummary
{
    public Guid Id { get; set; }
    public string CustomerName { get; set; } = string.Empty;
    public string Status { get; set; } = string.Empty;
    public decimal TotalAmount { get; set; }
}
