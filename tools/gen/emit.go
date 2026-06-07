package main

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// reserved flag names that must not be generated because they collide with the
// root command's persistent flags or the shared --data flag.
var reserved = map[string]bool{
	"endpoint": true,
	"token":    true,
	"output":   true,
	"help":     true,
	"data":     true,
}

// emit writes the generated command files (one per group) and the registry.
func emit(cmds []Command, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := removeGenerated(outDir); err != nil {
		return err
	}

	groups := map[string][]Command{}
	for _, c := range cmds {
		groups[c.Group] = append(groups[c.Group], c)
	}
	order := make([]string, 0, len(groups))
	for g := range groups {
		order = append(order, g)
	}
	sort.Strings(order)

	for _, g := range order {
		if err := writeFormatted(filepath.Join(outDir, fileBase(g)+"_gen.go"), emitGroup(g, groups[g])); err != nil {
			return fmt.Errorf("group %s: %w", g, err)
		}
	}
	return writeFormatted(filepath.Join(outDir, "commands_gen.go"), emitRegistry(order))
}

// removeGenerated deletes previously generated files so stale commands do not
// linger across regenerations.
func removeGenerated(outDir string) error {
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), "_gen.go") {
			if err := os.Remove(filepath.Join(outDir, e.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeFormatted(path, src string) error {
	formatted, err := format.Source([]byte(src))
	if err != nil {
		return fmt.Errorf("gofmt failed: %w\n--- source ---\n%s", err, src)
	}
	return os.WriteFile(path, formatted, 0o644)
}

// emitRegistry produces commands_gen.go with the AddCommands entry point.
func emitRegistry(groups []string) string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("package commands\n\n")
	b.WriteString("import (\n\t\"github.com/spf13/cobra\"\n\n\t\"github.com/whatwedo/moco-cli/internal/cli\"\n)\n\n")
	b.WriteString("// AddCommands registers all generated group commands on the root command.\n")
	b.WriteString("func AddCommands(root *cobra.Command, app *cli.App) {\n\troot.AddCommand(\n")
	for _, g := range groups {
		fmt.Fprintf(&b, "\t\tnew%sCmd(app),\n", pascal(g))
	}
	b.WriteString("\t)\n}\n")
	return b.String()
}

// emitGroup produces the file for a single group.
func emitGroup(group string, cmds []Command) string {
	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })

	var b strings.Builder
	b.WriteString(header())
	b.WriteString("package commands\n\n")
	b.WriteString(imports(cmds))

	// Group command.
	fmt.Fprintf(&b, "func new%sCmd(app *cli.App) *cobra.Command {\n", pascal(group))
	fmt.Fprintf(&b, "\tcmd := &cobra.Command{\n\t\tUse:   %s,\n\t\tShort: %s,\n\t}\n",
		goString(group), goString(groupShort(cmds[0].Tag)))
	b.WriteString("\tcmd.AddCommand(\n")
	for _, c := range cmds {
		fmt.Fprintf(&b, "\t\tnew%s%sCmd(app),\n", pascal(group), pascal(c.Name))
	}
	b.WriteString("\t)\n\treturn cmd\n}\n\n")

	// Subcommands.
	for _, c := range cmds {
		b.WriteString(emitCommand(group, c))
		b.WriteString("\n")
	}
	return b.String()
}

