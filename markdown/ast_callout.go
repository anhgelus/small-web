package markdown

import (
	"errors"
	"html/template"
	"strings"

	"git.anhgelus.world/anhgelus/small-web/dom"
)

var (
	ErrInvalidCallout = errors.New("invalid callout")
)

type astCallout struct {
	kind    string
	title   *astParagraph
	content []*astParagraph
}

func (a *astCallout) Eval(opt *Option) (template.HTML, *ParseError) {
	inner := dom.NewContentElement("div", make([]dom.Element, 0))
	for _, c := range a.content {
		ct, err := c.Eval(opt)
		if err != nil {
			return "", err
		}
		inner.Contents = append(inner.Contents, dom.NewParagraph(
			template.HTML(strings.TrimSpace(string(ct))),
		))
	}

	titleContent, err := a.title.Eval(opt)
	if err != nil {
		return "", err
	}
	titleContent = template.HTML(strings.TrimSpace(string(titleContent)))
	if len(titleContent) == 0 {
		titleContent = template.HTML(a.kind)
	}
	title := dom.NewLiteralContentElement("h4", titleContent)

	callout := dom.NewContentElement("div", make([]dom.Element, 0))
	callout.Contents = append(callout.Contents, title)
	if len(inner.Contents) > 0 {
		callout.Contents = append(callout.Contents, inner)
	}
	callout.SetAttribute("data-kind", a.kind)
	callout.ClassList().Add("callout")
	return callout.Render(), nil
}

func callout(lxs *lexers) (block, *ParseError) {
	callout := new(astCallout)
	if lxs.Current().Value != "[!" {
		return paragraph(lxs, false)
	}
	if !lxs.Next() {
		return nil, &ParseError{lxs: *lxs, internal: ErrInvalidCallout}
	}
	callout.kind = strings.ToLower(lxs.Current().Value)
	if !lxs.Next() || lxs.Current().Type != lexerCallout || lxs.Current().Value != "]" {
		return nil, &ParseError{lxs: *lxs, internal: ErrInvalidCallout}
	}
	var err *ParseError
	callout.title, err = paragraph(lxs, true)
	if err != nil {
		return nil, err
	}
	n := 0
	for lxs.Next() && n < 2 {
		current := lxs.Current()
		n = 0
		switch current.Type {
		case lexerBreak:
			n = len(current.Value)
		case lexerQuote:
		case lexerLiteral, lexerModifier, lexerCode, lexerExternal:
			p, err := paragraph(lxs, true)
			if err != nil {
				return nil, err
			}
			lxs.Before()
			callout.content = append(callout.content, p)
			n++
		default:
			// because the code did not use it
			lxs.Before()
			return callout, nil
		}
	}
	lxs.Before()
	return callout, nil
}
