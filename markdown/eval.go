package markdown

import "html/template"

func Parse(s string) (template.HTML, *ParseError) {
	lxs := lex(s)
	tree, err := ast(lxs)
	if err != nil {
		return "", err
	}
	return tree.Eval()
}

func ParseBytes(b []byte) (template.HTML, *ParseError) {
	return Parse(string(b))
}
