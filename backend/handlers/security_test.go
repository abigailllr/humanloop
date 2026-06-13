package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSafeLocalHMDFPath(t *testing.T) {
	os.Setenv("EXTRACTED_DIR", "/tmp/hl-extracted")
	defer os.Unsetenv("EXTRACTED_DIR")
	base, _ := filepath.Abs("/tmp/hl-extracted")

	good := filepath.Join(base, "abc.hmdf.json.gz")
	if got, ok := safeLocalHMDFPath(good); !ok || got != good {
		t.Errorf("expected %q allowed, got ok=%v path=%q", good, ok, got)
	}

	bad := []string{
		"",
		"s3://bucket/key",
		"/tmp/hl-extracted/../secret",
		"/tmp/hl-extracted/../../etc/passwd",
		"/etc/passwd",
		filepath.Join(base, "..", "x"),
	}
	for _, p := range bad {
		if _, ok := safeLocalHMDFPath(p); ok {
			t.Errorf("expected %q rejected, but it was allowed", p)
		}
	}
}

func TestValidWebhookURL(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"https://8.8.8.8", true},
		{"http://8.8.8.8", false},
		{"http://example.com", false},
		{"https://127.0.0.1", false},
		{"https://localhost", false},
		{"https://10.0.0.1", false},
		{"https://192.168.1.10", false},
		{"https://172.16.0.5", false},
		{"https://169.254.169.254", false},
		{"https://0.0.0.0", false},
		{"ftp://8.8.8.8", false},
		{"", false},
		{"not-a-url", false},
		{"https://", false},
	}
	for _, c := range cases {
		if got := validWebhookURL(c.url); got != c.want {
			t.Errorf("validWebhookURL(%q) = %v, want %v", c.url, got, c.want)
		}
	}
}

func TestProtectedHandlersUnauthorized(t *testing.T) {
	handlers := map[string]http.HandlerFunc{
		"GetProfile":         GetProfile,
		"GetUserStats":       GetUserStats,
		"GetCreditHistory":   GetCreditHistory,
		"GetUserSubmissions": GetUserSubmissions,
		"GetReferral":        GetReferral,
		"DeleteAccount":      DeleteAccount,
	}
	for name, h := range handlers {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked without auth context: %v", name, r)
				}
			}()
			h(rec, req)
		}()
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s without auth = %d, want %d", name, rec.Code, http.StatusUnauthorized)
		}
	}
}
