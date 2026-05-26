package handlers

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/abigailtech/humanloop/backend/middleware"
	"github.com/abigailtech/humanloop/backend/pipeline"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-secret")
	Pipeline = pipeline.New(1, os.TempDir(), nil)
	code := m.Run()
	Pipeline.Shutdown()
	os.Exit(code)
}

func authContext(r *http.Request) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "test-user")
	ctx = context.WithValue(ctx, middleware.UserEmailKey, "test@example.com")
	ctx = context.WithValue(ctx, middleware.UserNameKey, "Test User")
	return r.WithContext(ctx)
}

func buildVideoRequest(t *testing.T, fieldName, filename string, body []byte, extraFields map[string]string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatal(err)
	}
	fw.Write(body)
	for k, v := range extraFields {
		w.WriteField(k, v)
	}
	w.Close()
	req := httptest.NewRequest("POST", "/v1/submit/challenge-1", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return authContext(req)
}

func mp4Header() []byte {
	h := make([]byte, 512)
	h[0], h[1], h[2], h[3] = 0x00, 0x00, 0x00, 0x14
	copy(h[4:8], "ftyp")
	copy(h[8:12], "mp42")
	return h
}

func TestSubmit_NoVideo(t *testing.T) {
	req := httptest.NewRequest("POST", "/v1/submit/c1", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	req = authContext(req)
	rr := httptest.NewRecorder()
	Submit(rr, req)
	if rr.Code == http.StatusOK {
		t.Fatal("expected non-200 for missing video")
	}
}

func TestSubmit_InvalidMIME(t *testing.T) {
	body := make([]byte, 512)
	copy(body, []byte("PK\x03\x04"))
	req := buildVideoRequest(t, "video", "evil.zip", body, nil)
	rr := httptest.NewRecorder()
	Submit(rr, req)
	if rr.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("want 415, got %d", rr.Code)
	}
}

func TestSubmit_ValidMIME_ProceedsToStorage(t *testing.T) {
	body := mp4Header()
	req := buildVideoRequest(t, "video", "test.mp4", body, map[string]string{
		"robot": "so100",
	})
	rr := httptest.NewRecorder()
	Submit(rr, req)
	if rr.Code == http.StatusUnsupportedMediaType {
		t.Fatal("valid mp4 header should not be rejected as wrong MIME")
	}
}
