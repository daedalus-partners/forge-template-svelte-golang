package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

// API-only server; static assets are served by the separate web service

type infoResponse struct {
	Message string            `json:"message"`
	Email   string            `json:"email"`
	Env     map[string]string `json:"env"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		email := r.Header.Get("Cf-Access-Authenticated-User-Email")
		if email == "" {
			email = "(unknown; ensure Cloudflare Access is configured)"
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
		_ = json.NewEncoder(w).Encode(infoResponse{
			Message: "This is a Forge test application.",
			Email:   email,
			Env:     env,
		})
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
