package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/whatwedo/moco-cli/internal/client"
)

func TestWriteOutput(t *testing.T) {
	tests := []struct {
		name   string
		output string
		resp   string
		want   string
	}{
		{name: "json pretty", output: "json", resp: `{"a":1}`, want: "{\n  \"a\": 1\n}\n"},
		{name: "raw unchanged", output: "raw", resp: `{"a":1}`, want: "{\"a\":1}\n"},
		{name: "empty body", output: "json", resp: "", want: ""},
		{name: "non-json falls back", output: "json", resp: "not json", want: "not json\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			app := &App{Output: tt.output, Out: &out}
			if err := app.writeOutput([]byte(tt.resp)); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.String() != tt.want {
				t.Errorf("output = %q, want %q", out.String(), tt.want)
			}
		})
	}
}

func TestWriteOutputUnknownFormat(t *testing.T) {
	app := &App{Output: "xml", Out: &bytes.Buffer{}}
	err := app.writeOutput([]byte(`{}`))
	if _, ok := err.(*UsageError); !ok {
		t.Fatalf("expected *UsageError, got %T", err)
	}
}

func TestParseDataAndEncodeBody(t *testing.T) {
	body, err := ParseData(`{"name":"x"}`)
	if err != nil {
		t.Fatalf("ParseData: %v", err)
	}
	body["active"] = true
	raw, err := EncodeBody(body)
	if err != nil {
		t.Fatalf("EncodeBody: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	want := map[string]any{"name": "x", "active": true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("body = %v, want %v", got, want)
	}
}

func TestParseDataInvalid(t *testing.T) {
	if _, err := ParseData("not json"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestEncodeBodyEmpty(t *testing.T) {
	raw, err := EncodeBody(map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw != nil {
		t.Errorf("expected nil, got %q", raw)
	}
}

func TestExecuteEndToEnd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer srv.Close()

	var out bytes.Buffer
	app := &App{Version: "test", Output: "json", Out: &out, Client: client.New(srv.URL+"/api/v1", "tok", "test")}
	err := app.Execute(context.Background(), client.Request{Method: http.MethodGet, Path: "/projects"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out.String() != "{\n  \"id\": 1\n}\n" {
		t.Errorf("output = %q", out.String())
	}
}

func TestExecuteMissingConfig(t *testing.T) {
	app := &App{Version: "test", Out: &bytes.Buffer{}}
	err := app.Execute(context.Background(), client.Request{Method: http.MethodGet, Path: "/x"})
	if _, ok := err.(*UsageError); !ok {
		t.Fatalf("expected *UsageError, got %T (%v)", err, err)
	}
}
