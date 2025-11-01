package verify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// StaticVerifier loads predefined token fixtures and exposes verification helpers
// matching the SAPP verifier interface.
type StaticVerifier struct {
	tokens map[string]json.RawMessage
	meta   map[string]TokenInfo
}

// TokenInfo captures metadata surfaced alongside token payloads.
type TokenInfo struct {
	URI       string `json:"uri"`
	Type      string `json:"type"`
	Version   string `json:"version"`
	IssuedAt  string `json:"issued_at"`
	NotBefore string `json:"nbf"`
	ExpiresAt string `json:"exp"`
	Revoked   bool   `json:"revoked"`
	Hash      string `json:"hash"`
}

// FileMap links canonical token URIs to fixture filenames.
type FileMap map[string]string

var (
	// DefaultFileMap enumerates the sandbox fixtures bundled with RTGF docs.
DefaultFileMap = FileMap{
	"urn:lane2:token:RMT:EU:PSD3:3.2":         "rrmt-eu-psd3-2025.json",
	"urn:lane2:token:IMT:EU:SG:2025":          "imt-eu-sg-2025.json",
	"urn:lane2:token:CORT:VODAFONE.VISA:2025": "cort-vodafone-visa-2025.json",
	"urn:lane2:token:PSRT:VISA:ACQ-123":       "psrt-visa-acq-123.json",
}
)

// NewStaticVerifier reads fixtures from the provided filesystem rooted at baseDir.
func NewStaticVerifier(fsys fs.FS, baseDir string, files FileMap) (*StaticVerifier, error) {
	if files == nil {
		files = DefaultFileMap
	}
	if len(files) == 0 {
		return nil, errors.New("no token fixtures supplied")
	}
	resolved := make(map[string]json.RawMessage, len(files))
	meta := make(map[string]TokenInfo, len(files))
	for uri, name := range files {
		path := filepath.Join(baseDir, name)
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil, fmt.Errorf("read fixture %q (%s): %w", uri, path, err)
		}
		if !json.Valid(data) {
			return nil, fmt.Errorf("fixture %q (%s) is not valid JSON", uri, path)
		}
		resolved[uri] = append([]byte(nil), data...)
		var info TokenInfo
		if err := json.Unmarshal(data, &info); err == nil {
			info.URI = uri
			if info.Type == "" {
				info.Type = detectType(uri)
			}
			meta[uri] = info
		}
	}
	return &StaticVerifier{tokens: resolved, meta: meta}, nil
}

// VerifyRRMT ensures an RRMT token exists and carries the expected type discriminator.
func (v *StaticVerifier) VerifyRRMT(ctx context.Context, uri string) error {
	return v.verifyType(ctx, uri, "RRMT")
}

// VerifyCORT ensures a CORT token exists and carries the expected type discriminator.
func (v *StaticVerifier) VerifyCORT(ctx context.Context, uri string) error {
	return v.verifyType(ctx, uri, "CORT")
}

// VerifyPSRT ensures a PSRT token exists and carries the expected type discriminator.
func (v *StaticVerifier) VerifyPSRT(ctx context.Context, uri string) error {
	return v.verifyType(ctx, uri, "PSRT")
}

// Token returns the raw JSON payload for the given token URI.
func (v *StaticVerifier) Token(uri string) (json.RawMessage, bool) {
	data, ok := v.tokens[uri]
	if !ok {
		return nil, false
	}
	return append([]byte(nil), data...), true
}

// Metadata returns the structured TokenInfo when available.
func (v *StaticVerifier) Metadata(uri string) (TokenInfo, bool) {
	info, ok := v.meta[uri]
	return info, ok
}

func detectType(uri string) string {
	switch {
	case strings.Contains(uri, ":RRMT:"):
		return "RRMT"
	case strings.Contains(uri, ":CORT:"):
		return "CORT"
	case strings.Contains(uri, ":PSRT:"):
		return "PSRT"
	case strings.Contains(uri, ":IMT:"):
		return "IMT"
	case strings.Contains(uri, ":RMT:"):
		return "RMT"
	default:
		return ""
	}
}

func (v *StaticVerifier) verifyType(_ context.Context, uri, expectedType string) error {
	if v == nil {
		return errors.New("verifier is nil")
	}
	payload, ok := v.tokens[uri]
	if !ok {
		return fmt.Errorf("token %s not found", uri)
	}
	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return fmt.Errorf("decode %s payload: %w", uri, err)
	}
	if !strings.EqualFold(envelope.Type, expectedType) {
		return fmt.Errorf("token %s has unexpected type %q (want %s)", uri, envelope.Type, expectedType)
	}
	return nil
}
