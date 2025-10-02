package markdown

import "testing"

func TestQuote(t *testing.T) {
	content := `
> Bonsoir, je suis un **code**
avec une source
`
	lxs := lex(content)
	tree, err := ast(lxs)
	if err != nil {
		t.Fatal(err)
	}
	c, err := tree.Eval()
	if err != nil {
		t.Fatal(err)
	}
	if c != `<div class="quote"><blockquote>Bonsoir, je suis un <b>code</b></blockquote><p>avec une source</p></div>` {
		t.Errorf("failed, got %s", c)
		t.Logf("lxs: %s\ntree: %s", lxs, tree)
	}
}
