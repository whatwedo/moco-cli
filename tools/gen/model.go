package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// Command is the generator-internal representation of a CLI command.
type Command struct {
	Tag     string // original tag, e.g. "AccountCatalogServices"
	Group   string // kebab-case group, e.g. "account-catalog-services"
	Name    string // verb, e.g. "list", "get", "start-timer"
	Method  string // HTTP method, e.g. "GET"
	Path    string // path template, e.g. "/projects/{id}"
	Summary string // English summary from the spec (fallback description)
	Short   string // German short description

	PathParams []PathParam
	Query      []Flag
	Body       []Flag
	HasBody    bool // operation accepts a JSON body (--data available)
}

// PathParam is a positional argument taken from the path template.
type PathParam struct {
	Name string // original name in the template, e.g. "task_id"
	Arg  string // display name, e.g. "task-id"
}

// Flag is a query or body parameter exposed as a CLI flag.
type Flag struct {
	Flag     string // kebab-case flag name
	Key      string // original name (query key or body property)
	Type     string // "string", "int", "float64", "bool"
	Required bool
	Desc     string
}

var pathParamRe = regexp.MustCompile(`\{([^}]+)\}`)

// BuildCommands derives the complete, deterministic command list (sorted,
// collision-free) from the spec.
func BuildCommands(spec *Spec, tr *Translations) []Command {
	var cmds []Command

	paths := make([]string, 0, len(spec.Paths))
	for p := range spec.Paths {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	baseLen := groupBaseLengths(spec, paths)

	for _, path := range paths {
		item := spec.Paths[path]
		for _, mo := range item.operations() {
			tag := operationTag(mo.Op)
			cmds = append(cmds, buildCommand(spec, path, mo.Method, mo.Op, item.Parameters, baseLen[tag]))
		}
	}

	resolveCollisions(cmds)
	for i := range cmds {
		cmds[i].Short = tr.short(&cmds[i])
	}
	return cmds
}

func buildCommand(spec *Spec, path, method string, op *Operation, shared []*Parameter, baseLen int) Command {
	tag := operationTag(op)

	name := commandName(path, baseLen, method, op.Summary)
	if override, ok := nameOverrides[method+" "+path]; ok {
		name = override
	}

	cmd := Command{
		Tag:     tag,
		Group:   kebab(tag),
		Method:  method,
		Path:    path,
		Summary: op.Summary,
		Name:    name,
	}

	params := append(append([]*Parameter{}, shared...), op.Parameters...)
	params = resolveParams(spec, params)

	cmd.PathParams = pathParams(path)
	cmd.Query = queryFlags(params)

	if op.RequestBody != nil {
		if mt, ok := op.RequestBody.Content["application/json"]; ok && mt.Schema != nil {
			cmd.HasBody = true
			cmd.Body = bodyFlags(spec, mt.Schema)
		}
	}
	return cmd
}

// operationTag returns the first tag of an operation (group assignment).
func operationTag(op *Operation) string {
	if len(op.Tags) > 0 {
		return op.Tags[0]
	}
	return "Misc"
}

// groupBaseLengths determines, per tag, the length of the base prefix. Literals
// beyond this prefix qualify sub-resources or actions.
//
// The prefix is computed over the paths sharing the most common first segment.
// This keeps occasional global shortcut endpoints (e.g. /recurring_expenses
// alongside /projects/{id}/recurring_expenses) from collapsing the prefix to
// zero.
func groupBaseLengths(spec *Spec, paths []string) map[string]int {
	byTag := map[string][][]string{}
	for _, path := range paths {
		for _, mo := range spec.Paths[path].operations() {
			tag := operationTag(mo.Op)
			byTag[tag] = append(byTag[tag], literalSegments(path))
		}
	}
	out := map[string]int{}
	for tag, lists := range byTag {
		out[tag] = commonPrefixLen(majorityFirstSegment(lists))
	}
	return out
}

// majorityFirstSegment returns the literal lists sharing the most common first
// segment (ties broken by the lexicographically smallest segment).
func majorityFirstSegment(lists [][]string) [][]string {
	counts := map[string]int{}
	for _, l := range lists {
		if len(l) > 0 {
			counts[l[0]]++
		}
	}
	best := ""
	bestN := -1
	for seg, n := range counts {
		if n > bestN || (n == bestN && seg < best) {
			best, bestN = seg, n
		}
	}
	var out [][]string
	for _, l := range lists {
		if len(l) > 0 && l[0] == best {
			out = append(out, l)
		}
	}
	return out
}

func commonPrefixLen(lists [][]string) int {
	if len(lists) == 0 {
		return 0
	}
	n := len(lists[0])
	for _, l := range lists[1:] {
		n = min(n, len(l))
		for i := 0; i < n; i++ {
			if l[i] != lists[0][i] {
				n = i
				break
			}
		}
	}
	return n
}

// commandName builds the command name from a sub-resource qualifier and a verb.
//
// Literals beyond the common group prefix (baseLen) qualify sub-resources or
// actions. CRUD verbs are derived from the summary (PUT->update, PATCH->patch);
// otherwise the qualifier itself serves as the name.
func commandName(path string, baseLen int, method, summary string) string {
	extra := literalSegments(path)
	if baseLen <= len(extra) {
		extra = extra[baseLen:]
	} else {
		extra = nil
	}

	verb := crudVerb(method, summary)
	if verb == "" {
		if len(extra) == 0 {
			return methodVerb(method)
		}
		return kebab(strings.Join(extra, "-"))
	}
	if len(extra) == 0 {
		return verb
	}
	return kebab(strings.Join(extra, "-")) + "-" + verb
}

// crudVerb returns the standard CRUD verb based on the summary, otherwise "".
func crudVerb(method, summary string) string {
	switch firstWord(summary) {
	case "List":
		return "list"
	case "Get":
		return "get"
	case "Create":
		return "create"
	case "Delete":
		return "delete"
	case "Update":
		if method == "PATCH" {
			return "patch"
		}
		return "update"
	}
	return ""
}

// methodVerb returns a verb based on the HTTP method, for actions without a
// sub-resource qualifier.
func methodVerb(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT":
		return "update"
	case "PATCH":
		return "patch"
	case "DELETE":
		return "delete"
	default:
		return "get"
	}
}

