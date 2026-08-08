namespace OrderFlow.Domain;

public enum OrderStatus
{
    Created,
    Paid,
    Shipped,
    Cancelled
}

public class Order
{
    public Guid Id { get; set; }
    public string CustomerName { get; set; } = string.Empty;
    public OrderStatus Status { get; set; } = OrderStatus.Created;
    public List<OrderItem> Items { get; set; } = new();

    public decimal TotalAmount => Items.Sum(i => i.Price * i.Quantity);
}
