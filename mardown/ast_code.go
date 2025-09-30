package mardown

import (
	"errors"
	"fmt"
	"html/template"
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

func (a *astCode) Eval() (template.HTML, error) {
	switch a.codeType {
	case codeOneLine:
		return template.HTML(fmt.Sprintf("<code>%s</code>", template.HTMLEscapeString(a.content))), nil
	case codeMultiLine:
		return template.HTML(fmt.Sprintf("<pre><code>%s</code></pre>", template.HTMLEscapeString(a.content))), nil
	default:
		return "", ErrUnknownCodeType
	}
}

func code(lxs *lexers) (*astCode, error) {
	tree := new(astCode)
	current := lxs.Current().Value
	if len(current) == 3 {
		tree.codeType = codeMultiLine
	} else if len(current) == 1 {
		tree.codeType = codeOneLine
	} else {
		return nil, ErrInvalidCodeFormat
	}
	started := false
	for lxs.Next() && lxs.Current().Value != current {
		if lxs.Current().Type == lexerBreak {
			if tree.codeType == codeOneLine {
				return nil, ErrInvalidCodeFormat
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
