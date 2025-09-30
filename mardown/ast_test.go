package mardown

import "testing"

func TestAst(t *testing.T) {
	content := "bonsoir"
	lxs := lex(content)
	tree, err := ast(lxs)
	if err != nil {
		t.Fatal(err)
	}
	c, err := tree.Eval()
	if err != nil {
		t.Fatal(err)
	}
	if c != "<p>bonsoir</p>" {
		t.Errorf("failed, got %s", c)
		t.Logf("lxs: %s\ntree: %s", lxs, tree)
	}
	content = `
***bon*soir**
`
	lxs = lex(content)
	tree, err = ast(lxs)
	if err != nil {
		t.Fatal(err)
	}
	c, err = tree.Eval()
	if err != nil {
		t.Fatal(err)
	}
	if c != "<p><b><em>bon</em>soir</b></p>" {
		t.Errorf("failed, got %s", c)
		t.Logf("lxs: %s\ntree: %s", lxs, tree)
	}
	content = "je suis un `code`"
	lxs = lex(content)
	tree, err = ast(lxs)
	if err != nil {
		t.Fatal(err)
	}
	c, err = tree.Eval()
	if err != nil {
		t.Fatal(err)
	}
	if c != "<p>je suis un <code>code</code></p>" {
		t.Errorf("failed, got %s", c)
		t.Logf("lxs: %s\ntree: %s", lxs, tree)
	}
	content = `
> Bonsoir, je suis un **code**
avec une source
`
	lxs = lex(content)
	tree, err = ast(lxs)
	if err != nil {
		t.Fatal(err)
	}
	c, err = tree.Eval()
	if err != nil {
		t.Fatal(err)
	}
	if c != `<div class="quote"><blockquote>Bonsoir, je suis un <b>code</b></blockquote><p>avec une source</p></div>` {
		t.Errorf("failed, got %s", c)
		t.Logf("lxs: %s\ntree: %s", lxs, tree)
	}
}
