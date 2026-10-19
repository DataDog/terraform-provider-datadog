package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// schemaDoc renders a draft 2020-12 JSON Schema as Markdown. It is deliberately
// narrow: it handles the constructs the two checked-in contracts actually use
// (objects with properties, $defs, $ref, anyOf/oneOf, arrays, enums, defaults,
// required lists, additionalProperties) and ignores the rest of the spec.
//
// The point is provenance, not generality. The published annotation reference is
// now a projection of internal/contracts/tracking-field.schema.json, so a field
// that is not in the schema cannot appear in the docs, and a field that is added
// to the schema shows up in the docs on the next `make tfgen-docs`.
type schemaDoc struct {
	root map[string]any
	defs map[string]any
}

func newSchemaDoc(raw []byte) (*schemaDoc, error) {
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("parsing schema: %w", err)
	}
	defs, _ := root["$defs"].(map[string]any)
	return &schemaDoc{root: root, defs: defs}, nil
}

// renderBody writes the object tables for the root schema followed by one
// section per $def, in a stable order.
func (s *schemaDoc) renderBody(b *strings.Builder) {
	if desc := str(s.root["description"]); desc != "" {
		b.WriteString(desc + "\n\n")
	}

	if _, ok := s.root["properties"]; ok {
		b.WriteString("## Fields\n\n")
		s.writeObject(b, s.root, 2)
	}

	// A root that is only a choice between shapes (oneOf/anyOf of $refs) has no
	// properties of its own; name the alternatives so the reader knows which
	// $def section applies to them.
	for _, key := range []string{"oneOf", "anyOf"} {
		if alts, ok := s.root[key].([]any); ok && len(alts) > 0 {
			fmt.Fprintf(b, "This extension takes one of %d shapes:\n\n", len(alts))
			for _, alt := range alts {
				if m, ok := alt.(map[string]any); ok {
					if ref := str(m["$ref"]); ref != "" {
						name := refName(ref)
						fmt.Fprintf(b, "- [%s](#%s)\n", name, anchor(name))
					}
				}
			}
			b.WriteString("\n")
		}
	}

	if len(s.defs) == 0 {
		return
	}
	names := make([]string, 0, len(s.defs))
	for n := range s.defs {
		names = append(names, n)
	}
	sort.Strings(names)

	b.WriteString("## Component shapes\n\n")
	b.WriteString("Referenced by the fields above.\n\n")
	for _, n := range names {
		def, _ := s.defs[n].(map[string]any)
		if def == nil {
			continue
		}
		title := str(def["title"])
		if title == "" {
			title = n
		}
		fmt.Fprintf(b, "### %s\n\n", title)
		if desc := str(def["description"]); desc != "" {
			b.WriteString(desc + "\n\n")
		}
		fmt.Fprintf(b, "`$defs.%s`", n)
		if req := requiredList(def); len(req) > 0 {
			fmt.Fprintf(b, " · required: %s", strings.Join(backtickAll(req), ", "))
		}
		if ap, ok := def["additionalProperties"].(bool); ok && !ap {
			b.WriteString(" · unknown fields rejected")
		}
		b.WriteString("\n\n")

		if _, ok := def["properties"]; ok {
			s.writeObject(b, def, 3)
		}
		s.writeChoices(b, def)
	}
}

