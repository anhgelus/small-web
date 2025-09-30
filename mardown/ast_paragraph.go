package mardown

import (
	"errors"
	"html/template"
)

var (
	ErrInvalidParagraph         = errors.New("invalid paragraph")
	ErrInvalidCodeBlockPosition = errors.Join(ErrInvalidParagraph, errors.New("invalid code block position"))
)

type astParagraph struct {
	content []block
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
	return content, nil
}

func paragraph(lxs lexers) (block, error) {
	tree := new(astParagraph)
	n := 0
	lxs.current-- // because we do not use it before the next
	for lxs.Next() && n < 2 {
		switch lxs.Current().Type {
		case lexerBreak:
			n++
		case lexerLiteral:
			tree.content = append(tree.content, astLiteral(lxs.Current().Value))
		case lexerModifier:
			//TODO: handle
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
			//TODO: handle
		}
	}
	return tree, nil
}

type astLiteral string

func (a astLiteral) Eval() (template.HTML, error) {
	return template.HTML(template.HTMLEscapeString(string(a))), nil
}
