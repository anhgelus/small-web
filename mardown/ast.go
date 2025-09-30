package mardown

import (
	"errors"
	"fmt"
)

var ErrUnkownLexType = errors.New("unkown lex type")

type block interface {
	Eval() error
}

type tree struct {
	blocks []block
}

func (t *tree) Eval() error {
	return nil
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
	case lexerBreak:
	case lexerExternal:
	case lexerModifier:
	case lexerCode:
	case lexerEscape:
	case lexerQuote:
	case lexerLiteral:
	default:
		err = errors.Join(ErrUnkownLexType, fmt.Errorf("type received: %s", lxs.Current().Type))
	}
	return b, err
}
