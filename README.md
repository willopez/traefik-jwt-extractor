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

Create a static configuration file:

```yaml
# traefik.yml
api:
  insecure: true

experimental:
  plugins:
    jwt-extractor:
      moduleName: github.com/willopez/traefik-jwt-extractor
      version: v1.0.0 # Ensure a matching release/tag exists in your repo

entryPoints:
  web:
    address: ":80"

providers:
  docker: {}
  file:
    filename: /etc/traefik/dynamic_conf.yml
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
```

## Configuration

Available configuration options with their default values:

```yaml
cookieName: "sb-api-auth-token"  # Name of the Supabase auth cookie
```

## Example Docker Compose

Create a dynamic configuration file first:

```yaml
# dynamic_conf.yml
http:
  routers:
    api:
      rule: "Host(`api.example.com`)"
      service: "api-service"
      middlewares:
        - "supabase-jwt"

  services:
    api-service:
      loadBalancer:
        servers:
          - url: "http://api:8080"

  middlewares:
    supabase-jwt:
      plugin:
        jwt-extractor:
          cookieName: "sb-api-auth-token"
```

Then use both configuration files in your Docker Compose:

```yaml
version: '3.9'
services:
  traefik:
    image: traefik:v2.11
    ports:
      - "80:80"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./traefik.yml:/etc/traefik/traefik.yml:ro
      - ./dynamic_conf.yml:/etc/traefik/dynamic_conf.yml:ro

  api:
    image: your-api-service
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
