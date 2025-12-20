package markdown

import "testing"

func TestQuote(t *testing.T) {
	t.Run("quote", func(t *testing.T) {
		t.Run("simple", test(`
> Bonsoir
`, `<div class="quote"><blockquote>Bonsoir</blockquote></div>`))
		t.Run("source", test(`
> Bonsoir, je suis un **code**
avec une source
`, `<div class="quote"><blockquote>Bonsoir, je suis un <b>code</b></blockquote><p>avec une source</p></div>`))
	})
}
