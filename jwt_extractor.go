package traefik_jwt_extractor

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

// Config holds the plugin configuration.
type Config struct {
    CookieName string        `json:"cookieName,omitempty"`
    TTL        int           `json:"ttl"`
    Path       string        `json:"path"`
    Domain     string        `json:"domain"`
    HttpOnly   bool          `json:"httpOnly"`
    Secure     bool          `json:"secure"`
    SameSite   http.SameSite `json:"sameSite"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
    return &Config{
        CookieName: "sb-api-auth-token",
        TTL:        60,
        Path:       "/",
        HttpOnly:   true,
        Secure:     false,
        SameSite:   http.SameSiteLaxMode,
    }
}

type JwtExtractor struct {
    next   http.Handler
    name   string
    config *Config
}

// New creates a new middleware instance
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

// ServeHTTP implements the middleware logic:
// 1. Extracts the cookie value
// 2. Removes the "base64-" prefix
// 3. Decodes the base64 content
// 4. Parses the JSON payload
// 5. Sets the Authorization header with the access token
func (a *JwtExtractor) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    // Extract the named cookie from the request
    cookie, err := req.Cookie(a.config.CookieName)
    if err != nil {
        http.Error(rw, "Cookie not found", http.StatusUnauthorized)
        return
    }

    // Cookie value is expected to be in format: "base64-<base64-encoded-json>"
    cookieValue := strings.TrimPrefix(cookie.Value, "base64-")
    if cookieValue == cookie.Value {
        http.Error(rw, "Cookie value does not start with 'base64-'", http.StatusBadRequest)
        return
    }

    // Decode the base64 content to get the JSON payload
    decoded, err := base64.StdEncoding.DecodeString(cookieValue)
    if err != nil {
        http.Error(rw, "Failed to decode cookie value", http.StatusBadRequest)
        return
    }

    // Parse the JSON payload which should contain an access_token field
    var data map[string]interface{}
    if err := json.Unmarshal(decoded, &data); err != nil {
        http.Error(rw, "Failed to parse JSON", http.StatusBadRequest)
        return
    }

    // Extract and validate the access_token from the JSON payload
    accessToken, ok := data["access_token"].(string)
    if !ok {
        http.Error(rw, "No valid access_token found", http.StatusUnauthorized)
        return
    }

    // Set the Authorization header with the Bearer token for downstream services
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

    // Continue the middleware chain
    a.next.ServeHTTP(rw, req)
}