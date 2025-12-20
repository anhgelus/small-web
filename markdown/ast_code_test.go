package markdown

import "testing"

func TestCode(t *testing.T) {
	t.Run("code", func(t *testing.T) {
		t.Run("mono", test("`mono`", `<p><code>mono</code></p>`))
		t.Run("combo", test("bonjour `code` !", `<p>bonjour <code>code</code> !</p>`))
		t.Run("mult-line", test("```\n"+"raw\nhehe"+"```", `<pre><code>raw
hehe</code></pre>`))
	})
}
