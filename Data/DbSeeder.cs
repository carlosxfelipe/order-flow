using OrderFlow.Domain;

namespace OrderFlow.Data;

public static class DbSeeder
{
    public static async Task SeedAsync(AppDbContext dbContext)
    {
        if (!dbContext.Users.Any())
        {
            var adminUser = new User
            {
                Username = "admin",
                PasswordHash = BCrypt.Net.BCrypt.HashPassword("admin123"),
                Role = "Admin"
            };

            var carlosUser = new User
            {
                Username = "carlos",
                PasswordHash = BCrypt.Net.BCrypt.HashPassword("carlos123"),
                Role = "Customer"
            };

            var mariaUser = new User
            {
                Username = "maria",
                PasswordHash = BCrypt.Net.BCrypt.HashPassword("maria123"),
                Role = "Customer"
            };

            var luisUser = new User
            {
                Username = "luis",
                PasswordHash = BCrypt.Net.BCrypt.HashPassword("luis123"),
                Role = "Customer"
            };

            await dbContext.Users.AddRangeAsync(adminUser, carlosUser, mariaUser, luisUser);
            await dbContext.SaveChangesAsync(); // save to get IDs

            if (!dbContext.Orders.Any())
            {
                var order1 = new Order
                {
                    Id = Guid.NewGuid(),
                    UserId = carlosUser.Id,
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
                    UserId = mariaUser.Id,
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
                    UserId = luisUser.Id,
                    CustomerName = "Luís Felipe",
                    Status = OrderStatus.Shipped,
                    Items = new List<OrderItem>
                    {
                        new OrderItem { ProductName = "Coca-Cola Zero Lata", Price = 5.50m, Quantity = 1 }
                    }
                };

                await dbContext.Orders.AddRangeAsync(order1, order2, order3);
                await dbContext.SaveChangesAsync();
            }
        }
    }
}
