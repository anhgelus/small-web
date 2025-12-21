package markdown

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
qui a elle aussi une source :D

- Ceci est une liste
- pas ordonnée
1. et maintenant
2. elle l'est
- hehe

![Ceci est ma pfp :3](https://cdn.anhgelus.world/pfp.jpg)
[Ma pfp](https://now.anhgelus.world/) hehe :D
Elle est **magnifique**, n'est-ce pas ?
`

var parsed = `
<h1>Je suis un titre</h1>
<p>Avec une description classique, sur plusieurs lignes !</p>
<p>Et je peux mettre du texte en <b>gras</b>, en <em>italique</em> et les <b><em>deux en même temps</em></b> !</p>
<div class="quote"><blockquote>Je suis une magnifique citation sur plusieurs lignes</blockquote><p>avec une source</p></div>
<div class="quote"><blockquote>qui recommence après !</blockquote><p>qui a elle aussi une source :D</p></div>
<ul><li>Ceci est une liste</li><li>pas ordonnée</li></ul>
<ol><li>et maintenant</li><li>elle l&#39;est</li></ol>
<ul><li>hehe</li></ul>
<figure>
<img alt="Ceci est ma pfp :3" src="https://cdn.anhgelus.world/pfp.jpg" />
<figcaption><a href="https://now.anhgelus.world/" target="_blank" rel="noreferer">Ma pfp</a> hehe :D Elle est <b>magnifique</b>, n&#39;est-ce pas ?</figcaption>
</figure>
`

var parsedPoem = `
<h1>Je suis un titre</h1>
<p>Avec une description classique,<br />sur plusieurs lignes !</p>
<p>Et je peux mettre du texte en <b>gras</b>,<br />en <em>italique</em> et les <b><em>deux en même temps</em></b> !</p>
<div class="quote"><blockquote>Je suis une magnifique citation sur plusieurs lignes</blockquote><p>avec une source</p></div>
<div class="quote"><blockquote>qui recommence après !</blockquote><p>qui a elle aussi une source :D</p></div>
<ul><li>Ceci est une liste</li><li>pas ordonnée</li></ul>
<ol><li>et maintenant</li><li>elle l&#39;est</li></ol>
<ul><li>hehe</li></ul>
<figure>
<img alt="Ceci est ma pfp :3" src="https://cdn.anhgelus.world/pfp.jpg" />
<figcaption><a href="https://now.anhgelus.world/" target="_blank" rel="noreferer">Ma pfp</a> hehe :D Elle est <b>magnifique</b>, n&#39;est-ce pas ?</figcaption>
</figure>
`

func test(input, expected string) func(*testing.T) {
	return testWithOptions(nil, input, expected)
}

func testWithOptions(opt *Option, input, expected string) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		got, err := Parse(input, opt)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != expected {
			t.Errorf("invalid value, got %s", got)
		}
	}
}

func TestAst(t *testing.T) {
	t.Run("ast", func(t *testing.T) {
		t.Run("complete", test(raw, strings.ReplaceAll(parsed, "\n", "")))
		t.Run("poem", testWithOptions(&Option{Poem: true}, raw, strings.ReplaceAll(parsedPoem, "\n", "")))
	})
}
