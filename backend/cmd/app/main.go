package main

import (
    "crypto/rsa"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "math/big"
    "net/http"
    "os"
    "strings"
    "time"

    jwt "github.com/golang-jwt/jwt/v5"
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
        email, err := getCloudflareEmail(r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusUnauthorized)
            return
        }
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

// getCloudflareEmail extracts the authenticated email from Cloudflare Access.
// It prefers the direct header and falls back to validating the Access JWT.
func getCloudflareEmail(r *http.Request) (string, error) {
    if e := strings.TrimSpace(r.Header.Get("Cf-Access-Authenticated-User-Email")); e != "" {
        return e, nil
    }
    tokenStr := strings.TrimSpace(r.Header.Get("Cf-Access-Jwt-Assertion"))
    if tokenStr == "" {
        return "", errors.New("cloudflare access headers not present")
    }

    // Decode payload to get iss and aud for validation and JWKS discovery
    iss, auds, err := peekIssAndAud(tokenStr)
    if err != nil {
        return "", fmt.Errorf("invalid access token: %w", err)
    }
    audExpected := strings.TrimSpace(os.Getenv("CF_ACCESS_AUD"))
    if audExpected == "" {
        return "", errors.New("CF_ACCESS_AUD not configured")
    }
    if !contains(auds, audExpected) {
        return "", errors.New("access token audience mismatch")
    }
    jwksURL := strings.TrimSuffix(iss, "/") + "/cdn-cgi/access/certs"
    keys, err := fetchRSAPublicKeys(jwksURL)
    if err != nil {
        return "", fmt.Errorf("failed to fetch JWKS: %w", err)
    }

    // Parse and validate token (signature, iss, aud)
    parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
        if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        kid, _ := t.Header["kid"].(string)
        if kid == "" {
            return nil, errors.New("missing kid in token header")
        }
        k, ok := keys[kid]
        if !ok {
            return nil, fmt.Errorf("unknown kid: %s", kid)
        }
        return k, nil
    }, jwt.WithIssuer(iss), jwt.WithAudience(audExpected), jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}))
    if err != nil {
        return "", fmt.Errorf("invalid access token: %w", err)
    }
    if !parsed.Valid {
        return "", errors.New("invalid access token")
    }
    claims, ok := parsed.Claims.(jwt.MapClaims)
    if !ok {
        return "", errors.New("invalid access token claims")
    }
    if email, _ := claims["email"].(string); strings.TrimSpace(email) != "" {
        return email, nil
    }
    return "", errors.New("email claim missing in access token")
}

func contains(list []string, want string) bool {
    for _, v := range list {
        if v == want {
            return true
        }
    }
    return false
}

// peekIssAndAud decodes the JWT payload without verification to read iss and aud.
func peekIssAndAud(token string) (iss string, auds []string, err error) {
    parts := strings.Split(token, ".")
    if len(parts) != 3 {
        return "", nil, errors.New("malformed JWT")
    }
    payload, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        return "", nil, fmt.Errorf("decode payload: %w", err)
    }
    var m map[string]any
    if err := json.Unmarshal(payload, &m); err != nil {
        return "", nil, fmt.Errorf("unmarshal payload: %w", err)
    }
    iss, _ = m["iss"].(string)
    switch v := m["aud"].(type) {
    case string:
        auds = []string{v}
    case []any:
        for _, it := range v {
            if s, ok := it.(string); ok {
                auds = append(auds, s)
            }
        }
    }
    if iss == "" || len(auds) == 0 {
        return "", nil, errors.New("missing iss/aud in token")
    }
    return iss, auds, nil
}

// fetchRSAPublicKeys retrieves the JWKS and returns RSA public keys keyed by kid.
func fetchRSAPublicKeys(jwksURL string) (map[string]*rsa.PublicKey, error) {
    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get(jwksURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("jwks http %d", resp.StatusCode)
    }
    var body struct {
        Keys []struct {
            Kty string `json:"kty"`
            Kid string `json:"kid"`
            N   string `json:"n"`
            E   string `json:"e"`
        } `json:"keys"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
        return nil, err
    }
    out := make(map[string]*rsa.PublicKey)
    for _, k := range body.Keys {
        if strings.ToUpper(k.Kty) != "RSA" || k.N == "" || k.E == "" || k.Kid == "" {
            continue
        }
        nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
        if err != nil {
            continue
        }
        eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
        if err != nil || len(eBytes) == 0 {
            continue
        }
        // Convert exponent bytes (big-endian) to int
        e := 0
        for _, b := range eBytes {
            e = (e << 8) | int(b)
        }
        if e <= 0 {
            continue
        }
        pub := &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}
        out[k.Kid] = pub
    }
    if len(out) == 0 {
        return nil, errors.New("no RSA keys in JWKS")
    }
    return out, nil
}
