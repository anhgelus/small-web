package markdown

import (
	"errors"
	"html"
	"html/template"

	"git.anhgelus.world/anhgelus/small-web/dom"
)

var (
	ErrUnknownCodeType          = errors.New("unkown code type")
	ErrInvalidCodeFormat        = errors.New("invalid code format")
	ErrInvalidCodeBlockPosition = errors.Join(ErrInvalidParagraph, errors.New("invalid code block position"))
)

type codeType uint

const (
	codeOneLine   codeType = 1
	codeMultiLine codeType = 2
)

type astCode struct {
	content  string
	before   string
	codeType codeType
}

func (a *astCode) Eval(_ *Option) (template.HTML, *ParseError) {
	content := template.HTML(html.EscapeString(a.content))
	switch a.codeType {
	case codeOneLine:
		return dom.NewLiteralContentElement("code", content).Render(), nil
	case codeMultiLine:
		code := dom.NewContentElement("code", []dom.Element{dom.NewLiteralElement(content)})
		return dom.NewContentElement("pre", []dom.Element{code}).Render(), nil
	default:
		return "", &ParseError{lxs: lexers{}, internal: ErrUnknownCodeType}
	}
}

func code(lxs *lexers) (*astCode, *ParseError) {
	tree := new(astCode)
	current := lxs.Current().Value
	if len(current) == 3 {
		tree.codeType = codeMultiLine
	} else if len(current) == 1 {
		tree.codeType = codeOneLine
	} else {
		return nil, &ParseError{lxs: *lxs, internal: ErrInvalidCodeFormat}
	}
	started := false
	for lxs.Next() && lxs.Current().Value != current {
		if lxs.Current().Type == lexerBreak {
			if tree.codeType == codeOneLine {
				return nil, &ParseError{lxs: *lxs, internal: ErrInvalidCodeFormat}
			}
			if !started {
				started = true
			}
		}
		if started || tree.codeType == codeOneLine {
			tree.content += lxs.Current().Value
		} else {
			tree.before += lxs.Current().Value
		}
	}
	return tree, nil
}