// writeObject renders one object's properties as a table, then recurses into
// any property that is an inline object with its own fields. `group` is the
// reason this recursion exists: it is the most consequential field in the
// annotation and its sub-fields are declared inline rather than via $ref, so a
// flat table would show it as a bare "object" and say nothing useful.
//
// level is the Markdown heading depth of the section this table sits under, so
// nested sections stay correctly ranked whether they hang off the root Fields
// table or off a $defs section.
func (s *schemaDoc) writeObject(b *strings.Builder, obj map[string]any, level int) {
	props, _ := obj["properties"].(map[string]any)
	if len(props) == 0 {
		return
	}
	req := map[string]bool{}
	for _, r := range requiredList(obj) {
		req[r] = true
	}

	names := make([]string, 0, len(props))
	for n := range props {
		names = append(names, n)
	}
	sort.Strings(names)

	b.WriteString("| Field | Type | Required | Default | Notes |\n|---|---|---|---|---|\n")
	for _, n := range names {
		p, _ := props[n].(map[string]any)
		if p == nil {
			continue
		}
		required := "no"
		if req[n] {
			required = "**yes**"
		}
		def := "none"
		if d, ok := p["default"]; ok {
			def = "`" + fmt.Sprintf("%v", d) + "`"
		}
		fmt.Fprintf(b, "| `%s` | %s | %s | %s | %s |\n",
			n, s.typeOf(p), required, def, s.notesFor(p))
	}
	b.WriteString("\n")

	if ap, ok := obj["additionalProperties"].(bool); ok && !ap {
		b.WriteString("Unknown fields are rejected. A field absent from this table is not supported: " +
			"passing one fails the run at parse time.\n\n")
	}
	s.writeChoices(b, obj)

	// Expand inline sub-objects. $ref'd shapes are not expanded here; they get
	// their own section under "Component shapes".
	heading := strings.Repeat("#", level+1)
	for _, n := range names {
		p, _ := props[n].(map[string]any)
		if p == nil || str(p["$ref"]) != "" {
			continue
		}
		if sub, ok := p["properties"].(map[string]any); !ok || len(sub) == 0 {
			continue
		}
		fmt.Fprintf(b, "%s `%s`\n\n", heading, n)
		if desc := str(p["description"]); desc != "" {
			b.WriteString(desc + "\n\n")
		}
		s.writeObject(b, p, level+1)
	}
}

// writeChoices surfaces an object-level anyOf/oneOf constraint, which is how the
// annotation contract expresses \"a group must name read or search\".
func (s *schemaDoc) writeChoices(b *strings.Builder, obj map[string]any) {
	for _, key := range []string{"anyOf", "oneOf"} {
		alts, ok := obj[key].([]any)
		if !ok || len(alts) == 0 {
			continue
		}
		var clauses []string
		for _, alt := range alts {
			m, ok := alt.(map[string]any)
			if !ok {
				continue
			}
			if req := requiredList(m); len(req) > 0 {
				clauses = append(clauses, strings.Join(backtickAll(req), " + "))
			}
		}
		if len(clauses) == 0 {
			continue
		}
		word := "at least one of"
		if key == "oneOf" {
			word = "exactly one of"
		}
		fmt.Fprintf(b, "Constraint: %s %s must be present.\n\n", word, strings.Join(clauses, " or "))
	}
}

func (s *schemaDoc) typeOf(p map[string]any) string {
	if ref := str(p["$ref"]); ref != "" {
		n := refName(ref)
		return fmt.Sprintf("[%s](#%s)", n, anchor(n))
	}
	t := str(p["type"])
	if t == "array" {
		if items, ok := p["items"].(map[string]any); ok {
			return "array of " + strings.TrimPrefix(s.typeOf(items), "")
		}
		return "array"
	}
	if t == "" {
		return "unspecified"
	}
	return "`" + t + "`"
}

func (s *schemaDoc) notesFor(p map[string]any) string {
	var parts []string
	if desc := str(p["description"]); desc != "" {
		parts = append(parts, desc)
	}
	if enum, ok := p["enum"].([]any); ok && len(enum) > 0 {
		vals := make([]string, 0, len(enum))
		for _, e := range enum {
			vals = append(vals, "`"+fmt.Sprintf("%v", e)+"`")
		}
		parts = append(parts, "One of: "+strings.Join(vals, ", ")+".")
	}
	if pat := str(p["pattern"]); pat != "" {
		parts = append(parts, "Pattern: `"+pat+"`.")
	}
	var bounds []string
	if v, ok := p["minLength"]; ok {
		bounds = append(bounds, fmt.Sprintf("min length %v", v))
	}
	if v, ok := p["maxLength"]; ok {
		bounds = append(bounds, fmt.Sprintf("max length %v", v))
	}
	if u, ok := p["uniqueItems"].(bool); ok && u {
		bounds = append(bounds, "unique items")
	}
	if len(bounds) > 0 {
		parts = append(parts, capitalise(strings.Join(bounds, ", "))+".")
	}
	if items, ok := p["items"].(map[string]any); ok {
		if sub := s.notesFor(items); sub != "" && str(items["description"]) != "" {
			parts = append(parts, "Items: "+sub)
		}
	}
	return escapePipes(strings.Join(parts, " "))
}

func requiredList(obj map[string]any) []string {
	raw, ok := obj["required"].([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, r := range raw {
		if s := str(r); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func backtickAll(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = "`" + s + "`"
	}
	return out
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func refName(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

func anchor(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, s)
}

func capitalise(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
