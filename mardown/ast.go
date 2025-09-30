package mardown

import (
	"errors"
	"fmt"
	"html/template"
)

var ErrUnkownLexType = errors.New("unkown lex type")

type block interface {
	Eval() (template.HTML, error)
}

type tree struct {
	blocks []block
}

func (t *tree) Eval() (template.HTML, error) {
	return "", nil
}

func ast(lxs lexers) (*tree, error) {
	tr := new(tree)
	newLine := false
	for lxs.Next() {
		b, err := getBlock(lxs, newLine)
		if err != nil {
			return nil, err
		}
		tr.blocks = append(tr.blocks, b)
		newLine = lxs.Current().Type == lexerBreak
	}
	return tr, nil
}

func getBlock(lxs lexers, newLine bool) (block, error) {
	var b block
	var err error
	switch lxs.Current().Type {
	case lexerHeader:
		if !newLine {
			b, err = paragraph(lxs, false)
		} else {
			b, err = header(lxs)
		}
	case lexerExternal:
		if newLine && lxs.Current().Value == "!" {
			//TODO: handle
		} else {
			b, err = paragraph(lxs)
		}
	case lexerQuote:
	case lexerCode:
		if newLine && len(lxs.Current().Value) == 3 {
			//TODO: handle
		} else {
			b, err = paragraph(lxs)
		}
	case lexerLiteral, lexerEscape, lexerModifier:
		b, err = paragraph(lxs)
	case lexerBreak: // do nothing
	default:
		err = errors.Join(ErrUnkownLexType, fmt.Errorf("type received: %s", lxs.Current().Type))
	}
	return b, err
}
