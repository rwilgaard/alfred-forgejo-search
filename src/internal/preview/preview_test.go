package preview

import (
	"strings"
	"testing"
)

func TestBody(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"empty", "  ", "_No description provided._"},
		{"comment only", "<!-- template -->\r\n", "_No description provided._"},
		{"tasks", "- [x] done\n- [ ] todo", "- ☑ done\n- ☐ todo"},
		{"identifier", "set local_delivery.", "set `local" + lowLine + "delivery`."},
		{"emphasis kept", "_really_ works", "_really_ works"},
		{"code span line untouched", "use `a_b` and c_d", "use `a_b` and c_d"},
		{"fence untouched", "```\na_b\n```", "```\na_b\n```"},
		{"url untouched", "see https://x/a_b", "see https://x/a_b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Body(tt.in); got != tt.want {
				t.Errorf("Body(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBodyTruncates(t *testing.T) {
	in := strings.Repeat("line\n", MaxBody/4)
	got := Body(in)
	if len(got) > MaxBody+100 || !strings.HasSuffix(got, "Press ↩ to read it in Forgejo._") {
		t.Errorf("Body did not truncate: len %d, tail %q", len(got), got[len(got)-40:])
	}
}
