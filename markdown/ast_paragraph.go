package markdown

import (
	"errors"
	"html/template"
	"strings"

	"git.anhgelus.world/anhgelus/small-web/dom"
)

var (
	ErrInvalidParagraph = errors.New("invalid paragraph")
)

type astParagraph struct {
	content []block
	oneLine bool
}

func (a *astParagraph) Eval(opt *Option) (template.HTML, *ParseError) {
	var content template.HTML
	for _, c := range a.content {
		ct, err := c.Eval(opt)
		if err != nil {
			return "", err
		}
		content += ct
	}
	if a.oneLine {
		return content, nil
	}
	return dom.NewParagraph(
		template.HTML(strings.TrimSpace(string(content))),
	).Render(), nil
}

type astBreak struct{}

func (a astBreak) Eval(opt *Option) (template.HTML, *ParseError) {
	if opt.Poem {
		return dom.NewVoidElement("br").Render(), nil
	}
	return " ", nil
}

func paragraph(lxs *lexers, oneLine bool) (*astParagraph, *ParseError) {
	tree := new(astParagraph)
	tree.oneLine = oneLine
	maxBreak := 2
	if oneLine {
		maxBreak = 1
	}
	n := 0
	lxs.Before() // because we do not use it before the next
	for lxs.Next() && n < maxBreak {
		var err *ParseError
		var b block
		switch lxs.Current().Type {
		case lexerBreak:
			n += len(lxs.Current().Value)
		case lexerQuote, lexerList:
			if n > 0 {
				lxs.Before() // because we did not use it
				return tree, nil
			}
			b = astLiteral(lxs.Current().Value)
		case lexerLiteral, lexerHeading:
			b = astLiteral(lxs.Current().Value)
		case lexerReplace:
			b = astReplacer(lxs.Current().Value)
		case lexerModifier:
			var e error
			b, e = modifier(lxs)
			if e != nil {
				err = &ParseError{lxs: *lxs, internal: e}
			}
		case lexerExternal:
			if n > 0 && lxs.Current().Value == "![" {
				lxs.Before() // because we did not use it
				return tree, nil
			}
			if lxs.Current().Value != "[" {
				b = astLiteral(lxs.Current().Value)
			} else {
				b, err = external(lxs)
			}
		case lexerCode:
			if len(lxs.Current().Value) > 1 {
				err = &ParseError{lxs: *lxs, internal: ErrInvalidCodeBlockPosition}
			} else {
				b, err = code(lxs)
			}
		}

		if err != nil {
			return nil, err
		}

		if b != nil {
			if n > 0 && len(tree.content) != 0 {
				tree.content = append(tree.content, astBreak{})
			}
			tree.content = append(tree.content, b)
		}

		if lxs.Current().Type != lexerBreak {
			n = 0
		}
	}
	lxs.Before() // because we never handle the last item
	return tree, nil
}

type astLiteral string

func (a astLiteral) Eval(_ *Option) (template.HTML, *ParseError) {
	return template.HTML(template.HTMLEscapeString(string(a))), nil
}

type astReplacer string

func (a astReplacer) Eval(opt *Option) (template.HTML, *ParseError) {
	return template.HTML(opt.Replaces[[]rune(a)[0]]), nil
}
