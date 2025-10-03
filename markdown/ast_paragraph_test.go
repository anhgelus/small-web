package markdown

import "testing"

func TestParagraph(t *testing.T) {
	content := "bonsoir"
	lxs := lex(content)
	tree, err := ast(lxs)
	if err != nil {
		t.Fatal(err)
	}
	c, err := tree.Eval(nil)
	if err != nil {
		t.Fatal(err)
	}
	if c != "<p>bonsoir</p>" {
		t.Errorf("failed, got %s", c)
		t.Logf("lxs: %s\ntree: %s", lxs, tree)
	}
}
