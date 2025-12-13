package markdown

import (
	"html/template"
	"strings"

	"git.anhgelus.world/anhgelus/small-web/dom"
)

type astQuote struct {
	quote  []*astParagraph
	source []*astParagraph
}

func (a *astQuote) Eval(opt *Option) (template.HTML, *ParseError) {
	var quoteContent template.HTML
	for _, c := range a.quote {
		ct, err := c.Eval(opt)
		if err != nil {
			return "", err
		}
		quoteContent += ct
	}
	blockquote := dom.NewLiteralContentElement(
		"blockquote",
		template.HTML(strings.TrimSpace(string(quoteContent))),
	)
	var source template.HTML
	for _, c := range a.source {
		ct, err := c.Eval(opt)
		if err != nil {
			return "", err
		}
		source += ct
	}
	source = template.HTML(strings.TrimSpace(string(source)))
	quote := dom.NewContentElement("div", make([]dom.Element, 0))
	quote.ClassList().Add("quote")
	quote.Contents = append(quote.Contents, blockquote)
	if len(source) > 0 {
		quote.Contents = append(quote.Contents, dom.NewParagraph(source))
	}
	return quote.Render(), nil
}

func quote(lxs *lexers) (*astQuote, *ParseError) {
	tree := new(astQuote)
	n := 0
	quoteContinue := true
	source := false
	for lxs.Next() && n < 2 {
		switch lxs.Current().Type {
		case lexerBreak:
			n = len(lxs.Current().Value)
			quoteContinue = false
		case lexerQuote:
			n = 0
			if source {
				// because the code did not use it
				lxs.Before()
				return tree, nil
			}
			quoteContinue = true
		case lexerLiteral, lexerModifier, lexerCode, lexerExternal:
			n = 0
			if !quoteContinue {
				source = true
			}
			p, err := paragraph(lxs, true)
			if err != nil {
				return nil, err
			}
			lxs.Before() // because we must parse the line break

			if !source {
				tree.quote = append(tree.quote, p)
			} else {
				tree.source = append(tree.source, p)
			}
			n++
			quoteContinue = false
		default:
			// because the code did not use it
			lxs.Before()
			return tree, nil
		}
	}
	lxs.Before() // because the code did not use it
	return tree, nil
}
