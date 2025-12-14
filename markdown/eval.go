package markdown

import (
	"html/template"
)

type Option struct {
	ImageSource func(source string) string
	RenderLink  func(content, href string) template.HTML
	Replaces    map[rune]string
}

func Parse(s string, opt *Option) (template.HTML, *ParseError) {
	if opt == nil {
		opt = new(Option)
	}
	if opt.ImageSource == nil {
		opt.ImageSource = func(s string) string { return s }
	}
	if opt.RenderLink == nil {
		opt.RenderLink = RenderLink
	}
	if opt.Replaces == nil {
		opt.Replaces = make(map[rune]string, 0)
	}
	lxs := lex(s, opt)
	tree, err := ast(lxs)
	if err != nil {
		return "", err
	}
	return tree.Eval(opt)
}

func ParseBytes(b []byte, opt *Option) (template.HTML, *ParseError) {
	return Parse(string(b), opt)
}
