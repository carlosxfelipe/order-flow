using FastEndpoints;
using FastEndpoints.Testing;
using FluentAssertions;
using OrderFlow.Features.Orders.PayOrder;
using System.Net;
using TUnit.Core;
using OrderFlow.Data;
using Microsoft.Extensions.DependencyInjection;

namespace OrderFlow.Tests.Features.Orders;

[ClassDataSource<AppFixture>]
public class PayOrderTests : IAsyncDisposable
{
    private readonly AppFixture fixture = new();

    public ValueTask DisposeAsync() => fixture.DisposeAsync();


    [Test]
    public async Task PayOrder_ReturnsOk()
    {
        // Arrange
        var client = fixture.CreateClient();
        client.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", fixture.GetToken(userId: "2"));
        
        using var scope = fixture.Services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
        var orderId = db.Orders.First(o => o.UserId == 2 && o.Status == OrderFlow.Domain.OrderStatus.Created).Id;

        var request = new Request { OrderId = orderId };

        // Act
        var (response, _) = await client.POSTAsync<PayOrderEndpoint, Request, EmptyResponse>(request);

        // Assert
        response.StatusCode.Should().Be(HttpStatusCode.OK);
    }
}
