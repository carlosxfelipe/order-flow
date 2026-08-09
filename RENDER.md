# Render Deployment Guide

This guide covers important environment variables to configure when deploying this application to [Render](https://render.com/).

## 1. Environment and Seeding (SQLite)

By default, when this application runs inside a Docker container on Render, it defaults to the **Production** environment. 

Because of the following logic in `Program.cs`, the database migrations will run (`MigrateAsync`), but the initial data seed will be skipped:

```csharp
// Seed runs only in Development environment
if (app.Environment.IsDevelopment())
{
    await DbSeeder.SeedAsync(dbContext);
}
```

If you want Render to execute the `DbSeeder` and populate your database with initial data (useful for testing the API), you must add the following **Environment Variable** in the Render dashboard:

- **Key:** `ASPNETCORE_ENVIRONMENT`
- **Value:** `Development`

## 2. Future Migration to PostgreSQL (Neon DB)

The application is currently configured to use an ephemeral SQLite database, which will be wiped every time the Render instance restarts or sleeps. 

When you are ready to migrate to a persistent PostgreSQL database (such as Neon DB), you will need to override the default connection string by adding the following Environment Variable in Render:

- **Key:** `ConnectionStrings__DefaultConnection`
- **Value:** `Host=ep-cool-...neon.tech;Database=...` *(Your Postgres connection string)*

This automatically overrides the `DefaultConnection` defined in your `appsettings.json`.