// emitCommand produces the constructor for a single subcommand.
func emitCommand(group string, c Command) string {
	flags := selectFlags(c)

	var b strings.Builder
	fmt.Fprintf(&b, "func new%s%sCmd(app *cli.App) *cobra.Command {\n", pascal(group), pascal(c.Name))

	// Flag variables.
	for _, f := range flags {
		fmt.Fprintf(&b, "\tvar %s %s\n", f.varName, f.Type)
	}
	if c.HasBody {
		b.WriteString("\tvar data string\n")
	}

	// Command literal.
	b.WriteString("\tcmd := &cobra.Command{\n")
	fmt.Fprintf(&b, "\t\tUse:   %s,\n", goString(useString(c)))
	fmt.Fprintf(&b, "\t\tShort: %s,\n", goString(c.Short))
	fmt.Fprintf(&b, "\t\tArgs:  cobra.ExactArgs(%d),\n", len(c.PathParams))
	b.WriteString("\t\tRunE: func(cmd *cobra.Command, args []string) error {\n")
	b.WriteString(emitRunBody(c, flags))
	b.WriteString("\t\t},\n\t}\n")

	// Flag registration.
	if len(flags) > 0 || c.HasBody {
		b.WriteString("\tf := cmd.Flags()\n")
	}
	for _, f := range flags {
		fmt.Fprintf(&b, "\tf.%s(&%s, %s, %s, %s)\n", flagFunc(f.Type), f.varName, goString(f.Flag), flagDefault(f.Type), goString(f.Desc))
	}
	if c.HasBody {
		b.WriteString("\tf.StringVar(&data, \"data\", \"\", \"request body as JSON object\")\n")
	}

	b.WriteString("\treturn cmd\n}\n")
	return b.String()
}

