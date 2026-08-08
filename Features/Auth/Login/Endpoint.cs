using FastEndpoints;
using FastEndpoints.Security;
using Microsoft.EntityFrameworkCore;
using OrderFlow.Data;

namespace OrderFlow.Features.Auth.Login;

public class LoginEndpoint(AppDbContext dbContext) : Endpoint<LoginRequest, LoginResponse>
{
    public override void Configure()
    {
        Post("/api/auth/login");
        AllowAnonymous();
        Summary(s =>
        {
            s.Summary = "Authenticates a user";
            s.Description = "Validates the user's credentials and returns a JWT Bearer token.";
        });
    }

    public override async Task HandleAsync(LoginRequest req, CancellationToken ct)
    {
        var user = await dbContext.Users.FirstOrDefaultAsync(u => u.Username == req.Username, ct);

        if (user is null || !BCrypt.Net.BCrypt.Verify(req.Password, user.PasswordHash))
        {
            ThrowError("Invalid username or password.");
            return; // required for analyzer
        }

        var jwtToken = JwtBearer.CreateToken(
            o =>
            {
                o.SigningKey = "EstaEumaChaveSuperSecretaParaOOrderFlow123!@#";
                o.ExpireAt = DateTime.UtcNow.AddDays(1);
                o.User.Claims.Add(("Username", user.Username));
                o.User.Claims.Add(("UserId", user.Id.ToString()));
                o.User.Roles.Add(user.Role);
            });

        await Send.OkAsync(new LoginResponse
        {
            Token = jwtToken
        }, cancellation: ct);
    }
}
