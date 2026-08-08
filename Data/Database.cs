using OrderFlow.Domain;

namespace OrderFlow.Data;

public static class Database
{
    // Simple in-memory storage for our orders
    public static readonly Dictionary<Guid, Order> Orders = new();
}
