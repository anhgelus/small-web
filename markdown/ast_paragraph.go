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

func paragraph(lxs *lexers, oneLine bool) (*astParagraph, *ParseError) {
	tree := new(astParagraph)
	tree.oneLine = oneLine
	maxBreak := 2
	if oneLine {
		maxBreak = 1
	}
	n := 0
	asLiteral := func(conv func(s string) block) {
		s := lxs.Current().Value
		// replace line break by space
		if n > 0 && len(tree.content) != 0 {
			s = " " + s
		}
		tree.content = append(tree.content, conv(s))
	}
	lxs.current-- // because we do not use it before the next
	for lxs.Next() && n < maxBreak {
		switch lxs.Current().Type {
		case lexerBreak:
			n += len(lxs.Current().Value)
		case lexerQuote, lexerList:
			if n > 0 {
				lxs.Before() // because we did not use it
				return tree, nil
			}
			asLiteral(toAstLiteral)
		case lexerLiteral, lexerHeading:
			asLiteral(toAstLiteral)
		case lexerReplace:
			asLiteral(toAstReplacer)
		case lexerModifier:
			// replace line break by space
			if n > 0 {
				tree.content = append(tree.content, astLiteral(" "))
			}
			mod, err := modifier(lxs)
			if err != nil {
				return nil, &ParseError{lxs: *lxs, internal: err}
			}
			tree.content = append(tree.content, mod)
		case lexerExternal:
			if n > 0 && lxs.Current().Value == "![" {
				lxs.Before() // because we did not use it
				return tree, nil
			}
			if lxs.Current().Value != "[" {
				asLiteral(toAstLiteral)
			} else {
				ext, err := external(lxs)
				if err != nil {
					return nil, err
				}
				tree.content = append(tree.content, ext)
			}
		case lexerCode:
			if len(lxs.Current().Value) > 1 {
				return nil, &ParseError{lxs: *lxs, internal: ErrInvalidCodeBlockPosition}
			}
			b, err := code(lxs)
			if err != nil {
				return nil, err
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

func toAstLiteral(s string) block {
	return astLiteral(s)
}

type astReplacer string

func (a astReplacer) Eval(opt *Option) (template.HTML, *ParseError) {
	return template.HTML(opt.Replaces[[]rune(a)[0]]), nil
}

func toAstReplacer(s string) block {
	return astReplacer(s)
}
