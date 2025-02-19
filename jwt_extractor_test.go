package traefik_jwt_extractor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockHandler struct {
	headers http.Header
}

func newMockHandler() *mockHandler {
	return &mockHandler{
		headers: make(http.Header),
	}
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.headers = r.Header
}

func TestCreateConfig(t *testing.T) {
	config := CreateConfig()
	if config == nil {
		t.Fatal("Expected config to not be nil")
	}
	if config.CookieName != "sb-api-auth-token" {
		t.Errorf("Expected cookie name to be 'sb-api-auth-token', got %s", config.CookieName)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "Valid config",
			config: &Config{
				CookieName: "test_cookie",
			},
			wantErr: false,
		},
		{
			name: "Empty cookie name",
			config: &Config{
				CookieName: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			middleware, err := New(ctx, handler, tt.config, "test")
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && middleware == nil {
				t.Error("Expected middleware to not be nil")
			}
		})
	}
}

func TestExtractJwtCookie(t *testing.T) {
	tests := []struct {
		name           string
		cookieValue    string
		expectedStatus int
		expectedToken  string
	}{
		{
			name: "valid cookie",
			cookieValue: func() string {
				payload := map[string]string{"access_token": "test-jwt-token"}
				jsonData, _ := json.Marshal(payload)
				return "base64-" + base64.StdEncoding.EncodeToString(jsonData)
			}(),
			expectedStatus: http.StatusOK,
			expectedToken:  "Bearer test-jwt-token",
		},
		{
			name:           "missing cookie",
			cookieValue:    "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid base64 prefix",
			cookieValue:    "invalid-prefix-xyz",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid base64 content",
			cookieValue:    "base64-not-base64-content",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid json payload",
			cookieValue: func() string {
				return "base64-" + base64.StdEncoding.EncodeToString([]byte("invalid json"))
			}(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing access token",
			cookieValue: func() string {
				payload := map[string]string{"other_field": "value"}
				jsonData, _ := json.Marshal(payload)
				return "base64-" + base64.StdEncoding.EncodeToString(jsonData)
			}(),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mock := newMockHandler()
			ctx := context.Background()
			middleware, _ := New(ctx, mock, CreateConfig(), "test")

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{
					Name:  "sb-api-auth-token",
					Value: tt.cookieValue,
				})
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			middleware.ServeHTTP(rr, req)

			// Assert response status
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// For successful cases, check the Authorization header
			if tt.expectedStatus == http.StatusOK {
				authHeader := mock.headers.Get("Authorization")
				if authHeader != tt.expectedToken {
					t.Errorf("expected Authorization header %s, got %s", tt.expectedToken, authHeader)
				}
			}
		})
	}
}
