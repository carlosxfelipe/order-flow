namespace OrderFlow.Features.Orders.CreateOrder;

public class Response
{
    public Guid OrderId { get; set; }
    public string Message { get; set; } = string.Empty;
}
