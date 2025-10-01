package mardown

import (
	"strings"
	"testing"
)

var raw = `
# Je suis un titre
Avec une description classique,
sur plusieurs lignes !

Et je peux mettre du texte en **gras**,
en *italique* et les **_deux en même temps_** !

> Je suis une magnifique citation
> sur plusieurs lignes
avec une source
> qui recommence après !

- Ceci est une liste
- pas ordonnée
1. et maintenant
2. elle l'est
- hehe
`

var parsed = `
<h1>Je suis un titre</h1>
<p>Avec une description classique, sur plusieurs lignes !</p>
<p>Et je peux mettre du texte en <b>gras</b>, en <em>italique</em> et les <b><em>deux en même temps</em></b> !</p>
<div class="quote"><blockquote>Je suis une magnifique citation sur plusieurs lignes</blockquote><p>avec une source</p></div>
<div class="quote"><blockquote>qui recommence après !</blockquote></div>
<ul><li>Ceci est une liste</li><li>pas ordonnée</li></ul>
<ol><li>et maintenant</li><li>elle l&#39;est</li></ol>
<ul><li>hehe</li></ul>
`

func TestAst(t *testing.T) {
	lxs := lex(raw)
	tree, err := ast(lxs)
	if err != nil {
		t.Fatal(err)
	}
	res, err := tree.Eval()
	if err != nil {
		t.Fatal(err)
	}
	wanted := strings.ReplaceAll(parsed, "\n", "")
	if string(res) != wanted {
		t.Errorf("invalid string, got\n%s", res)
		t.Logf("wanted\n%s", wanted)
	}
}
