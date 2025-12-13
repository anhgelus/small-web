package markdown

import (
	"errors"
	"html/template"
	"strings"

	"git.anhgelus.world/anhgelus/small-web/dom"
)

var ErrInvalidHeader = errors.New("invalid header")

type astHeader struct {
	level   uint
	content *astParagraph
}

func (a *astHeader) Eval(opt *Option) (template.HTML, *ParseError) {
	if a.level > 6 {
		return "", &ParseError{lxs: lexers{}, internal: ErrInvalidCodeFormat}
	}
	var content template.HTML
	content, err := a.content.Eval(opt)
	if err != nil {
		return "", err
	}
	return dom.NewHeading(
		a.level,
		template.HTML(strings.TrimSpace(string(content))),
	).Render(), nil
}

func header(lxs *lexers) (*astHeader, *ParseError) {
	b := &astHeader{level: uint(len(lxs.Current().Value))}
	if !lxs.Next() {
		return nil, &ParseError{lxs: *lxs, internal: ErrInvalidHeader}
	}
	var err *ParseError
	b.content, err = paragraph(lxs, true)
	if err != nil {
		return nil, err
	}
	return b, nil
}
