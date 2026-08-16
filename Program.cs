using FastEndpoints;
using FastEndpoints.Swagger;
using FastEndpoints.Security;
using Microsoft.EntityFrameworkCore;
using OrderFlow.Data;
using Scalar.AspNetCore;
using Microsoft.AspNetCore.RateLimiting;

var builder = WebApplication.CreateBuilder(args);

// Lê a chave do appsettings.json ou Variável de Ambiente (ex: no Render). Falha na inicialização se a chave não existir.
var jwtSecret = builder.Configuration["JwtSecret"] ?? throw new InvalidOperationException("ALERTA DE SEGURANÇA: A variável JwtSecret não foi configurada no ambiente!");
builder.Services.AddAuthenticationJwtBearer(s => s.SigningKey = jwtSecret);
builder.Services.AddAuthorization();

// Configura limite de requisições para proteger a API (Rate Limiting)
builder.Services.AddRateLimiter(options =>
{
    options.RejectionStatusCode = 429; // Retorna 'Too Many Requests' se passar do limite
    options.AddFixedWindowLimiter("Global", opt =>
    {
        opt.Window = TimeSpan.FromMinutes(1);
        opt.PermitLimit = 100; // Máximo de 100 requisições por minuto
        opt.QueueLimit = 0;
    });
});

builder.Services.AddFastEndpoints();
builder.Services.SwaggerDocument(o =>
{
    o.EnableJWTBearerAuth = true;
}); // Geração do documento OpenAPI
builder.Services.AddDbContext<AppDbContext>(options =>
    options.UseSqlite(builder.Configuration.GetConnectionString("DefaultConnection")));

var app = builder.Build();
app.UseRateLimiter(); // Ativa o middleware de Rate Limiting
app.UseAuthentication();
app.UseAuthorization();
app.UseFastEndpoints(c =>
{
    c.Endpoints.Configurator = ep =>
    {
        // Aplica a política "Global" de limite em todos os endpoints
        ep.Options(b => b.RequireRateLimiting("Global"));
    };

    // Rejeita a requisição (erro 400) se o JSON contiver campos extras não esperados
    c.Serializer.Options.UnmappedMemberHandling = System.Text.Json.Serialization.JsonUnmappedMemberHandling.Disallow;
});

using (var scope = app.Services.CreateScope())
{
    var dbContext = scope.ServiceProvider.GetRequiredService<AppDbContext>();
    await dbContext.Database.MigrateAsync();

    // Rodamos o Seed apenas em ambiente de Desenvolvimento
    if (app.Environment.IsDevelopment())
    {
        await DbSeeder.SeedAsync(dbContext);
    }
}

// Disponibiliza o JSON do OpenAPI gerado pelo FastEndpoints
app.UseOpenApi(c => c.Path = "/openapi/v1.json");

// Adiciona o Scalar UI apontando para o JSON gerado
app.MapScalarApiReference(options =>
{
    options.WithOpenApiRoutePattern("/openapi/v1.json");
    options.WithTheme(ScalarTheme.Default);
    options.WithClassicLayout();

    // Workaround temporário para dar um respiro no rodapé do layout Classic
    options.WithCustomCss("body { padding-bottom: 80px !important; }");
});

// Redireciona a raiz "/" para a documentação no Scalar
app.MapGet("/", () => Results.Redirect("/scalar/v1")).ExcludeFromDescription();

app.Run();
