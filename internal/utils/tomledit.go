package utils

import (
	"fmt"
	"strings"
)

// SetTOMLValue sets key = literal inside [section] of a TOML document, keeping comments, order and formatting of
// everything else. literal must already be TOML (a quoted string, a number, true/false). An existing line keeps its
// indentation and trailing comment; a missing key is added after the section's last key; a missing section is
// appended. Only plain "key = value" lines of a "[section]" table are handled, which is what config.toml uses.
func SetTOMLValue(doc, section, key, literal string) (string, error) {
	if strings.ContainsAny(literal, "\n\r") {
		return "", fmt.Errorf("value for %s.%s must be a single line", section, key)
	}
	lines := strings.Split(doc, "\n")
	start, end := -1, len(lines) // [start, end): the section's lines after its header
	for i, l := range lines {
		t := strings.TrimSpace(l)
		if !strings.HasPrefix(t, "[") {
			continue
		}
		if start >= 0 {
			end = i
			break
		}
		if name := strings.TrimSpace(strings.Trim(stripComment(t), "[]")); name == section {
			start = i + 1
		}
	}
	if start < 0 {
		trimmed := strings.TrimRight(doc, "\n")
		return trimmed + "\n\n[" + section + "]\n" + key + " = " + literal + "\n", nil
	}
	last := start - 1
	for i := start; i < end; i++ {
		t := strings.TrimSpace(lines[i])
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		last = i
		k, _, ok := strings.Cut(t, "=")
		if !ok || strings.TrimSpace(k) != key {
			continue
		}
		eq := strings.Index(lines[i], "=")
		lhs := lines[i][:eq] // keeps indentation and the spacing before '=' (aligned keys)
		comment := ""
		if c := trailingComment(lines[i][eq+1:]); c != "" {
			comment = " " + c
		}
		lines[i] = lhs + "= " + literal + comment
		return strings.Join(lines, "\n"), nil
	}
	ins := append([]string{}, lines[:last+1]...)
	ins = append(ins, key+" = "+literal)
	return strings.Join(append(ins, lines[last+1:]...), "\n"), nil
}

// stripComment drops a trailing "# ..." outside quotes.
func stripComment(s string) string {
	if c := trailingComment(s); c != "" {
		return strings.TrimSpace(strings.TrimSuffix(strings.TrimRight(s, " \t"), c))
	}
	return s
}

// trailingComment returns the "# ..." part of a value, ignoring '#' inside basic or literal strings.
func trailingComment(s string) string {
	var quote byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0:
			if c == '\\' && quote == '"' {
				i++
			} else if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '#':
			return strings.TrimRight(s[i:], " \t")
		}
	}
	return ""
}
