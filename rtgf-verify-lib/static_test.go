package verify

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"
)

func TestStaticVerifierHappyPath(t *testing.T) {
	fsys := fstest.MapFS{
		"rrmt-eu-psd3-2025.json":       {Data: []byte(`{"type":"RRMT","nbf":"2000-01-01T00:00:00Z","exp":"2100-01-01T00:00:00Z","revoked":false}`)},
		"imt-eu-sg-2025.json":          {Data: []byte(`{"type":"IMT","nbf":"2000-01-01T00:00:00Z","exp":"2100-01-01T00:00:00Z","revoked":false}`)},
		"cort-vodafone-visa-2025.json": {Data: []byte(`{"type":"CORT","nbf":"2000-01-01T00:00:00Z","exp":"2100-01-01T00:00:00Z","revoked":false}`)},
		"psrt-visa-acq-123.json":       {Data: []byte(`{"type":"PSRT","nbf":"2000-01-01T00:00:00Z","exp":"2100-01-01T00:00:00Z","revoked":false}`)},
	}
	verifier, err := NewStaticVerifier(fsys, ".", nil)
	if err != nil {
		t.Fatalf("NewStaticVerifier: %v", err)
	}

	ctx := context.Background()
	for uri := range DefaultFileMap {
		switch expectedType(uri) {
		case "RRMT":
			if err := verifier.VerifyRRMT(ctx, uri); err != nil {
				t.Fatalf("VerifyRRMT %s: %v", uri, err)
			}
		case "CORT":
			if err := verifier.VerifyCORT(ctx, uri); err != nil {
				t.Fatalf("VerifyCORT %s: %v", uri, err)
			}
		case "PSRT":
			if err := verifier.VerifyPSRT(ctx, uri); err != nil {
				t.Fatalf("VerifyPSRT %s: %v", uri, err)
			}
		case "RMT", "IMT":
			// no signature verification required, but metadata must exist
		default:
			t.Fatalf("unexpected token type for %s", uri)
		}
		if payload, ok := verifier.Token(uri); !ok || len(payload) == 0 {
			t.Fatalf("expected payload for %s", uri)
		}
		if info, ok := verifier.Metadata(uri); !ok || info.URI != uri {
			t.Fatalf("expected metadata for %s", uri)
		}
	}
}

func TestStaticVerifierUnknownToken(t *testing.T) {
	fsys := fstest.MapFS{
		"rrmt-eu-psd3-2025.json": {Data: []byte(`{"type":"RRMT"}`)},
	}
	fileMap := FileMap{
		"urn:lane2:token:RMT:EU:PSD3:3.2": "rrmt-eu-psd3-2025.json",
	}
	verifier, err := NewStaticVerifier(fsys, ".", fileMap)
	if err != nil {
		t.Fatalf("NewStaticVerifier: %v", err)
	}
	err = verifier.VerifyRRMT(context.Background(), "urn:lane2:token:RMT:UNKNOWN")
	if err == nil {
		t.Fatalf("expected error for unknown token")
	}
}

func expectedType(uri string) string {
	switch {
	case strings.Contains(uri, ":RRMT:") || strings.Contains(uri, ":RMT:"):
		return "RRMT"
	case strings.Contains(uri, ":CORT:"):
		return "CORT"
	case strings.Contains(uri, ":PSRT:"):
		return "PSRT"
	case strings.Contains(uri, ":IMT:"):
		return "IMT"
	default:
		return ""
	}
}

func TestNewStaticVerifierInvalidJSON(t *testing.T) {
	fsys := fstest.MapFS{
		"broken.json": {Data: []byte(`{"type":`)},
	}
	fileMap := FileMap{
		"urn:lane2:token:RMT:BROKEN": "broken.json",
	}
	_, err := NewStaticVerifier(fsys, ".", fileMap)
	if err == nil || !strings.Contains(err.Error(), "not valid JSON") {
		t.Fatalf("expected invalid JSON error, got %v", err)
	}
}

func TestNewStaticVerifierTypeFallback(t *testing.T) {
	fsys := fstest.MapFS{
		"imt.json": {Data: []byte(`{"nbf":"2000-01-01T00:00:00Z","exp":"2100-01-01T00:00:00Z"}`)},
	}
	fileMap := FileMap{
		"urn:lane2:token:IMT:EU:SG:2025": "imt.json",
	}
	verifier, err := NewStaticVerifier(fsys, ".", fileMap)
	if err != nil {
		t.Fatalf("NewStaticVerifier: %v", err)
	}
	info, ok := verifier.Metadata("urn:lane2:token:IMT:EU:SG:2025")
	if !ok {
		t.Fatalf("expected metadata for IMT")
	}
	if info.Type != "IMT" {
		t.Fatalf("expected fallback type IMT, got %s", info.Type)
	}
}

func TestVerifyRRMTTypeMismatch(t *testing.T) {
	fsys := fstest.MapFS{
		"rrmt.json": {Data: []byte(`{"type":"CORT"}`)},
	}
	fileMap := FileMap{
		"urn:lane2:token:RMT:EU:PSD3:3.2": "rrmt.json",
	}
	verifier, err := NewStaticVerifier(fsys, ".", fileMap)
	if err != nil {
		t.Fatalf("NewStaticVerifier: %v", err)
	}
	err = verifier.VerifyRRMT(context.Background(), "urn:lane2:token:RMT:EU:PSD3:3.2")
	if err == nil || !strings.Contains(err.Error(), "unexpected type") {
		t.Fatalf("expected unexpected type error, got %v", err)
	}
}

func TestDetectTypeTable(t *testing.T) {
	cases := map[string]string{
		"urn:lane2:token:RRMT:EU:PSD3:3.2": "RRMT",
		"urn:lane2:token:CORT:FOO":         "CORT",
		"urn:lane2:token:PSRT:BAR":         "PSRT",
		"urn:lane2:token:IMT:EU:SG:2025":   "IMT",
		"urn:lane2:token:RMT:EU:PSD3:3.2":  "RMT",
		"urn:lane2:token:UNKNOWN":          "",
	}
	for uri, want := range cases {
		if got := detectType(uri); got != want {
			t.Fatalf("detectType(%s) = %q, want %q", uri, got, want)
		}
	}
}
