package markdown

import (
	"html/template"

	"git.anhgelus.world/anhgelus/small-web/dom"
)

type astQuote struct {
	quote  []*astParagraph
	source []*astParagraph
}

func (a *astQuote) Eval(opt *Option) (template.HTML, *ParseError) {
	quoteContent, err := evalBlock(a.quote, opt)
	if err != nil {
		return "", err
	}
	blockquote := dom.NewLiteralContentElement(
		"blockquote",
		template.HTML(quoteContent),
	)
	source, err := evalBlock(a.source, opt)
	if err != nil {
		return "", err
	}
	quote := dom.NewContentElement("div", make([]dom.Element, 0))
	quote.ClassList().Add("quote")
	quote.Contents = append(quote.Contents, blockquote)
	if len(source) > 0 {
		quote.Contents = append(quote.Contents, dom.NewParagraph(source))
	}
	return quote.Render(), nil
}

func quote(lxs *lexers) (block, *ParseError) {
	tree := new(astQuote)
	n := 0
	quoteContinue := true
	source := false
	for lxs.Next() && n < 2 {
		current := lxs.Current()
		n = 0
		switch current.Type {
		case lexerBreak:
			n = len(current.Value)
			quoteContinue = false
		case lexerQuote:
			if source {
				// because the code did not use it
				lxs.Before()
				return tree, nil
			}
			quoteContinue = true
		case lexerCallout:
			if len(tree.quote) == 0 {
				return callout(lxs)
			}
			fallthrough
		case lexerLiteral, lexerModifier, lexerCode, lexerExternal:
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
