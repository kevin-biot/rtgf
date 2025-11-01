package verify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type Service struct {
	revEpoch atomic.Uint64
	verifier TokenVerifier
}

type VerifyRequest struct {
	Tokens struct {
		RMT  string `json:"rmt"`
		IMT  string `json:"imt"`
		CORT string `json:"cort"`
		PSRT string `json:"psrt"`
		AMLS string `json:"amls,omitempty"`
		AMLV string `json:"amlv,omitempty"`
	} `json:"tokens"`
}

type VerifyResponse struct {
	Valid    bool   `json:"valid"`
	RevEpoch uint64 `json:"revEpoch"`
	Reason   string `json:"reason,omitempty"`
}

type RevocationResponse struct {
	RevEpoch uint64 `json:"revEpoch"`
}

func NewService(initial uint64, verifier TokenVerifier) *Service {
	s := &Service{verifier: verifier}
	s.revEpoch.Store(initial)
	return s
}

func (s *Service) HandleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, VerifyResponse{Valid: false, RevEpoch: s.revEpoch.Load(), Reason: "invalid_request"})
		return
	}
	valid, reason := s.validateTokens(r.Context(), req)
	resp := VerifyResponse{Valid: valid, RevEpoch: s.revEpoch.Load(), Reason: reason}
	respondJSON(w, resp)
}

func (s *Service) HandleRevocationsGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	respondJSON(w, RevocationResponse{RevEpoch: s.revEpoch.Load()})
}

func (s *Service) HandleRevocationsBump(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	s.revEpoch.Add(1)
	respondJSON(w, RevocationResponse{RevEpoch: s.revEpoch.Load()})
}

func (s *Service) validateTokens(ctx context.Context, req VerifyRequest) (bool, string) {
	required := map[string]string{
		"rmt":  req.Tokens.RMT,
		"imt":  req.Tokens.IMT,
		"cort": req.Tokens.CORT,
		"psrt": req.Tokens.PSRT,
	}
	for key, value := range required {
		if strings.TrimSpace(value) == "" {
			return false, fmt.Sprintf("missing_%s", key)
		}
	}
	if s.verifier != nil {
		if err := validateWindows(currentTime(), s.verifier, req); err != nil {
			return false, err.Error()
		}
		if err := s.verifier.VerifyRRMT(ctx, req.Tokens.RMT); err != nil {
			return false, "invalid_rrmt"
		}
		if err := s.verifier.VerifyCORT(ctx, req.Tokens.CORT); err != nil {
			return false, "invalid_cort"
		}
		if err := s.verifier.VerifyPSRT(ctx, req.Tokens.PSRT); err != nil {
			return false, "invalid_psrt"
		}
	}
	return true, ""
}

func respondJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

var ErrInvalidMethod = errors.New("invalid method")

type TokenVerifier interface {
	VerifyRRMT(ctx context.Context, uri string) error
	VerifyCORT(ctx context.Context, uri string) error
	VerifyPSRT(ctx context.Context, uri string) error
	Token(uri string) (json.RawMessage, bool)
}

func validateWindows(now time.Time, provider TokenVerifier, req VerifyRequest) error {
	type tokenMeta struct {
		NotBefore string `json:"nbf"`
		Expires   string `json:"exp"`
		Revoked   bool   `json:"revoked"`
	}
	for _, uri := range []string{req.Tokens.RMT, req.Tokens.IMT, req.Tokens.CORT, req.Tokens.PSRT} {
		if uri == "" {
			continue
		}
		payload, ok := provider.Token(uri)
		if !ok || len(payload) == 0 {
			return fmt.Errorf("metadata_missing:%s", uri)
		}
		var meta tokenMeta
		if err := json.Unmarshal(payload, &meta); err != nil {
			return fmt.Errorf("metadata_invalid:%s", uri)
		}
		if meta.Revoked {
			return fmt.Errorf("token_revoked:%s", uri)
		}
		if meta.NotBefore != "" {
			nbf, err := time.Parse(time.RFC3339, meta.NotBefore)
			if err != nil {
				return fmt.Errorf("invalid_nbf:%s", uri)
			}
			if now.Before(nbf) {
				return fmt.Errorf("token_not_yet_valid:%s", uri)
			}
		}
		if meta.Expires != "" {
			exp, err := time.Parse(time.RFC3339, meta.Expires)
			if err != nil {
				return fmt.Errorf("invalid_exp:%s", uri)
			}
			if now.After(exp) {
				return fmt.Errorf("token_expired:%s", uri)
			}
		}
	}
	return nil
}

func currentTime() time.Time {
	if fixed := os.Getenv("FIXED_TIME"); fixed != "" {
		if ts, err := time.Parse(time.RFC3339, fixed); err == nil {
			return ts.UTC()
		}
	}
	return time.Now().UTC()
}
