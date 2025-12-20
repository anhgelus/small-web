package markdown

import "testing"

func TestModifier(t *testing.T) {
	t.Run("modifiers", func(t *testing.T) {
		t.Run("combo", test(`**bo*n*soir**, ça ***va* bien** ?`, `<p><b>bo<em>n</em>soir</b>, ça <b><em>va</em> bien</b> ?</p>`))
	})
}
