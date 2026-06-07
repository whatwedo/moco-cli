package commands_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/whatwedo/moco-cli/internal/cli"
	"github.com/whatwedo/moco-cli/internal/client"
	"github.com/whatwedo/moco-cli/internal/commands"
)

// run executes the CLI with the given args against srv and returns stdout.
func run(t *testing.T, srvURL string, args ...string) (string, error) {
	t.Helper()
	root, app := cli.NewRootCmd("test")
	commands.AddCommands(root, app)

	var out bytes.Buffer
	app.Out = &out
	app.Client = client.New(srvURL+"/api/v1", "tok", "test")
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs(args)

	err := root.Execute()
	return out.String(), err
}

func TestGetSubstitutesPathParam(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		_, _ = w.Write([]byte(`{"id":123}`))
	}))
	defer srv.Close()

	out, err := run(t, srv.URL, "projects", "get", "123", "--output", "raw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/v1/projects/123" {
		t.Errorf("request = %s %s", gotMethod, gotPath)
	}
	if out != "{\"id\":123}\n" {
		t.Errorf("output = %q", out)
	}
}

func TestListSendsQueryFlag(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	if _, err := run(t, srv.URL, "projects", "list", "--page", "3"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "page=3" {
		t.Errorf("query = %q", gotQuery)
	}
}

func TestCreateMergesBodyFlagsOverData(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	// --data sets the base body, the --name flag overrides one field.
	_, err := run(t, srv.URL, "companies", "create",
		"--data", `{"type":"customer","name":"old"}`, "--name", "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody != `{"name":"new","type":"customer"}` {
		t.Errorf("body = %q", gotBody)
	}
}

func TestAPIErrorReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer srv.Close()

	if _, err := run(t, srv.URL, "projects", "get", "999"); err == nil {
		t.Fatal("expected an error for HTTP 404")
	}
}
