# Traefik JWT Extractor for Supabase

A Traefik middleware plugin that extracts JWT tokens from Supabase authentication cookies and adds them to the Authorization header. This middleware is specifically designed to work with Supabase's authentication system, making it easier to integrate Supabase-authenticated applications with other services through Traefik.

## Features

- Extracts JWT tokens from Supabase auth cookies
- Automatically adds Bearer token to Authorization header
- Configurable cookie settings
- Handles base64-encoded JSON payloads
- Secure cookie handling

## Installation

### Static Configuration

Add the plugin to your Traefik static configuration:

```yaml
experimental:
  plugins:
    jwt-extractor:
      moduleName: github.com/willopez/extract-jwt-cookie
      version: v1.0.0
```

### Dynamic Configuration

Configure the middleware in your dynamic configuration:

```yaml
http:
  middlewares:
    supabase-jwt:
      plugin:
        jwt-extractor:
          cookieName: "sb-api-auth-token"
          path: "/"
          ttl: 60
          httpOnly: true
          secure: true
          sameSite: 1
```

## Configuration

Available configuration options with their default values:

```yaml
cookieName: "sb-api-auth-token"  # Name of the Supabase auth cookie
path: "/"                        # Cookie path
domain: ""                       # Cookie domain
ttl: 60                         # Time to live in minutes
httpOnly: true                  # HttpOnly flag
secure: false                   # Set to true if using HTTPS
sameSite: 1                     # SameSite policy (1: Lax mode)
```

## Example Docker Compose

```yaml
version: '3.9'
services:
  traefik:
    image: traefik:v2.11
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--entrypoints.web.address=:80"
      - "--experimental.plugins.jwt-extractor.modulename=github.com/willopez/extract-jwt-cookie"
      - "--experimental.plugins.jwt-extractor.version=v1.0.0"
    ports:
      - "80:80"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock

  api:
    image: your-api-service
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.api.rule=Host(`api.example.com`)"
      - "traefik.http.routers.api.middlewares=supabase-jwt"
      - "traefik.http.middlewares.supabase-jwt.plugin.jwt-extractor.cookieName=sb-api-auth-token"
      - "traefik.http.middlewares.supabase-jwt.plugin.jwt-extractor.secure=true"
```

## How It Works

1. The middleware looks for a Supabase authentication cookie (default: `sb-api-auth-token`)
2. Extracts the base64-encoded JSON payload (format: `base64-<encoded-data>`)
3. Decodes and parses the JSON to find the access token
4. Adds the token to the Authorization header as `Bearer <token>`

## Security Considerations

- Always use HTTPS in production
- Enable the `secure` cookie flag in production
- Consider your SameSite cookie policy needs
- Ensure your Supabase configuration aligns with these settings

## Development

To develop or test the plugin locally:

1. Clone the repository
2. Run tests: `go test ./...`
3. Use Traefik's local plugin development mode

## License

This project is licensed under the MIT License - see the LICENSE file for details.
