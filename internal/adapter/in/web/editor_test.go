package in

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"app-mobile-downloader/internal/shared/server"

	"github.com/go-fuego/fuego"
)

func TestEditorUpstreamURL(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		t.Setenv("EDITOR_UPSTREAM_URL", "")
		if got := editorUpstreamURL(); got != "http://127.0.0.1:9090" {
			t.Fatalf("editorUpstreamURL() = %q", got)
		}
	})

	t.Run("custom", func(t *testing.T) {
		t.Setenv("EDITOR_UPSTREAM_URL", "  http://localhost:9999  ")
		if got := editorUpstreamURL(); got != "http://localhost:9999" {
			t.Fatalf("editorUpstreamURL() = %q", got)
		}
	})
}

func TestRewriteEditorRequest(t *testing.T) {
	t.Run("trims editor prefix", func(t *testing.T) {
		target, _ := url.Parse("http://localhost:9090")
		originalDirector := func(r *http.Request) {
			r.URL.Scheme = target.Scheme
			r.URL.Host = target.Host
		}
		req := httptest.NewRequest(http.MethodGet, "http://example.com/editor/assets/app.js", nil)

		rewriteEditorRequest(target, originalDirector, req)

		if req.URL.Path != "/assets/app.js" {
			t.Fatalf("path = %q, want /assets/app.js", req.URL.Path)
		}
		if req.Host != "localhost:9090" {
			t.Fatalf("host = %q, want localhost:9090", req.Host)
		}
		if req.Header.Get("X-Forwarded-Prefix") != "/editor" {
			t.Fatalf("X-Forwarded-Prefix = %q", req.Header.Get("X-Forwarded-Prefix"))
		}
	})

	t.Run("preserves existing prefix header", func(t *testing.T) {
		target, _ := url.Parse("http://localhost:9090")
		req := httptest.NewRequest(http.MethodGet, "http://example.com/editor/", nil)
		req.Header.Set("X-Forwarded-Prefix", "/custom")

		rewriteEditorRequest(target, func(r *http.Request) {}, req)

		if req.Header.Get("X-Forwarded-Prefix") != "/custom" {
			t.Fatalf("X-Forwarded-Prefix = %q", req.Header.Get("X-Forwarded-Prefix"))
		}
	})
}

func TestHandleEditorProxyError(t *testing.T) {
	rr := httptest.NewRecorder()
	handleEditorProxyError(rr, errors.New("upstream down"))
	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "editor upstream unavailable") {
		t.Fatalf("body = %q", rr.Body.String())
	}
}

func TestEditorHandler(t *testing.T) {
	// Verifica que editorHandler no cause panic con upstream inválido
	t.Run("invalid upstream", func(t *testing.T) {
		t.Setenv("EDITOR_UPSTREAM_URL", "://bad")
		defer t.Setenv("EDITOR_UPSTREAM_URL", "")
		fs := fuego.NewServer()
		s := &server.Server{Server: fs}
		editorHandler(s)
	})

	// Verifica que editorHandler registra el proxy sin panic
	t.Run("registers proxy", func(t *testing.T) {
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer upstream.Close()

		t.Setenv("EDITOR_UPSTREAM_URL", upstream.URL)
		defer t.Setenv("EDITOR_UPSTREAM_URL", "")

		fs := fuego.NewServer()
		s := &server.Server{Server: fs}
		editorHandler(s)
	})
}
