using FastEndpoints;
using FastEndpoints.Testing;
using FluentAssertions;
using OrderFlow.Features.Orders.CreateOrder;
using System.Net;
using TUnit.Core;

namespace OrderFlow.Tests.Features.Orders;

[ClassDataSource<AppFixture>]
public class CreateOrderTests : IAsyncDisposable
{
    private readonly AppFixture fixture = new();

    public ValueTask DisposeAsync() => fixture.DisposeAsync();


    [Test]
    public async Task CreateOrder_ReturnsOkAndOrderId()
    {
        // Arrange
        var client = fixture.CreateClient();
        client.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", fixture.GetToken());
        var request = new Request
        {
            CustomerName = "Test User"
        };

        // Act
        var (response, result) = await client.POSTAsync<CreateOrderEndpoint, Request, Response>(request);

        // Assert
        response.StatusCode.Should().Be(HttpStatusCode.OK);
        result.OrderId.Should().NotBeEmpty();
    }
}
