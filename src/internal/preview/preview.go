// Package preview prepares issue and pull request bodies for Alfred UI's
// markdownarea, which is not GitHub/Forgejo markdown.
package preview

import (
	"regexp"
	"strings"
	"unicode"
)

// MaxBody caps the body sent per row: every row of a list goes to the window
// on each load, and the sheet is a preview, not a reader.
const MaxBody = 20000

var (
	htmlComment = regexp.MustCompile(`(?s)<!--.*?-->`)
	// Task list items; the markdownarea shows "[x]" literally.
	taskDone = regexp.MustCompile(`(?m)^(\s*[-*+]\s+)\[[xX]\]`)
	taskOpen = regexp.MustCompile(`(?m)^(\s*[-*+]\s+)\[ \]`)
)

// Body prepares a body for the markdownarea: issue templates' HTML comments
// are dropped, task list boxes become ☑/☐, identifiers with underscores become
// code spans, and long bodies are cut at a line break.
func Body(body string) string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	body = strings.TrimSpace(htmlComment.ReplaceAllString(body, ""))
	body = taskDone.ReplaceAllString(body, "${1}☑")
	body = taskOpen.ReplaceAllString(body, "${1}☐")
	body = codeIdentifiers(body)
	if body == "" {
		return "_No description provided._"
	}
	if len(body) > MaxBody {
		cut := body[:MaxBody]
		if i := strings.LastIndexByte(cut, '\n'); i > 0 {
			cut = cut[:i]
		}
		body = strings.ToValidUTF8(cut, "") + "\n\n_… truncated. Press ↩ to read it in Forgejo._"
	}
	return body
}

// lowLine stands in for "_" inside identifiers. The markdownarea pairs any
// two underscores as emphasis, even inside code spans, and ignores backslash
// escapes; U+02CD MODIFIER LETTER LOW MACRON looks the same in Menlo (same
// width) and isn't markdown. Preview only: copied text has the look-alike.
const lowLine = "ˍ"

// codeIdentifiers shows words with an underscore inside them, such as
// tc_redirect_* or local_delivery, as code. Fenced blocks, lines with code
// spans, links and HTML are left alone.
func codeIdentifiers(body string) string {
	lines := strings.Split(body, "\n")
	inFence := false
	for i, line := range lines {
		if t := strings.TrimSpace(line); strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence || !strings.Contains(line, "_") || strings.Contains(line, "`") {
			continue
		}
		words := strings.Split(line, " ")
		for j, w := range words {
			words[j] = codeWord(w)
		}
		lines[i] = strings.Join(words, " ")
	}
	return strings.Join(lines, "\n")
}

func codeWord(w string) string {
	if strings.ContainsAny(w, "<>[]()") || strings.Contains(w, "://") {
		return w
	}
	core := strings.TrimRight(w, ".,;:!?")
	rs := []rune(core)
	for k := 1; k < len(rs)-1; k++ {
		if rs[k] == '_' && isWordRune(rs[k-1]) {
			return "`" + strings.ReplaceAll(core, "_", lowLine) + "`" + w[len(core):]
		}
	}
	return w
}

func isWordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }
