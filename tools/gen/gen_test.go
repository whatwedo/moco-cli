package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestKebab(t *testing.T) {
	cases := map[string]string{
		"AccountCatalogServices":  "account-catalog-services",
		"AccountWebHooks":         "account-web-hooks",
		"VatCodeSales":            "vat-code-sales",
		"UserWorkTimeAdjustments": "user-work-time-adjustments",
		"start_timer":             "start-timer",
		"per_page":                "per-page",
		"timesheet.pdf":           "timesheet-pdf",
	}
	for in, want := range cases {
		if got := kebab(in); got != want {
			t.Errorf("kebab(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLiteralSegments(t *testing.T) {
	cases := map[string][]string{
		"/projects":            {"projects"},
		"/projects/{id}":       {"projects"},
		"/invoices/{id}.pdf":   {"invoices", "pdf"},
		"/invoices/{id}/x.pdf": {"invoices", "x.pdf"},
		"/a/{id}/b/{bid}/c":    {"a", "b", "c"},
	}
	for path, want := range cases {
		got := literalSegments(path)
		if len(got) != len(want) {
			t.Errorf("literalSegments(%q) = %v, want %v", path, got, want)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("literalSegments(%q) = %v, want %v", path, got, want)
				break
			}
		}
	}
}

func TestCommandName(t *testing.T) {
	cases := []struct {
		path    string
		baseLen int
		method  string
		summary string
		want    string
	}{
		{"/projects", 1, "GET", "List projects", "list"},
		{"/projects/{id}", 1, "GET", "Get project", "get"},
		{"/projects", 1, "POST", "Create project", "create"},
		{"/projects/{id}", 1, "PUT", "Update project", "update"},
		{"/projects/{id}", 1, "PATCH", "Update project", "patch"},
		{"/projects/{id}", 1, "DELETE", "Delete project", "delete"},
		{"/projects/{id}/archive", 1, "PUT", "Archive project", "archive"},
		{"/x/{id}/items", 1, "GET", "List items", "items-list"},
		{"/x/{id}/items/{id}", 1, "GET", "Get item", "items-get"},
		// Non-CRUD action without a qualifier falls back to the HTTP method.
		{"/taggings/{type}/{id}", 1, "PUT", "Replace taggings", "update"},
		{"/session", 1, "POST", "Verify API key", "create"},
		{"/session", 1, "DELETE", "Sign out", "delete"},
	}
	for _, c := range cases {
		if got := commandName(c.path, c.baseLen, c.method, c.summary); got != c.want {
			t.Errorf("commandName(%q, %d, %q, %q) = %q, want %q", c.path, c.baseLen, c.method, c.summary, got, c.want)
		}
	}
}

func TestBuildCommandsIntegration(t *testing.T) {
	spec := &Spec{Paths: map[string]*PathItem{
		"/projects": {
			Get:  &Operation{Tags: []string{"Projects"}, Summary: "List projects"},
			Post: &Operation{Tags: []string{"Projects"}, Summary: "Create project", RequestBody: jsonBody("name", "active")},
		},
		"/projects/{id}": {
			Get:    &Operation{Tags: []string{"Projects"}, Summary: "Get project"},
			Delete: &Operation{Tags: []string{"Projects"}, Summary: "Delete project"},
		},
		"/projects/{id}/archive": {
			Put: &Operation{Tags: []string{"Projects"}, Summary: "Archive project"},
		},
	}}
	tr := &Translations{
		Nouns:   map[string]Noun{"Projects": {"Projekt", "Projekte"}},
		Actions: map[string]string{"projects/archive": "Projekt archivieren"},
	}

	cmds := BuildCommands(spec, tr)
	byName := map[string]Command{}
	for _, c := range cmds {
		byName[c.Name] = c
	}

	if got := byName["list"].Short; got != "Projekte auflisten" {
		t.Errorf("list short = %q", got)
	}
	if got := byName["archive"].Short; got != "Projekt archivieren" {
		t.Errorf("archive short = %q", got)
	}
	create, ok := byName["create"]
	if !ok || !create.HasBody {
		t.Fatalf("create command missing or has no body")
	}
	if len(create.Body) != 2 || create.Body[0].Flag != "active" || create.Body[1].Flag != "name" {
		t.Errorf("create body flags = %+v", create.Body)
	}
	if get := byName["get"]; len(get.PathParams) != 1 || get.PathParams[0].Name != "id" {
		t.Errorf("get path params = %+v", byName["get"].PathParams)
	}
}

func jsonBody(props ...string) *RequestBody {
	schema := &Schema{Type: "object", Properties: map[string]*Schema{}}
	for _, p := range props {
		schema.Properties[p] = &Schema{Type: "string"}
	}
	return &RequestBody{Content: map[string]*MediaType{"application/json": {Schema: schema}}}
}

func TestBodyFlagsSkipsNonScalar(t *testing.T) {
	schema := &Schema{Type: "object", Properties: map[string]*Schema{
		"name":  {Type: "string"},
		"count": {Type: "integer"},
		"tags":  {Type: "array"},
		"meta":  {Type: "object"},
	}}
	got := bodyFlags(&Spec{}, schema)
	if len(got) != 2 {
		t.Fatalf("expected 2 scalar flags, got %d: %+v", len(got), got)
	}
	if got[0].Flag != "count" || got[1].Flag != "name" {
		t.Errorf("flags = %+v, want count, name", got)
	}
}

func TestSelectFlagsDropsReservedAndDuplicates(t *testing.T) {
	c := Command{
		Group:   "g",
		Name:    "x",
		HasBody: true,
		Query:   []Flag{{Flag: "page", Key: "page", Type: "int"}, {Flag: "token", Key: "token", Type: "string"}},
		Body:    []Flag{{Flag: "page", Key: "page", Type: "string"}, {Flag: "name", Key: "name", Type: "string"}},
	}
	got := selectFlags(c)
	kinds := map[string]string{}
	for _, f := range got {
		kinds[f.Flag] = f.kind
	}
	// "token" is reserved (--data flag), "page" body flag duplicates the query flag.
	if len(got) != 2 {
		t.Fatalf("expected 2 flags, got %d: %+v", len(got), got)
	}
	if kinds["page"] != "query" || kinds["name"] != "body" {
		t.Errorf("kinds = %v", kinds)
	}
	if _, ok := kinds["token"]; ok {
		t.Error("reserved flag 'token' should have been dropped")
	}
}

func TestGoString(t *testing.T) {
	cases := map[string]string{
		`abc`:       `"abc"`,
		`a"b`:       `"a\"b"`,
		`a\b`:       `"a\\b"`,
		"Größe":     `"Größe"`, // UTF-8 preserved
		"line\nbrk": `"line\nbrk"`,
	}
	for in, want := range cases {
		if got := goString(in); got != want {
			t.Errorf("goString(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestEmitGroupProducesValidGo(t *testing.T) {
	cmds := []Command{
		{Group: "projects", Name: "get", Method: "GET", Path: "/projects/{id}", Short: "Projekt abrufen",
			PathParams: []PathParam{{Name: "id", Arg: "id"}}},
		{Group: "projects", Name: "list", Method: "GET", Path: "/projects", Short: "Projekte auflisten",
			Query: []Flag{{Flag: "page", Key: "page", Type: "int"}, {Flag: "ids", Key: "ids", Type: "[]string"}}},
		{Group: "projects", Name: "create", Method: "POST", Path: "/projects", Short: "Projekt erstellen",
			HasBody: true, Body: []Flag{{Flag: "name", Key: "name", Type: "string"}}},
	}

	src := emitGroup("projects", cmds)
	if _, err := parser.ParseFile(token.NewFileSet(), "projects_gen.go", src, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, src)
	}

	for _, want := range []string{
		"func newProjectsCmd",
		"url.PathEscape(args[0])",
		"f.StringArrayVar(&q1, \"ids\"",
		"cli.ParseData(data)",
		`Short: "Projekt abrufen"`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("generated source missing %q", want)
		}
	}
}
