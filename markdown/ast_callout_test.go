package markdown

import "testing"

func TestCallout(t *testing.T) {
	t.Run("callout", func(t *testing.T) {
		t.Run("simple", test(`
> [!NOTE]
`, `<div data-kind="note" class="callout"><h4>note</h4><div></div></div>`))
		t.Run("multiline", test(`
> [!NOTE] Hey :3
> content 1
> content 2
`, `<div data-kind="note" class="callout"><h4>Hey :3</h4><div><p>content 1</p><p>content 2</p></div></div>`))
	})
}