// emitRunBody produces the RunE body: path assembly, query, body and the call.
func emitRunBody(c Command, flags []genFlag) string {
	var b strings.Builder

	// Path with substituted path parameters.
	fmt.Fprintf(&b, "\t\t\tpath := %s\n", goString(c.Path))
	for i, p := range c.PathParams {
		fmt.Fprintf(&b, "\t\t\tpath = strings.Replace(path, %s, url.PathEscape(args[%d]), 1)\n",
			goString("{"+p.Name+"}"), i)
	}

	// Query parameters.
	queryFlagsPresent := false
	for _, f := range flags {
		if f.kind == "query" {
			queryFlagsPresent = true
			break
		}
	}
	if queryFlagsPresent {
		b.WriteString("\t\t\tq := url.Values{}\n")
		for _, f := range flags {
			if f.kind != "query" {
				continue
			}
			if f.Type == "[]string" {
				fmt.Fprintf(&b, "\t\t\tfor _, v := range %s {\n\t\t\t\tq.Add(%s, v)\n\t\t\t}\n",
					f.varName, goString(f.Key))
				continue
			}
			fmt.Fprintf(&b, "\t\t\tif cmd.Flags().Changed(%s) {\n\t\t\t\tq.Set(%s, fmt.Sprint(%s))\n\t\t\t}\n",
				goString(f.Flag), goString(f.Key), f.varName)
		}
	}

	// Request body.
	if c.HasBody {
		b.WriteString("\t\t\tbody, err := cli.ParseData(data)\n\t\t\tif err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
		for _, f := range flags {
			if f.kind != "body" {
				continue
			}
			fmt.Fprintf(&b, "\t\t\tif cmd.Flags().Changed(%s) {\n\t\t\t\tbody[%s] = %s\n\t\t\t}\n",
				goString(f.Flag), goString(f.Key), f.varName)
		}
		b.WriteString("\t\t\traw, err := cli.EncodeBody(body)\n\t\t\tif err != nil {\n\t\t\t\treturn err\n\t\t\t}\n")
	}

	// Request.
	b.WriteString("\t\t\treturn app.Execute(cmd.Context(), client.Request{\n")
	fmt.Fprintf(&b, "\t\t\t\tMethod: %s,\n", goString(c.Method))
	b.WriteString("\t\t\t\tPath:   path,\n")
	if queryFlagsPresent {
		b.WriteString("\t\t\t\tQuery:  q,\n")
	}
	if c.HasBody {
		b.WriteString("\t\t\t\tBody:   raw,\n")
	}
	b.WriteString("\t\t\t})\n")
	return b.String()
}

// genFlag is a flag prepared for code generation.
type genFlag struct {
	Flag    string
	Key     string
	Type    string
	Desc    string
	kind    string // "query" or "body"
	varName string
}

// selectFlags merges query and body flags, dropping reserved or duplicate
// names so flag registration never conflicts.
func selectFlags(c Command) []genFlag {
	used := map[string]bool{}
	var out []genFlag

	add := func(f Flag, kind string) {
		if reserved[f.Flag] || used[f.Flag] {
			fmt.Fprintf(os.Stderr, "gen: skipping %s flag %q on %s/%s (reserved or duplicate)\n", kind, f.Flag, c.Group, c.Name)
			return
		}
		used[f.Flag] = true
		out = append(out, genFlag{
			Flag:    f.Flag,
			Key:     f.Key,
			Type:    f.Type,
			Desc:    f.Desc,
			kind:    kind,
			varName: varName(kind, len(out)),
		})
	}

	for _, f := range c.Query {
		add(f, "query")
	}
	for _, f := range c.Body {
		add(f, "body")
	}
	return out
}

func varName(kind string, idx int) string {
	if kind == "query" {
		return fmt.Sprintf("q%d", idx)
	}
	return fmt.Sprintf("v%d", idx)
}

func useString(c Command) string {
	parts := []string{c.Name}
	for _, p := range c.PathParams {
		parts = append(parts, "<"+p.Arg+">")
	}
	return strings.Join(parts, " ")
}

// imports computes the import block for a group file based on the features its
// commands use.
func imports(cmds []Command) string {
	needFmt, needURL, needStrings := false, false, false
	for _, c := range cmds {
		if len(c.PathParams) > 0 {
			needURL = true
			needStrings = true
		}
		for _, f := range selectFlags(c) {
			if f.kind == "query" {
				needURL = true
				if f.Type != "[]string" {
					needFmt = true // scalar query values are formatted with fmt.Sprint
				}
			}
		}
	}

	var std []string
	if needFmt {
		std = append(std, "\"fmt\"")
	}
	if needURL {
		std = append(std, "\"net/url\"")
	}
	if needStrings {
		std = append(std, "\"strings\"")
	}

	var b strings.Builder
	b.WriteString("import (\n")
	for _, s := range std {
		fmt.Fprintf(&b, "\t%s\n", s)
	}
	if len(std) > 0 {
		b.WriteString("\n")
	}
	b.WriteString("\t\"github.com/spf13/cobra\"\n\n")
	b.WriteString("\t\"github.com/whatwedo/moco-cli/internal/cli\"\n")
	b.WriteString("\t\"github.com/whatwedo/moco-cli/internal/client\"\n")
	b.WriteString(")\n\n")
	return b.String()
}

func flagFunc(t string) string {
	switch t {
	case "int":
		return "IntVar"
	case "bool":
		return "BoolVar"
	case "float64":
		return "Float64Var"
	case "[]string":
		return "StringArrayVar"
	default:
		return "StringVar"
	}
}

func flagDefault(t string) string {
	switch t {
	case "int":
		return "0"
	case "bool":
		return "false"
	case "float64":
		return "0"
	case "[]string":
		return "nil"
	default:
		return `""`
	}
}

// groupShort returns the German short description for a group command.
func groupShort(tag string) string {
	if n, ok := nouns[tag]; ok {
		return n.Plural + " verwalten"
	}
	return tag
}

func header() string {
	return "// Code generated by tools/gen; DO NOT EDIT.\n\n"
}

// pascal converts a kebab-case identifier to PascalCase.
func pascal(s string) string {
	var b strings.Builder
	for _, part := range strings.Split(s, "-") {
		if part == "" {
			continue
		}
		r := []rune(part)
		r[0] = []rune(strings.ToUpper(string(r[0])))[0]
		b.WriteString(string(r))
	}
	return b.String()
}

func fileBase(group string) string {
	return strings.ReplaceAll(group, "-", "_")
}

// goString renders s as a Go double-quoted string literal, preserving UTF-8
// (e.g. German umlauts) for readability.
func goString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
