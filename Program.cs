using FastEndpoints;
using FastEndpoints.Swagger;
using Microsoft.EntityFrameworkCore;
using OrderFlow.Data;
using Scalar.AspNetCore;

var builder = WebApplication.CreateBuilder(args);
builder.Services.AddFastEndpoints();
builder.Services.SwaggerDocument(); // Geração do documento OpenAPI
builder.Services.AddDbContext<AppDbContext>(options =>
    options.UseSqlite(builder.Configuration.GetConnectionString("DefaultConnection")));

var app = builder.Build();
app.UseFastEndpoints();

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
});

// Redireciona a raiz "/" para a documentação no Scalar
app.MapGet("/", () => Results.Redirect("/scalar")).ExcludeFromDescription();

app.Run();
