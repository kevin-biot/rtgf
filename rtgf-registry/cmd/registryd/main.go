package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/kevin-biot/rtgf/rtgf-registry/internal/api"
	"github.com/kevin-biot/rtgf/rtgf-registry/internal/verify"
	verifylib "github.com/kevin-biot/rtgf/rtgf-verify-lib"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	staticDir := flag.String("static-dir", "../registry/static/tokens", "path to token fixtures")
	flag.Parse()

	fsys := os.DirFS(*staticDir)
	server, err := api.NewServer(api.Config{
		StaticFS: fsys,
	})
	if err != nil {
		log.Fatalf("init server: %v", err)
	}

	staticVerifier, err := verifylib.NewStaticVerifier(fsys, ".", nil)
	if err != nil {
		log.Fatalf("init static verifier: %v", err)
	}
	verifyService := verify.NewService(1, staticVerifier)
	mux := http.NewServeMux()
	mux.Handle("/", server)
	mux.HandleFunc("/verify", verifyService.HandleVerify)
	mux.HandleFunc("/revocations", verifyService.HandleRevocationsGet)
	mux.HandleFunc("/revocations/bump", verifyService.HandleRevocationsBump)

	log.Printf("rtgf-registryd listening on %s (static dir: %s)", *addr, *staticDir)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
