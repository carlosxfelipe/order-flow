using FastEndpoints;
using FastEndpoints.Testing;
using FluentAssertions;
using OrderFlow.Features.Auth.Login;
using System.Net;
using TUnit.Core;

namespace OrderFlow.Tests.Features.Auth;

[ClassDataSource<AppFixture>]
public class LoginTests : IAsyncDisposable
{
    private readonly AppFixture fixture = new();

    public ValueTask DisposeAsync() => fixture.DisposeAsync();


    [Test]
    public async Task Login_WithValidCredentials_ReturnsOkAndToken()
    {
        // Arrange
        var client = fixture.CreateClient();
        var request = new LoginRequest
        {
            Username = "admin",
            Password = "admin123"
        };

        // Act
        var (response, result) = await client.POSTAsync<LoginEndpoint, LoginRequest, LoginResponse>(request);

        // Assert
        response.StatusCode.Should().Be(HttpStatusCode.OK);
        result.Token.Should().NotBeEmpty();
    }

    [Test]
    public async Task Login_WithInvalidCredentials_ReturnsBadRequest()
    {
        // Arrange
        var client = fixture.CreateClient();
        var request = new LoginRequest
        {
            Username = "admin",
            Password = "wrongpassword"
        };

        // Act
        var (response, _) = await client.POSTAsync<LoginEndpoint, LoginRequest, EmptyResponse>(request);

        // Assert
        response.StatusCode.Should().Be(HttpStatusCode.BadRequest);
    }
}
