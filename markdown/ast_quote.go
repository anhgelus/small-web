package markdown

import (
	"fmt"
	"html/template"
	"strings"
)

type astQuote struct {
	quote  []*astParagraph
	source []*astParagraph
}

func (a *astQuote) Eval() (template.HTML, *ParseError) {
	var quote template.HTML
	for _, c := range a.quote {
		ct, err := c.Eval()
		if err != nil {
			return "", err
		}
		quote += ct
	}
	quote = template.HTML(fmt.Sprintf("<blockquote>%s</blockquote>", trimSpace(quote)))
	var source template.HTML
	for _, c := range a.source {
		ct, err := c.Eval()
		if err != nil {
			return "", err
		}
		source += ct
	}
	source = template.HTML(strings.TrimSpace(string(source)))
	if len(source) > 0 {
		return template.HTML(fmt.Sprintf(`<div class="quote">%s<p>%s</p></div>`, quote, source)), nil
	}
	return template.HTML(fmt.Sprintf(`<div class="quote">%s</div>`, quote)), nil
}

func quote(lxs *lexers) (*astQuote, *ParseError) {
	tree := new(astQuote)
	n := 0
	quoteContinue := true
	source := false
	for lxs.Next() && n < 2 {
		switch lxs.Current().Type {
		case lexerBreak:
			n += len(lxs.Current().Value)
			quoteContinue = false
		case lexerQuote:
			n = 0
			if source {
				// because the code did not use it
				lxs.Before()
				return tree, nil
			}
			quoteContinue = true
		case lexerLiteral, lexerModifier, lexerCode:
			n = 0
			if !quoteContinue {
				source = true
			}
			p, err := paragraph(lxs, true)
			if err != nil {
				return nil, err
			}

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
	return tree, nil
}
