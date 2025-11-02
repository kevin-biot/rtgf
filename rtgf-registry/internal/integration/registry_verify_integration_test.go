package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kevin-biot/rtgf/rtgf-registry/internal/api"
	"github.com/kevin-biot/rtgf/rtgf-registry/internal/verify"
	verifylib "github.com/kevin-biot/rtgf/rtgf-verify-lib"
)

func TestVerifyEndpointHappyPath(t *testing.T) {
	server := newIntegrationServer(t, defaultFS())
	defer server.Close()

	resp := doVerifyRequest(t, server, verifyPayload{
		RMT:  "urn:lane2:token:RMT:EU:PSD3:3.2",
		IMT:  "urn:lane2:token:IMT:EU:SG:2025",
		CORT: "urn:lane2:token:CORT:VODAFONE.VISA:2025",
		PSRT: "urn:lane2:token:PSRT:VISA:ACQ-123",
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
	var body struct {
		Valid  bool   `json:"valid"`
		Reason string `json:"reason"`
	}
	decodeBody(t, resp, &body)
	if !body.Valid {
		t.Fatalf("expected valid response, got reason %s", body.Reason)
	}
}

func TestVerifyEndpointRevokedToken(t *testing.T) {
	fs := defaultFS()
	fs["rrmt-eu-psd3-2025.json"] = &fstest.MapFile{
		Data: []byte(`{"type":"RRMT","nbf":"2000-01-01T00:00:00Z","exp":"2100-01-01T00:00:00Z","revoked":true}`),
	}
	server := newIntegrationServer(t, fs)
	defer server.Close()

	resp := doVerifyRequest(t, server, verifyPayload{
		RMT:  "urn:lane2:token:RMT:EU:PSD3:3.2",
		IMT:  "urn:lane2:token:IMT:EU:SG:2025",
		CORT: "urn:lane2:token:CORT:VODAFONE.VISA:2025",
		PSRT: "urn:lane2:token:PSRT:VISA:ACQ-123",
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
	var body struct {
		Valid  bool   `json:"valid"`
		Reason string `json:"reason"`
	}
	decodeBody(t, resp, &body)
	if body.Valid {
		t.Fatalf("expected invalid response for revoked token")
	}
	if want := "token_revoked"; body.Reason == "" || !strings.Contains(body.Reason, want) {
		t.Fatalf("expected reason containing %q, got %s", want, body.Reason)
	}
}

func TestVerifyEndpointInvalidType(t *testing.T) {
	fs := defaultFS()
	fs["rrmt-eu-psd3-2025.json"] = &fstest.MapFile{
		Data: []byte(`{"type":"CORT","nbf":"2000-01-01T00:00:00Z","exp":"2100-01-01T00:00:00Z","revoked":false}`),
	}
	server := newIntegrationServer(t, fs)
	defer server.Close()

	resp := doVerifyRequest(t, server, verifyPayload{
		RMT:  "urn:lane2:token:RMT:EU:PSD3:3.2",
		IMT:  "urn:lane2:token:IMT:EU:SG:2025",
		CORT: "urn:lane2:token:CORT:VODAFONE.VISA:2025",
		PSRT: "urn:lane2:token:PSRT:VISA:ACQ-123",
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
	var body struct {
		Valid  bool   `json:"valid"`
		Reason string `json:"reason"`
	}
	decodeBody(t, resp, &body)
	if body.Valid {
		t.Fatalf("expected invalid response for type mismatch")
	}
	if body.Reason != "invalid_rrmt" {
		t.Fatalf("expected invalid_rrmt reason, got %s", body.Reason)
	}
}

func TestJWKSRotation(t *testing.T) {
	fs := defaultFS()
	server1 := newIntegrationServer(t, fs)
	t.Cleanup(server1.Close)

	resp1, err := http.Get(server1.URL + "/jwks.json")
	if err != nil {
		t.Fatalf("fetch jwks: %v", err)
	}
	defer resp1.Body.Close()
	var jwks1 struct {
		Keys []struct {
			Kid string `json:"kid"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp1.Body).Decode(&jwks1); err != nil {
		t.Fatalf("decode jwks1: %v", err)
	}
	if len(jwks1.Keys) == 0 || jwks1.Keys[0].Kid != "test-key-1" {
		t.Fatalf("unexpected jwks1 %+v", jwks1.Keys)
	}

	// rotate key
	fsRotated := defaultFS()
	fsRotated["jwks.json"] = &fstest.MapFile{
		Data: []byte(`{"keys":[{"kty":"OKP","crv":"Ed25519","kid":"test-key-2","x":"BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"}]}`),
	}
	server2 := newIntegrationServer(t, fsRotated)
	t.Cleanup(server2.Close)

	resp2, err := http.Get(server2.URL + "/jwks.json")
	if err != nil {
		t.Fatalf("fetch jwks rotated: %v", err)
	}
	defer resp2.Body.Close()
	var jwks2 struct {
		Keys []struct {
			Kid string `json:"kid"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&jwks2); err != nil {
		t.Fatalf("decode jwks2: %v", err)
	}
	if len(jwks2.Keys) == 0 || jwks2.Keys[0].Kid != "test-key-2" {
		t.Fatalf("expected rotated kid, got %+v", jwks2.Keys)
	}
}

// helpers

type verifyPayload struct {
	RMT  string
	IMT  string
	CORT string
	PSRT string
}

func doVerifyRequest(t *testing.T, server *httptest.Server, payload verifyPayload) *http.Response {
	t.Helper()
	reqBody := map[string]any{
		"tokens": map[string]string{
			"rmt":  payload.RMT,
			"imt":  payload.IMT,
			"cort": payload.CORT,
			"psrt": payload.PSRT,
		},
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	resp, err := http.Post(server.URL+"/verify", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("post verify: %v", err)
	}
	return resp
}

func decodeBody(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func newIntegrationServer(t *testing.T, fsys fstest.MapFS) *httptest.Server {
	t.Helper()
	apiServer, err := api.NewServer(api.Config{StaticFS: fsys})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	staticVerifier, err := verifylib.NewStaticVerifier(fsys, ".", nil)
	if err != nil {
		t.Fatalf("NewStaticVerifier: %v", err)
	}
	verifyService := verify.NewService(1, staticVerifier)

	mux := http.NewServeMux()
	mux.Handle("/", apiServer)
	mux.HandleFunc("/verify", verifyService.HandleVerify)
	mux.HandleFunc("/revocations", verifyService.HandleRevocationsGet)
	mux.HandleFunc("/revocations/bump", verifyService.HandleRevocationsBump)

	return httptest.NewServer(mux)
}

func defaultFS() fstest.MapFS {
	return fstest.MapFS{
		"rrmt-eu-psd3-2025.json": {Data: []byte(`{"type":"RRMT","nbf":"2000-01-01T00:00:00Z","exp":"2100-01-01T00:00:00Z","revoked":false}`)},
		"imt-eu-sg-2025.json":    {Data: []byte(`{"type":"IMT","nbf":"2000-01-01T00:00:00Z","exp":"2100-01-01T00:00:00Z","revoked":false}`)},
		"cort-vodafone-visa-2025.json": {
			Data: []byte(`{"type":"CORT","nbf":"2000-01-01T00:00:00Z","exp":"2100-01-01T00:00:00Z","revoked":false}`),
		},
		"psrt-visa-acq-123.json": {
			Data: []byte(`{"type":"PSRT","nbf":"2000-01-01T00:00:00Z","exp":"2100-01-01T00:00:00Z","revoked":false}`),
		},
		"jwks.json": {
			Data: []byte(`{"keys":[{"kty":"OKP","crv":"Ed25519","kid":"test-key-1","x":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}]}`),
		},
	}
}
