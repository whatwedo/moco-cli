package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAPIErrorMessage(t *testing.T) {
	withBody := &APIError{StatusCode: 404, Status: "404 Not Found", Body: []byte(`{"message":"nope"}`)}
	if msg := withBody.Error(); !strings.Contains(msg, "404 Not Found") || !strings.Contains(msg, "nope") {
		t.Errorf("Error() = %q, want status and body", msg)
	}
	noBody := &APIError{StatusCode: 500, Status: "500 Internal Server Error"}
	if msg := noBody.Error(); !strings.Contains(msg, "500 Internal Server Error") {
		t.Errorf("Error() = %q, want status", msg)
	}
}

func TestDoSendsRequest(t *testing.T) {
	var gotMethod, gotPath, gotQuery, gotAuth, gotUA, gotBody, gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		gotAccept = r.Header.Get("Accept")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "secret", "moco-cli/test")
	resp, err := c.Do(context.Background(), Request{
		Method: http.MethodPost,
		Path:   "/projects",
		Query:  url.Values{"page": []string{"2"}},
		Body:   []byte(`{"name":"x"}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != `{"ok":true}` {
		t.Errorf("body = %q", resp)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q", gotMethod)
	}
	if gotPath != "/api/v1/projects" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "page=2" {
		t.Errorf("query = %q", gotQuery)
	}
	if gotAuth != "Token token=secret" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotUA != "moco-cli/test" {
		t.Errorf("User-Agent = %q", gotUA)
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept = %q", gotAccept)
	}
	if gotBody != `{"name":"x"}` {
		t.Errorf("body = %q", gotBody)
	}
}

func TestDoAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"invalid"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "t", "ua")
	_, err := c.Do(context.Background(), Request{Method: http.MethodGet, Path: "/x"})
	if err == nil {
		t.Fatal("expected an error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
	if string(apiErr.Body) != `{"error":"invalid"}` {
		t.Errorf("body = %q", apiErr.Body)
	}
}

func TestDoNoBodyNoContentType(t *testing.T) {
	var hadContentType bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hadContentType = r.Header["Content-Type"]
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New(srv.URL, "t", "ua")
	if _, err := c.Do(context.Background(), Request{Method: http.MethodDelete, Path: "/x/1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hadContentType {
		t.Error("Content-Type should not be set without a body")
	}
}
