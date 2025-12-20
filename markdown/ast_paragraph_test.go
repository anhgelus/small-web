package markdown

import "testing"

func TestParagraph(t *testing.T) {
	t.Run("paragraph", func(t *testing.T) {
		t.Run("simple", test("bonsoir", `<p>bonsoir</p>`))
	})
	t.Run("replacer", func(t *testing.T) {
		opt := &Option{
			Replaces: map[rune]string{'~': "&thinsp;"},
		}
		t.Run("empty", testWithOptions(opt, "bonsoir", `<p>bonsoir</p>`))
		t.Run("simple", testWithOptions(opt, "bonsoir~!", `<p>bonsoir&thinsp;!</p>`))
	})
}
