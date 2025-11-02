package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestHealthz(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content type %q", ct)
	}
}

func TestTokenLookupByURI(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/tokens?uri=urn:lane2:token:RRMT:EU:PSD3:3.2", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"type":"RRMT"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestTokenLookupByTypeAndSlug(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/tokens/cort/cort-slug", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"type":"CORT"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestTokenLookupNotFound(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/tokens?uri=urn:lane2:token:RRMT:UNKNOWN", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestTokenLookupMethodNotAllowed(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/tokens", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestTokenLookupInvalidSlug(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/tokens/cort/bad..slug", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCatalogEndpoint(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var payload struct {
		RegistrySnapshotID string `json:"registrySnapshotId"`
		Issuers            []struct {
			Iss  string `json:"iss"`
			JWKS string `json:"jwks"`
		} `json:"issuers"`
		Corridors []struct {
			ID             string   `json:"id"`
			RequiredTokens []string `json:"requiredTokens"`
		} `json:"corridors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal catalog: %v", err)
	}
	if payload.RegistrySnapshotID == "" {
		t.Fatalf("expected registry snapshot id")
	}
	if len(payload.Issuers) == 0 || payload.Issuers[0].JWKS == "" {
		t.Fatalf("expected issuer jwks")
	}
	if len(payload.Corridors) == 0 || payload.Corridors[0].ID != "EU:THA:CRAFT-01" {
		t.Fatalf("unexpected corridors: %+v", payload.Corridors)
	}
}

func TestCatalogEndpointRespectsBaseURL(t *testing.T) {
	t.Setenv("RTGF_URL", "https://registry.example.com")
	fsys := fstest.MapFS{
		"jwks.json":  {Data: []byte(`{"keys":[]}`)},
		"token.json": {Data: []byte(`{"type":"RRMT"}`)},
	}
	entries := map[string]TokenEntry{
		"urn:test:token": {
			URI:       "urn:test:token",
			Type:      "RRMT",
			Slug:      "token",
			Filename:  "token.json",
			Hash:      "sha256:test",
			Version:   "v1",
			IssuedAt:  "2025-01-01T00:00:00Z",
			NotBefore: "2025-01-01T00:00:00Z",
			ExpiresAt: "2026-01-01T00:00:00Z",
			Revoked:   false,
		},
	}
	server, err := NewServer(Config{
		StaticFS: fsys,
		Tokens:   entries,
	})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
	var payload struct {
		Issuers []struct {
			JWKS string `json:"jwks"`
		} `json:"issuers"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal catalog: %v", err)
	}
	if len(payload.Issuers) == 0 || payload.Issuers[0].JWKS != "https://registry.example.com/jwks.json" {
		t.Fatalf("expected JWKS URL with base, got %+v", payload.Issuers)
	}
}

func TestJWKSHandler(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/jwks.json", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content type %s", ct)
	}
	if !strings.Contains(rec.Body.String(), `"keys"`) {
		t.Fatalf("expected keys in JWKS response: %s", rec.Body.String())
	}
}

func TestJWKSMethodNotAllowed(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/jwks.json", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	fsys := fstest.MapFS{
		"rrmt-eu-psd3-2025.json": {Data: []byte(`{"type":"RRMT","version":"v1","nbf":"2025-01-01T00:00:00Z","exp":"2026-01-01T00:00:00Z","revoked":false}`)},
		"cort-example.json":      {Data: []byte(`{"type":"CORT"}`)},
		"jwks.json":              {Data: []byte(`{"keys":[{"kty":"OKP","crv":"Ed25519","kid":"test","x":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}]}`)},
	}
	entries := map[string]TokenEntry{
		"urn:lane2:token:RRMT:EU:PSD3:3.2": {
			URI:       "urn:lane2:token:RRMT:EU:PSD3:3.2",
			Type:      "RRMT",
			Slug:      "rrmt-slug",
			Filename:  "rrmt-eu-psd3-2025.json",
			Hash:      "sha256:test",
			Version:   "v1",
			IssuedAt:  "2025-01-01T00:00:00Z",
			NotBefore: "2025-01-01T00:00:00Z",
			ExpiresAt: "2026-01-01T00:00:00Z",
			Revoked:   false,
		},
		"urn:lane2:token:CORT:TEST": {
			URI:       "urn:lane2:token:CORT:TEST",
			Type:      "CORT",
			Slug:      "cort-slug",
			Filename:  "cort-example.json",
			Hash:      "sha256:test2",
			Version:   "v1",
			IssuedAt:  "2025-01-01T00:00:00Z",
			NotBefore: "2025-01-01T00:00:00Z",
			ExpiresAt: "2026-01-01T00:00:00Z",
			Revoked:   false,
		},
	}
	server, err := NewServer(Config{
		StaticFS: fsys,
		Tokens:   entries,
	})
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}
	return server
}
