package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"
)

// Config encapsulates the resources exposed by the registry API.
type Config struct {
	StaticFS fs.FS
	Tokens   map[string]TokenEntry
}

// Server serves the registry HTTP interface backed by static fixtures.
type Server struct {
	cfg       Config
	mux       *http.ServeMux
	tokens    map[string]TokenEntry
	slugIndex map[string]TokenEntry
	jwks      []byte
	jwksURL   string
}

// TokenEntry describes a published token and associated transparency metadata.
type TokenEntry struct {
	URI       string `json:"uri"`
	Type      string `json:"type"`
	Slug      string `json:"slug,omitempty"`
	Filename  string `json:"-"`
	Hash      string `json:"hash"`
	Version   string `json:"version"`
	IssuedAt  string `json:"issued_at"`
	NotBefore string `json:"nbf"`
	ExpiresAt string `json:"exp"`
	Revoked   bool   `json:"revoked"`
}

// DefaultTokens enumerates the static fixture metadata served by the registry.
var DefaultTokens = map[string]TokenEntry{
	"urn:lane2:token:RRMT:EU:PSD3:3.2": {
		URI:       "urn:lane2:token:RRMT:EU:PSD3:3.2",
		Type:      "RRMT",
		Slug:      "eu-psd3-2025",
		Filename:  "rrmt-eu-psd3-2025.json",
		Hash:      "sha256:69da10672d528c5c245d911424159dfa9703843b28e8810a5f42d3386251ed27",
		Version:   "2025.10",
		IssuedAt:  "2025-10-01T00:00:00Z",
		NotBefore: "2025-10-01T00:00:00Z",
		ExpiresAt: "2026-10-01T00:00:00Z",
		Revoked:   false,
	},
	"urn:lane2:token:CORT:VODAFONE.VISA:2025": {
		URI:       "urn:lane2:token:CORT:VODAFONE.VISA:2025",
		Type:      "CORT",
		Slug:      "vodafone-visa-2025",
		Filename:  "cort-vodafone-visa-2025.json",
		Hash:      "sha256:e90a7d9cb49299a8a2809f6e6f8268810c3036a5a50bdb82c03b4d94358196da",
		Version:   "2025-Q4",
		IssuedAt:  "2025-10-01T00:00:00Z",
		NotBefore: "2025-10-01T00:00:00Z",
		ExpiresAt: "2026-04-01T00:00:00Z",
		Revoked:   false,
	},
	"urn:lane2:token:PSRT:VISA:ACQ-123": {
		URI:       "urn:lane2:token:PSRT:VISA:ACQ-123",
		Type:      "PSRT",
		Slug:      "visa-acq-123",
		Filename:  "psrt-visa-acq-123.json",
		Hash:      "sha256:395f95cce27efd422446b973fa06cb2cccc7eeed828ed697daf402859747657b",
		Version:   "2025-01",
		IssuedAt:  "2025-10-01T00:00:00Z",
		NotBefore: "2025-10-01T00:00:00Z",
		ExpiresAt: "2026-01-01T00:00:00Z",
		Revoked:   false,
	},
	"urn:lane2:token:RMT:EU:PSD3:3.2": {
		URI:       "urn:lane2:token:RMT:EU:PSD3:3.2",
		Type:      "RMT",
		Slug:      "eu-psd3-2025",
		Filename:  "rrmt-eu-psd3-2025.json",
		Hash:      "sha256:69da10672d528c5c245d911424159dfa9703843b28e8810a5f42d3386251ed27",
		Version:   "2025.10",
		IssuedAt:  "2025-10-01T00:00:00Z",
		NotBefore: "2025-10-01T00:00:00Z",
		ExpiresAt: "2026-10-01T00:00:00Z",
		Revoked:   false,
	},
	"urn:lane2:token:IMT:EU:SG:2025": {
		URI:       "urn:lane2:token:IMT:EU:SG:2025",
		Type:      "IMT",
		Slug:      "eu-sg-2025",
		Filename:  "imt-eu-sg-2025.json",
		Hash:      "sha256:86273b5161f566be632b7cf629f8c37c46a125fddfe8c2ad0ca8f583182d5fa1",
		Version:   "2025.10",
		IssuedAt:  "2025-10-01T00:00:00Z",
		NotBefore: "2025-10-01T00:00:00Z",
		ExpiresAt: "2026-10-01T00:00:00Z",
		Revoked:   false,
	},
}

