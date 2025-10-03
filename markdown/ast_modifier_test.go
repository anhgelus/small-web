package markdown

import "testing"

func TestModifier(t *testing.T) {
	content := `
**bo*n*soir**, ça ***va* bien** ?
`
	lxs := lex(content)
	tree, err := ast(lxs)
	if err != nil {
		t.Fatal(err)
	}
	c, err := tree.Eval(nil)
	if err != nil {
		t.Fatal(err)
	}
	if c != "<p><b>bo<em>n</em>soir</b>, ça <b><em>va</em> bien</b> ?</p>" {
		t.Errorf("failed, got %s", c)
		t.Logf("lxs: %s\ntree: %s", lxs, tree)
	}
}
