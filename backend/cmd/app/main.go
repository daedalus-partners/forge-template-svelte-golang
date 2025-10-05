package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

// API-only server; static assets are served by the separate web service

type whoamiResponse struct {
	Email string `json:"email"`
}

type envResponse struct {
	Env map[string]string `json:"env"`
}

func main() {
	mux := http.NewServeMux()

	// whoami: returns Cloudflare Access authenticated email (empty if not present)
	mux.HandleFunc("/api/whoami", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		email := r.Header.Get("Cf-Access-Authenticated-User-Email")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(whoamiResponse{Email: email})
	})

	// env: returns all environment variables visible to the process
	mux.HandleFunc("/api/env", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		env := map[string]string{}
		for _, e := range os.Environ() {
			if i := strings.IndexByte(e, '='); i > 0 {
				k := e[:i]
				v := e[i+1:]
				env[k] = v
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(envResponse{Env: env})
	})

	// healthz for basic container health
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// No root handler; unknown routes return 404. Only /api/* is served.

	addr := ":" + envOr("PORT", "8080")
	log.Printf("template app listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
