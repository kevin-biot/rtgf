package verify

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stubVerifier struct {
	verifyErr error
	tokens    map[string]string
}

func (s *stubVerifier) VerifyRRMT(ctx context.Context, uri string) error { return s.verifyErr }
func (s *stubVerifier) VerifyCORT(ctx context.Context, uri string) error { return s.verifyErr }
func (s *stubVerifier) VerifyPSRT(ctx context.Context, uri string) error { return s.verifyErr }
func (s *stubVerifier) Token(uri string) (json.RawMessage, bool) {
	if s.tokens == nil {
		return nil, false
	}
	val, ok := s.tokens[uri]
	if !ok {
		return nil, false
	}
	return json.RawMessage(val), true
}

func happyTokens() map[string]string {
	return map[string]string{
		"urn:lane2:token:RMT:EU:PSD3:3.2":         `{"nbf":"2000-01-01T00:00:00Z","exp":"2100-01-01T00:00:00Z","revoked":false}`,
		"urn:lane2:token:IMT:EU:SG:2025":          `{"nbf":"2000-01-01T00:00:00Z","exp":"2100-01-01T00:00:00Z","revoked":false}`,
		"urn:lane2:token:CORT:VODAFONE.VISA:2025": `{"nbf":"2000-01-01T00:00:00Z","exp":"2100-01-01T00:00:00Z","revoked":false}`,
		"urn:lane2:token:PSRT:VISA:ACQ-123":       `{"nbf":"2000-01-01T00:00:00Z","exp":"2100-01-01T00:00:00Z","revoked":false}`,
	}
}

func TestVerifyHandler(t *testing.T) {
	svc := NewService(42, &stubVerifier{tokens: happyTokens()})
	payload := VerifyRequest{}
	payload.Tokens.RMT = "urn:lane2:token:RMT:EU:PSD3:3.2"
	payload.Tokens.IMT = "urn:lane2:token:IMT:EU:SG:2025"
	payload.Tokens.CORT = "urn:lane2:token:CORT:VODAFONE.VISA:2025"
	payload.Tokens.PSRT = "urn:lane2:token:PSRT:VISA:ACQ-123"

	body, _ := json.Marshal(payload)
	rec := httptest.NewRecorder()
	svc.HandleVerify(rec, httptest.NewRequest(http.MethodPost, "/verify", bytes.NewReader(body)))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
	var resp VerifyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal resp: %v", err)
	}
	if !resp.Valid || resp.RevEpoch != 42 {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestVerifyMissingToken(t *testing.T) {
	svc := NewService(1, &stubVerifier{tokens: happyTokens()})
	payload := VerifyRequest{}
	payload.Tokens.IMT = "urn:lane2:token:IMT:EU:SG:2025"
	payload.Tokens.CORT = "urn:lane2:token:CORT:VODAFONE.VISA:2025"
	payload.Tokens.PSRT = "urn:lane2:token:PSRT:VISA:ACQ-123"
	body, _ := json.Marshal(payload)
	rec := httptest.NewRecorder()
	svc.HandleVerify(rec, httptest.NewRequest(http.MethodPost, "/verify", bytes.NewReader(body)))
	var resp VerifyResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Valid || resp.Reason != "missing_rmt" {
		t.Fatalf("expected missing_rmt got %+v", resp)
	}
}

func TestVerifyExpiredToken(t *testing.T) {
	tokens := happyTokens()
	tokens["urn:lane2:token:RMT:EU:PSD3:3.2"] = `{"nbf":"2000-01-01T00:00:00Z","exp":"2001-01-01T00:00:00Z","revoked":false}`
	svc := NewService(1, &stubVerifier{tokens: tokens})
	payload := VerifyRequest{}
	payload.Tokens.RMT = "urn:lane2:token:RMT:EU:PSD3:3.2"
	payload.Tokens.IMT = "urn:lane2:token:IMT:EU:SG:2025"
	payload.Tokens.CORT = "urn:lane2:token:CORT:VODAFONE.VISA:2025"
	payload.Tokens.PSRT = "urn:lane2:token:PSRT:VISA:ACQ-123"
	body, _ := json.Marshal(payload)
	rec := httptest.NewRecorder()
	svc.HandleVerify(rec, httptest.NewRequest(http.MethodPost, "/verify", bytes.NewReader(body)))
	var resp VerifyResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Valid || !strings.Contains(resp.Reason, "token_expired") {
		t.Fatalf("expected token_expired got %+v", resp)
	}
}

func TestRevocationBump(t *testing.T) {
	svc := NewService(1, &stubVerifier{})
	rec := httptest.NewRecorder()
	svc.HandleRevocationsBump(rec, httptest.NewRequest(http.MethodPost, "/revocations/bump", nil))

	var out RevocationResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if out.RevEpoch != 2 {
		t.Fatalf("expected revEpoch 2 got %d", out.RevEpoch)
	}
}
