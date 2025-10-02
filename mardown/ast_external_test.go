package mardown

import "testing"

func TestExternal(t *testing.T) {
	lxs := lex("[content](href)")
	tree, err := ast(lxs)
	if err != nil {
		t.Fatal(err)
	}
	got, err := tree.Eval()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `<p><a href="href">content</a></p>` {
		t.Errorf("invalid value, got %s", got)
	}

	lxs = lex("![image alt](image src)")
	tree, err = ast(lxs)
	if err != nil {
		t.Fatal(err)
	}
	got, err = tree.Eval()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `<figure><img alt="image alt" src="image src"></figure>` {
		t.Errorf("invalid value, got %s", got)
	}
}
