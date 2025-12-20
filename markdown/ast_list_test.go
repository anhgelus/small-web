package markdown

import (
	"strings"
	"testing"
)

var rw = `
- item A
- item B
* item C

1. item 1
2. item 2 
`

var expected = `
<ul>
<li>item A</li>
<li>item B</li>
<li>item C</li>
</ul>
<ol>
<li>item 1</li>
<li>item 2</li>
</ol>
`

func TestList(t *testing.T) {
	t.Run("lists", func(t *testing.T) {
		t.Run("combo", test(rw, strings.ReplaceAll(expected, "\n", "")))
	})
}
