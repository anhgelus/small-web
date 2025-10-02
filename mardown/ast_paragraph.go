package mardown

import (
	"errors"
	"fmt"
	"html/template"
)

var (
	ErrInvalidParagraph = errors.New("invalid paragraph")
)

type astParagraph struct {
	content []block
	oneLine bool
}

func (a *astParagraph) Eval() (template.HTML, *ParseError) {
	var content template.HTML
	for _, c := range a.content {
		ct, err := c.Eval()
		if err != nil {
			return "", err
		}
		content += ct
	}
	if a.oneLine {
		return content, nil
	}
	return template.HTML(fmt.Sprintf("<p>%s</p>", trimSpace(content))), nil
}

func paragraph(lxs *lexers, oneLine bool) (*astParagraph, *ParseError) {
	tree := new(astParagraph)
	tree.oneLine = oneLine
	maxBreak := 2
	if oneLine {
		maxBreak = 1
	}
	n := 0
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
			tree.content = append(tree.content, astLiteral(lxs.Current().Value))
		case lexerLiteral, lexerHeader:
			s := lxs.Current().Value
			// replace line break by space
			if n > 0 {
				s = " " + s
			}
			n = 0
			tree.content = append(tree.content, astLiteral(s))
		case lexerModifier:
			n = 0
			mod, err := modifier(lxs)
			if err != nil {
				return nil, &ParseError{lxs: *lxs, internal: err}
			}
			tree.content = append(tree.content, mod)
		case lexerExternal:
			n = 0
			if lxs.Current().Value == "!" {
				tree.content = append(tree.content, astLiteral(lxs.Current().Value))
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
			n = 0
			b, err := code(lxs)
			if err != nil {
				return nil, err
			}
			tree.content = append(tree.content, b)
			return tree, nil
		}
	}
	if !lxs.Finished() {
		lxs.Before() // because we never handle the last item
	}
	return tree, nil
}

type astLiteral string

func (a astLiteral) Eval() (template.HTML, *ParseError) {
	return template.HTML(template.HTMLEscapeString(string(a))), nil
}
