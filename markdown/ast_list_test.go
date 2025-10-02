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
	lxs := lex(rw)
	tree, err := ast(lxs)
	if err != nil {
		t.Fatal(err)
	}
	got, err := tree.Eval()
	if err != nil {
		t.Fatal(err)
	}
	exp := strings.ReplaceAll(expected, "\n", "")
	if string(got) != exp {
		t.Errorf("invalid value, got %s", got)
		t.Logf("expected %s", exp)
	}
}
