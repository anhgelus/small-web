package markdown

import "testing"

func TestExternal(t *testing.T) {
	t.Run("link", func(t *testing.T) {
		t.Run("simple", test("[content](href)", `<p><a href="href">content</a></p>`))
		t.Run("combo", test("Hey, [link](href)", `<p>Hey, <a href="href">link</a></p>`))
	})
	t.Run("image", func(t *testing.T) {
		t.Run("simple", test("![image alt](image src)", `<figure><img alt="image alt" src="image src"></figure>`))
		t.Run("combo", test(`
Avant la source
![image alt](image src)
source 1
source 2

Hors de la source
`, `<p>Avant la source</p><figure><img alt="image alt" src="image src"><figcaption>source 1 source 2</figcaption></figure><p>Hors de la source</p>`))
	})
}
