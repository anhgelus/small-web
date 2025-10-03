package markdown

import "testing"

func TestExternal(t *testing.T) {
	lxs := lex("[content](href)")
	tree, err := ast(lxs)
	if err != nil {
		t.Fatal(err)
	}
	got, err := tree.Eval(nil)
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
	got, err = tree.Eval(nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `<figure><img alt="image alt" src="image src"></figure>` {
		t.Errorf("invalid value, got %s", got)
	}

	lxs = lex(`
![image alt](image src)
source 1
source 2

Hors de la source
`)
	tree, err = ast(lxs)
	if err != nil {
		t.Fatal(err)
	}
	got, err = tree.Eval(nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `<figure><img alt="image alt" src="image src"><figcaption>source 1 source 2</figcaption></figure><p>Hors de la source</p>` {
		t.Errorf("invalid value, got %s", got)
	}

}
