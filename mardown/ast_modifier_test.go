package mardown

import "testing"

func TestModifier(t *testing.T) {
	content := `
***bon*soir**
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
	if c != "<p><b><em>bon</em>soir</b></p>" {
		t.Errorf("failed, got %s", c)
		t.Logf("lxs: %s\ntree: %s", lxs, tree)
	}
}
