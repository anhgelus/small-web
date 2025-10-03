package markdown

import "html/template"

type Option struct {
	ImageSource func(string) string
}

func Parse(s string, opt *Option) (template.HTML, *ParseError) {
	lxs := lex(s)
	tree, err := ast(lxs)
	if err != nil {
		return "", err
	}
	if opt == nil {
		opt = new(Option)
	}
	if opt.ImageSource == nil {
		opt.ImageSource = func(s string) string { return s }
	}
	return tree.Eval(opt)
}

func ParseBytes(b []byte, opt *Option) (template.HTML, *ParseError) {
	return Parse(string(b), opt)
}