// NewServer validates configuration and returns a ready-to-serve API instance.
func NewServer(cfg Config) (*Server, error) {
	if cfg.StaticFS == nil {
		return nil, errors.New("StaticFS is required")
	}
	tokenCatalog := cfg.Tokens
	if len(tokenCatalog) == 0 {
		tokenCatalog = DefaultTokens
	}

	tokens := make(map[string]TokenEntry, len(tokenCatalog))
	slugIndex := make(map[string]TokenEntry, len(tokenCatalog))
	for uri, entry := range tokenCatalog {
		tokens[uri] = entry
		if entry.Type != "" && entry.Slug != "" {
			key := strings.ToLower(entry.Type) + ":" + entry.Slug
			slugIndex[key] = entry
		}
	}

	jwksData, err := fs.ReadFile(cfg.StaticFS, "jwks.json")
	if err != nil {
		return nil, fmt.Errorf("load jwks: %w", err)
	}
	baseURL := strings.TrimSuffix(os.Getenv("RTGF_URL"), "/")
	jwksURL := "/jwks.json"
	if baseURL != "" {
		jwksURL = baseURL + jwksURL
	}

	s := &Server{
		cfg:       Config{StaticFS: cfg.StaticFS, Tokens: tokenCatalog},
		mux:       http.NewServeMux(),
		tokens:    tokens,
		slugIndex: slugIndex,
		jwks:      jwksData,
		jwksURL:   jwksURL,
	}
	s.routes()
	return s, nil
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", s.handleHealth)
	s.mux.HandleFunc("/tokens", s.handleTokenByURI)
	s.mux.Handle("/tokens/", http.HandlerFunc(s.handleTokenByType))
	s.mux.HandleFunc("/catalog", s.handleCatalog)
	s.mux.HandleFunc("/.well-known/rtgf/catalog.json", s.handleCatalog)
	s.mux.HandleFunc("/jwks.json", s.handleJWKS)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleTokenByURI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	uri := strings.TrimSpace(r.URL.Query().Get("uri"))
	if uri == "" {
		http.Error(w, "missing uri query parameter", http.StatusBadRequest)
		return
	}
	entry, ok := s.tokens[uri]
	if !ok {
		http.NotFound(w, r)
		return
	}
	s.serveStaticJSON(w, r, entry)
}

func (s *Server) handleTokenByType(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	trimmed := strings.TrimPrefix(r.URL.Path, "/tokens/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	tokenType := strings.ToLower(parts[0])
	slug := parts[1]
	if slug == "" || strings.Contains(slug, "..") {
		http.Error(w, "invalid token identifier", http.StatusBadRequest)
		return
	}
	entry, err := s.lookupBySlug(tokenType, slug)
	if err != nil {
		if errors.Is(err, errNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.serveStaticJSON(w, r, entry)
}

func (s *Server) handleCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	type issuer struct {
		Iss  string `json:"iss"`
		JWKS string `json:"jwks"`
	}
	type corridor struct {
		ID             string   `json:"id"`
		RequiredTokens []string `json:"requiredTokens"`
	}
	type catalog struct {
		RegistrySnapshotID string     `json:"registrySnapshotId"`
		Issuers            []issuer   `json:"issuers"`
		Corridors          []corridor `json:"corridors"`
	}
	payload := catalog{
		RegistrySnapshotID: "sha256:rtgf-catalog-2025-10-01",
		Issuers: []issuer{
			{Iss: "did:org:rtgf.eu", JWKS: s.jwksURL},
		},
		Corridors: []corridor{
			{ID: "EU:THA:CRAFT-01", RequiredTokens: []string{"RMT", "IMT", "CORT", "PSRT"}},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleJWKS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(s.jwks)
}
func (s *Server) serveStaticJSON(w http.ResponseWriter, r *http.Request, entry TokenEntry) {
	data, err := fs.ReadFile(s.cfg.StaticFS, entry.Filename)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

var errNotFound = errors.New("not found")

func (s *Server) lookupBySlug(tokenType, slug string) (TokenEntry, error) {
	key := tokenType + ":" + slug
	if entry, ok := s.slugIndex[key]; ok {
		return entry, nil
	}
	return TokenEntry{}, errNotFound
}
