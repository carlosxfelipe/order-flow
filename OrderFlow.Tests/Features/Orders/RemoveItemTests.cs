using Microsoft.EntityFrameworkCore;
using FastEndpoints;
using FastEndpoints.Testing;
using FluentAssertions;
using OrderFlow.Features.Orders.RemoveItem;
using System.Net;
using TUnit.Core;
using OrderFlow.Data;
using Microsoft.Extensions.DependencyInjection;

namespace OrderFlow.Tests.Features.Orders;

[ClassDataSource<AppFixture>]
public class RemoveItemTests : IAsyncDisposable
{
    private readonly AppFixture fixture = new();

    public ValueTask DisposeAsync() => fixture.DisposeAsync();


    [Test]
    public async Task RemoveItem_ReturnsOk()
    {
        // Arrange
        var client = fixture.CreateClient();
        client.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", fixture.GetToken(userId: "2"));
        
        using var scope = fixture.Services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
        
        // Find an order item that belongs to user 2's created order
        var order = db.Orders.Include(o => o.Items).First(o => o.UserId == 2 && o.Status == OrderFlow.Domain.OrderStatus.Created && o.Items.Any());
        var itemId = order.Items.First().Id;

        var request = new Request 
        { 
            OrderId = order.Id,
            ItemId = itemId
        };

        // Act
        var (response, _) = await client.DELETEAsync<RemoveItemEndpoint, Request, EmptyResponse>(request);

        // Assert
        response.StatusCode.Should().Be(HttpStatusCode.OK);
    }
}
