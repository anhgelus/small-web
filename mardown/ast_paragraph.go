package mardown

import (
	"errors"
	"fmt"
	"html/template"
)

var (
	ErrInvalidParagraph         = errors.New("invalid paragraph")
	ErrInvalidCodeBlockPosition = errors.Join(ErrInvalidParagraph, errors.New("invalid code block position"))
)

type astParagraph struct {
	content []block
	oneLine bool
}

func (a *astParagraph) Eval() (template.HTML, error) {
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
	return template.HTML(fmt.Sprintf("<p>%s</p>", content)), nil
}

func paragraph(lxs lexers, oneLine bool) (*astParagraph, error) {
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
			n++
		case lexerLiteral, lexerHeader:
			tree.content = append(tree.content, astLiteral(lxs.Current().Value))
		case lexerModifier:
			mod, err := modifier(lxs)
			if err != nil {
				return nil, err
			}
			tree.content = append(tree.content, mod)
		case lexerQuote:
			//TODO: handle
		case lexerEscape:
			//TODO: handle
		case lexerExternal:
			//TODO: handle
		case lexerCode:
			if len(lxs.Current().Value) == 3 {
				if n == 0 {
					return nil, ErrInvalidCodeBlockPosition
				}
				lxs.current-- // because we do not use it before the next
				return tree, nil
			}
			mod, err := modifier(lxs)
			if err != nil {
				return nil, err
			}
			tree.content = append(tree.content, mod)
		}
	}
	return tree, nil
}

type astLiteral string

func (a astLiteral) Eval() (template.HTML, error) {
	return template.HTML(template.HTMLEscapeString(string(a))), nil
}
