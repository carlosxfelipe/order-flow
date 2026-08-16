using FastEndpoints;
using FastEndpoints.Testing;
using FluentAssertions;
using OrderFlow.Features.Orders.AddItem;
using System.Net;
using TUnit.Core;
using OrderFlow.Data;
using Microsoft.Extensions.DependencyInjection;
using System.Net.Http.Json;

namespace OrderFlow.Tests.Features.Orders;

[ClassDataSource<AppFixture>]
public class AddItemTests : IAsyncDisposable
{
    private readonly AppFixture fixture = new();

    public ValueTask DisposeAsync() => fixture.DisposeAsync();


    [Test]
    public async Task AddItem_ReturnsOk()
    {
        // Arrange
        var client = fixture.CreateClient();
        client.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", fixture.GetToken(userId: "2"));

        using var scope = fixture.Services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
        var orderId = db.Orders.First(o => o.UserId == 2 && o.Status == OrderFlow.Domain.OrderStatus.Created).Id;

        var request = new Request
        {
            OrderId = orderId,
            ProductName = "New Product",
            Quantity = 1,
            Price = 50.0m
        };

        // Act
        var httpResponse = await client.PostAsJsonAsync($"/api/orders/{orderId}/items", request);
        var body = await httpResponse.Content.ReadAsStringAsync();

        // Assert
        httpResponse.StatusCode.Should().Be(HttpStatusCode.OK, because: body);
    }
}
