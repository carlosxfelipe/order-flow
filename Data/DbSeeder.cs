using OrderFlow.Domain;

namespace OrderFlow.Data;

public static class DbSeeder
{
    public static async Task SeedAsync(AppDbContext dbContext)
    {
        // Se já tiver qualquer pedido, não fazemos o seed
        if (dbContext.Orders.Any())
        {
            return;
        }

        var order1 = new Order
        {
            Id = Guid.NewGuid(),
            CustomerName = "Carlos Felipe",
            Status = OrderStatus.Created,
            Items = new List<OrderItem>
            {
                new OrderItem { ProductName = "Coca-Cola Lata 350ml", Price = 5.50m, Quantity = 2 },
                new OrderItem { ProductName = "Hambúrguer Clássico", Price = 25.00m, Quantity = 1 }
            }
        };

        var order2 = new Order
        {
            Id = Guid.NewGuid(),
            CustomerName = "Maria Isabelle",
            Status = OrderStatus.Paid,
            Items = new List<OrderItem>
            {
                new OrderItem { ProductName = "Coca-Cola 2L", Price = 12.00m, Quantity = 1 },
                new OrderItem { ProductName = "Pizza de Calabresa", Price = 45.00m, Quantity = 1 }
            }
        };

        var order3 = new Order
        {
            Id = Guid.NewGuid(),
            CustomerName = "Luís Felipe",
            Status = OrderStatus.Shipped,
            Items = new List<OrderItem>
            {
                new OrderItem { ProductName = "Coca-Cola Zero Lata", Price = 5.50m, Quantity = 1 }
            }
        };

        await dbContext.Orders.AddRangeAsync(order1, order2, order3);

        if (!dbContext.Users.Any())
        {
            var adminUser = new User
            {
                Username = "admin",
                // admin123
                PasswordHash = BCrypt.Net.BCrypt.HashPassword("admin123")
            };
            await dbContext.Users.AddAsync(adminUser);
        }

        await dbContext.SaveChangesAsync();
    }
}