// resolveCollisions ensures command names are unique within each group. On
// conflicts the path is used as a discriminator, then the method, and finally a
// numeric suffix as a guaranteed-unique fallback.
func resolveCollisions(cmds []Command) {
	rekeyByPath := func() map[string][]int {
		m := map[string][]int{}
		for i := range cmds {
			k := cmds[i].Group + "/" + cmds[i].Name
			m[k] = append(m[k], i)
		}
		return m
	}

	for _, idxs := range rekeyByPath() {
		if len(idxs) < 2 {
			continue
		}
		for _, i := range idxs {
			cmds[i].Name = kebab(strings.Join(literalSegments(cmds[i].Path), "-"))
		}
	}

	for _, idxs := range rekeyByPath() {
		if len(idxs) < 2 {
			continue
		}
		for _, i := range idxs {
			cmds[i].Name = cmds[i].Name + "-" + strings.ToLower(cmds[i].Method)
		}
	}

	for _, idxs := range rekeyByPath() {
		if len(idxs) < 2 {
			continue
		}
		sort.Slice(idxs, func(a, b int) bool { return cmds[idxs[a]].Path < cmds[idxs[b]].Path })
		for n, i := range idxs[1:] {
			cmds[i].Name = fmt.Sprintf("%s-%d", cmds[i].Name, n+2)
			fmt.Fprintf(os.Stderr, "gen: resolved name collision: %s/%s (%s %s)\n",
				cmds[i].Group, cmds[i].Name, cmds[i].Method, cmds[i].Path)
		}
	}
}

func resolveParams(spec *Spec, params []*Parameter) []*Parameter {
	out := make([]*Parameter, 0, len(params))
	for _, p := range params {
		if p.Ref != "" {
			// Parameter refs do not occur in the MOCO spec; ignore defensively.
			continue
		}
		out = append(out, p)
	}
	return out
}

func pathParams(path string) []PathParam {
	var out []PathParam
	for _, m := range pathParamRe.FindAllStringSubmatch(path, -1) {
		name := m[1]
		out = append(out, PathParam{Name: name, Arg: kebab(name)})
	}
	return out
}

func queryFlags(params []*Parameter) []Flag {
	var out []Flag
	for _, p := range params {
		if p.In != "query" {
			continue
		}
		out = append(out, Flag{
			Flag:     kebab(p.Name),
			Key:      p.Name,
			Type:     goType(p.Schema),
			Required: p.Required,
			Desc:     firstLine(p.Description),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Flag < out[j].Flag })
	return out
}

func bodyFlags(spec *Spec, schema *Schema) []Flag {
	schema = spec.resolve(schema)
	if schema == nil {
		return nil
	}
	required := map[string]bool{}
	for _, r := range schema.Required {
		required[r] = true
	}
	var out []Flag
	for _, name := range spec.sortedSchemaProps(schema) {
		prop := spec.resolve(schema.Properties[name])
		if prop == nil || !isScalar(prop.Type) {
			continue // nested structures only via --data
		}
		out = append(out, Flag{
			Flag:     kebab(name),
			Key:      name,
			Type:     goType(prop),
			Required: required[name],
			Desc:     firstLine(prop.Description),
		})
	}
	return out
}

func isScalar(t string) bool {
	switch t {
	case "string", "integer", "number", "boolean":
		return true
	}
	return false
}

func goType(s *Schema) string {
	if s == nil {
		return "string"
	}
	switch s.Type {
	case "integer":
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "array":
		return "[]string"
	default:
		return "string"
	}
}

func firstWord(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

// literalSegments returns the literal path segments. Placeholders ({id}) are
// removed; a file extension on a placeholder ("{id}.pdf") is kept as a literal
// ("pdf").
func literalSegments(path string) []string {
	var out []string
	for _, seg := range strings.Split(strings.Trim(path, "/"), "/") {
		lit := strings.Trim(pathParamRe.ReplaceAllString(seg, ""), ".")
		if lit == "" {
			continue
		}
		out = append(out, lit)
	}
	return out
}

// kebab converts an identifier (CamelCase, snake_case, with spaces) to
// kebab-case.
func kebab(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		switch {
		case r == '_' || r == ' ' || r == '/' || r == '.':
			b.WriteByte('-')
		case unicode.IsUpper(r):
			if i > 0 {
				prev := runes[i-1]
				var next rune
				if i+1 < len(runes) {
					next = runes[i+1]
				}
				if unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && unicode.IsLower(next)) {
					b.WriteByte('-')
				}
			}
			b.WriteRune(unicode.ToLower(r))
		default:
			b.WriteRune(r)
		}
	}
	out := b.String()
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	return strings.Trim(out, "-")
}
