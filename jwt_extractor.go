package traefik_jwt_extractor

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

// Config holds the plugin configuration for JWT extraction.
type Config struct {
    // CookieName is the name of the cookie containing the JWT token
    CookieName string `json:"cookieName,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
    return &Config{
        CookieName: "sb-api-auth-token",
    }
}

// JwtExtractor implements a Traefik middleware that extracts a JWT token
// from a cookie and adds it to the Authorization header.
type JwtExtractor struct {
    next   http.Handler // The next handler in the middleware chain
    name   string      // Name of the middleware instance
    config *Config     // Configuration for this middleware
}

// New creates and validates a new JWT extractor middleware instance.
// It ensures the configuration is valid before creating the middleware.
// Returns an error if the configuration is invalid.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
    if config.CookieName == "" {
        return nil, fmt.Errorf("cookieName cannot be empty")
    }

    return &JwtExtractor{
        next:   next,
        name:   name,
        config: config,
    }, nil
}

// ServeHTTP processes the request by:
// 1. Extracting a cookie containing a base64-encoded JSON payload
// 2. Decoding the base64 content
// 3. Parsing the JSON to find an access_token
// 4. Adding the token to the Authorization header
//
// The cookie value must be in the format: "base64-<base64-encoded-json>"
// The JSON payload must contain an "access_token" field
// If successful, adds "Authorization: Bearer <token>" header
func (a *JwtExtractor) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    // Extract the named cookie from the request
    cookie, err := req.Cookie(a.config.CookieName)
    if err != nil {
        http.Error(rw, "Cookie not found", http.StatusUnauthorized)
        return
    }

    // Extract and validate the base64 prefix
    // Cookie format must be: base64-<actual-base64-content>
    cookieValue := strings.TrimPrefix(cookie.Value, "base64-")
    if cookieValue == cookie.Value {
        http.Error(rw, "Cookie value does not start with 'base64-'", http.StatusBadRequest)
        return
    }

    // Decode and validate the base64 content
    decoded, err := base64.StdEncoding.DecodeString(cookieValue)
    if err != nil {
        http.Error(rw, "Failed to decode cookie value", http.StatusBadRequest)
        return
    }

    // Parse and validate the JSON structure
    var data map[string]interface{}
    if err := json.Unmarshal(decoded, &data); err != nil {
        http.Error(rw, "Failed to parse JSON", http.StatusBadRequest)
        return
    }

    // Extract and validate the JWT access token
    accessToken, ok := data["access_token"].(string)
    if !ok {
        http.Error(rw, "No valid access_token found", http.StatusUnauthorized)
        return
    }

    // Add the Bearer token to the Authorization header
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

    // Continue processing the request
    a.next.ServeHTTP(rw, req)
}