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
	for lxs.Next() {
		b, err := getBlock(lxs)
		if err != nil {
			return nil, err
		}
		tr.blocks = append(tr.blocks, b)
	}
	return tr, nil
}

func getBlock(lxs lexers) (block, error) {
	var b block
	var err error
	switch lxs.Current().Type {
	case lexerHeader:
		b, err = header(lxs)
	case lexerExternal:
	case lexerModifier:
	case lexerCode:
	case lexerEscape:
	case lexerQuote:
	case lexerBreak:
	case lexerLiteral:
		b = astLiteral(lxs.Current().Value)
	default:
		err = errors.Join(ErrUnkownLexType, fmt.Errorf("type received: %s", lxs.Current().Type))
	}
	return b, err
}
