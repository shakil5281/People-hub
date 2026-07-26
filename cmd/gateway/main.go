package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

func newProxy(target string) *httputil.ReverseProxy {
	u, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(u)
}

func main() {
	apiTarget := getEnv("API_TARGET", "http://localhost:5050")
	webTarget := getEnv("WEB_TARGET", "http://localhost:3050")
	iisTarget := getEnv("IIS_TARGET", "http://localhost:8082")
	port := getEnv("GATEWAY_PORT", "8081")

	apiProxy := newProxy(apiTarget)
	webProxy := newProxy(webTarget)
	iisProxy := newProxy(iisTarget)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path

		switch {
		case strings.HasPrefix(p, "/people-hub/api/"):
			r.URL.Path = strings.TrimPrefix(p, "/people-hub")
			apiProxy.ServeHTTP(w, r)

		case strings.HasPrefix(p, "/people-hub/swagger/"):
			r.URL.Path = strings.TrimPrefix(p, "/people-hub")
			apiProxy.ServeHTTP(w, r)

		case strings.HasPrefix(p, "/people-hub/uploads/"):
			r.URL.Path = strings.TrimPrefix(p, "/people-hub")
			apiProxy.ServeHTTP(w, r)

		case p == "/people-hub/health":
			r.URL.Path = "/health"
			apiProxy.ServeHTTP(w, r)

		case strings.HasPrefix(p, "/people-hub"):
			webProxy.ServeHTTP(w, r)

		case strings.HasPrefix(p, "/contact"):
			iisProxy.ServeHTTP(w, r)

		case strings.HasPrefix(p, "/api/"),
			strings.HasPrefix(p, "/swagger/"),
			strings.HasPrefix(p, "/uploads/"):
			apiProxy.ServeHTTP(w, r)

		case p == "/health":
			apiProxy.ServeHTTP(w, r)

		case p == "/":
			http.Redirect(w, r, "/people-hub", http.StatusFound)

		default:
			webProxy.ServeHTTP(w, r)
		}
	})

	log.Printf("Gateway listening on :%s", port)
	log.Printf("  /people-hub/*   → %s (Next.js)", webTarget)
	log.Printf("  /people-hub/api/* → %s", apiTarget)
	log.Printf("  /contact/* → %s (IIS)", iisTarget)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Gateway failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
