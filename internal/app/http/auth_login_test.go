package app

import (
	"encoding/base64"
	"net/url"
	"strings"
	"testing"

	"app-mobile-downloader/internal/shared/configuration"
)

func TestBuildLoginURL(t *testing.T) {
	conf := configuration.Conf{
		OIDCLoginURL:            "https://issuer.example/login",
		OIDCGoogleLoginURL:      "https://issuer.example/google-login",
		OIDCAuthorizationEndpoint: "https://issuer.example/authorize",
		OIDCClientID:            "client-1",
		OIDCRedirectURI:         "https://app.example/auth/callback",
		OIDCScopes:              "openid email",
	}

	t.Run("default login url", func(t *testing.T) {
		raw, err := buildLoginURL(conf, "state-1", false)
		if err != nil {
			t.Fatalf("buildLoginURL() error = %v", err)
		}
		u, _ := url.Parse(raw)
		if u.String() == "" {
			t.Fatal("expected non-empty url")
		}
		q := u.Query()
		if q.Get("client_id") != "client-1" || q.Get("redirect_uri") != conf.OIDCRedirectURI || q.Get("state") != "state-1" {
			t.Fatalf("unexpected query: %v", q)
		}
		if q.Get("scope") != "openid email" {
			t.Fatalf("scope = %q", q.Get("scope"))
		}
	})

	t.Run("prefer google login url", func(t *testing.T) {
		raw, err := buildLoginURL(conf, "state-2", true)
		if err != nil {
			t.Fatalf("buildLoginURL() error = %v", err)
		}
		if !strings.HasPrefix(raw, conf.OIDCGoogleLoginURL) {
			t.Fatalf("expected google login base, got %q", raw)
		}
	})

	t.Run("fallback to authorization endpoint", func(t *testing.T) {
		conf.OIDCLoginURL = ""
		conf.OIDCGoogleLoginURL = ""
		raw, err := buildLoginURL(conf, "state-3", false)
		if err != nil {
			t.Fatalf("buildLoginURL() error = %v", err)
		}
		if !strings.HasPrefix(raw, conf.OIDCAuthorizationEndpoint) {
			t.Fatalf("expected authorization endpoint base, got %q", raw)
		}
	})
}

func TestBuildDirectGoogleLoginURL(t *testing.T) {
	conf := configuration.Conf{
		OIDCUpstreamGoogleClientID: "google-client",
		OIDCIssuer:                 "https://casdoor.example/",
		OIDCClientID:               "einar-app-mobile-downloader-dev",
		PROJECT_NAME:               "mobile-downloader",
		OIDCRedirectURI:            "https://app.example/auth/callback",
		OIDCScopes:                 "openid profile email",
	}

	raw, err := buildDirectGoogleLoginURL(conf, "opaque-state")
	if err != nil {
		t.Fatalf("buildDirectGoogleLoginURL() error = %v", err)
	}

	u, _ := url.Parse(raw)
	q := u.Query()
	if q.Get("client_id") != "google-client" {
		t.Fatalf("client_id = %q", q.Get("client_id"))
	}
	if q.Get("redirect_uri") != "https://casdoor.example/callback" {
		t.Fatalf("redirect_uri = %q", q.Get("redirect_uri"))
	}

	decoded, err := base64.StdEncoding.DecodeString(q.Get("state"))
	if err != nil {
		t.Fatalf("state is not valid base64: %v", err)
	}
	packed := string(decoded)
	for _, fragment := range []string{"application=einar-app-mobile-downloader", "provider=provider_google_einar", "method=signup", "state=opaque-state"} {
		if !strings.Contains(packed, fragment) {
			t.Fatalf("packed state %q does not contain %q", packed, fragment)
		}
	}
}

func TestBuildDirectGoogleLoginURLErrors(t *testing.T) {
	tests := []struct {
		name string
		conf configuration.Conf
	}{
		{name: "missing google client id", conf: configuration.Conf{OIDCIssuer: "https://issuer", OIDCClientID: "x-project-y", PROJECT_NAME: "project"}},
		{name: "missing issuer", conf: configuration.Conf{OIDCUpstreamGoogleClientID: "google", OIDCClientID: "x-project-y", PROJECT_NAME: "project"}},
		{name: "cannot derive app name", conf: configuration.Conf{OIDCUpstreamGoogleClientID: "google", OIDCIssuer: "https://issuer", OIDCClientID: "client", PROJECT_NAME: "project"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := buildDirectGoogleLoginURL(tt.conf, "state"); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestDeriveCasdoorAppName(t *testing.T) {
	got, err := deriveCasdoorAppName("einar-mobile-downloader-dev", "mobile-downloader")
	if err != nil {
		t.Fatalf("deriveCasdoorAppName() error = %v", err)
	}
	if got != "einar-mobile-downloader" {
		t.Fatalf("got %q", got)
	}

	if _, err := deriveCasdoorAppName("", "project"); err == nil {
		t.Fatal("expected error for empty values")
	}
	if _, err := deriveCasdoorAppName("client-without-slug", "project"); err == nil {
		t.Fatal("expected error when slug cannot be derived")
	}
}

func TestRandomState(t *testing.T) {
	state, err := randomState()
	if err != nil {
		t.Fatalf("randomState() error = %v", err)
	}
	if len(state) == 0 {
		t.Fatal("expected non-empty state")
	}
	if strings.Contains(state, "=") {
		t.Fatal("expected raw url encoding without padding")
	}
}

func TestIsHTTPS(t *testing.T) {
	if !isHTTPS("https://example.com/callback") {
		t.Fatal("expected https url to be detected")
	}
	if isHTTPS("http://example.com/callback") || isHTTPS("://bad") {
		t.Fatal("expected non-https values to be false")
	}
}

func TestFirstNonEmptyScope(t *testing.T) {
	if got := firstNonEmptyScope(" ", "openid email"); got != "openid email" {
		t.Fatalf("firstNonEmptyScope() = %q", got)
	}
}
