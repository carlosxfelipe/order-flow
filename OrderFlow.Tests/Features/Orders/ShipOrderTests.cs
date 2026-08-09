using FastEndpoints;
using FastEndpoints.Testing;
using FluentAssertions;
using OrderFlow.Features.Orders.ShipOrder;
using System.Net;
using TUnit.Core;
using OrderFlow.Data;
using Microsoft.Extensions.DependencyInjection;

namespace OrderFlow.Tests.Features.Orders;

[ClassDataSource<AppFixture>]
public class ShipOrderTests : IAsyncDisposable
{
    private readonly AppFixture fixture = new();

    public ValueTask DisposeAsync() => fixture.DisposeAsync();


    [Test]
    public async Task ShipOrder_ReturnsOk()
    {
        // Arrange
        var client = fixture.CreateClient();
        client.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", fixture.GetToken(role: "Admin"));
        
        using var scope = fixture.Services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
        var orderId = db.Orders.First(o => o.Status == OrderFlow.Domain.OrderStatus.Paid).Id;

        var request = new Request { OrderId = orderId };

        // Act
        var (response, _) = await client.POSTAsync<ShipOrderEndpoint, Request, EmptyResponse>(request);

        // Assert
        response.StatusCode.Should().Be(HttpStatusCode.OK);
    }
}
