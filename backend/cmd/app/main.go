package main

import (
    "embed"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "path"
    "path/filepath"
    "strings"
)

//go:embed static/*
var staticFS embed.FS

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

    // Static file server with SPA fallback
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Try to serve static asset under static/
        p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
        if p == "" || p == "/" {
            p = "index.html"
        }
        // security: prevent path traversal
        p = filepath.ToSlash(p)
        if strings.Contains(p, "..") {
            http.Error(w, "bad path", 400)
            return
        }
        data, err := staticFS.ReadFile("static/" + p)
        if err != nil {
            // fallback to index.html for SPA routes
            data, err = staticFS.ReadFile("static/index.html")
            if err != nil {
                http.NotFound(w, r)
                return
            }
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            _, _ = w.Write(data)
            return
        }
        // very minimal content type detection
        if strings.HasSuffix(p, ".js") {
            w.Header().Set("Content-Type", "application/javascript")
        } else if strings.HasSuffix(p, ".css") {
            w.Header().Set("Content-Type", "text/css")
        } else if strings.HasSuffix(p, ".html") {
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
        }
        _, _ = w.Write(data)
    })

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


