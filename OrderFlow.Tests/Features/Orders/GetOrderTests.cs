using FastEndpoints;
using FastEndpoints.Testing;
using FluentAssertions;
using OrderFlow.Features.Orders.GetOrder;
using System.Net;
using TUnit.Core;
using OrderFlow.Data;
using Microsoft.Extensions.DependencyInjection;

namespace OrderFlow.Tests.Features.Orders;

[ClassDataSource<AppFixture>]
public class GetOrderTests : IAsyncDisposable
{
    private readonly AppFixture fixture = new();

    public ValueTask DisposeAsync() => fixture.DisposeAsync();


    [Test]
    public async Task GetOrder_ReturnsOkAndOrderDetails()
    {
        // Arrange
        var client = fixture.CreateClient();
        client.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", fixture.GetToken());
        
        using var scope = fixture.Services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
        var orderId = db.Orders.First().Id;

        var request = new Request { Id = orderId };

        // Act
        var (response, result) = await client.GETAsync<GetOrderEndpoint, Request, Response>(request);

        // Assert
        response.StatusCode.Should().Be(HttpStatusCode.OK);
        result.Id.Should().Be(orderId);
    }
}
