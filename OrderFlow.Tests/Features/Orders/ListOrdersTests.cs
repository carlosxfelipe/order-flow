using FastEndpoints;
using FastEndpoints.Testing;
using FluentAssertions;
using OrderFlow.Features.Orders.ListOrders;
using System.Net;
using TUnit.Core;

namespace OrderFlow.Tests.Features.Orders;

[ClassDataSource<AppFixture>]
public class ListOrdersTests : IAsyncDisposable
{
    private readonly AppFixture fixture = new();

    public ValueTask DisposeAsync() => fixture.DisposeAsync();


    [Test]
    public async Task ListOrders_ReturnsOkAndList()
    {
        // Arrange
        var client = fixture.CreateClient();
        client.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", fixture.GetToken(userId: "2"));
        
        // Act
        var (response, result) = await client.GETAsync<ListOrdersEndpoint, Response>();

        // Assert
        response.StatusCode.Should().Be(HttpStatusCode.OK);
        result.Should().NotBeNull();
        result.Orders.Should().NotBeEmpty();
    }
}
