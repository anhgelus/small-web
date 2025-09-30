package mardown

import (
	"errors"
	"fmt"
	"html/template"
)

var (
	ErrInternalError         = errors.New("internal error")
	ErrInvalidModifier       = errors.Join(ErrInvalidParagraph, errors.New("invalid modifier organization"))
	ErrInvalidTypeInModifier = errors.Join(ErrInvalidParagraph, errors.New("invalid type in modifier"))
)

type modifierTag string

const (
	boldTag modifierTag = "b"
	emTag   modifierTag = "em"
	codeTag modifierTag = "code"
)

type astModifier struct {
	symbols string
	tag     modifierTag
	content []block
	parent  *astModifier
}

func (a *astModifier) Eval() (template.HTML, error) {
	var content template.HTML
	for _, c := range a.content {
		ct, err := c.Eval()
		if err != nil {
			return "", err
		}
		content += ct
	}
	return template.HTML(fmt.Sprintf("<%s>%s</%s>", a.tag, content, a.tag)), nil
}

func modifier(lxs *lexers) (*astModifier, error) {
	current := lxs.Current().Value
	mod := modifierDetect(current)
	modInside := mod
	if len(mod.content) > 0 {
		var ok bool
		modInside, ok = mod.content[0].(*astModifier)
		if modInside.content != nil && !ok {
			return nil, ErrInternalError
		}
		// getting the last modifier
		for len(modInside.content) > 0 {
			modInside, ok = modInside.content[0].(*astModifier)
			if modInside.content != nil && !ok {
				return nil, ErrInternalError
			}
		}
	}
	n := len(modInside.symbols)
	var s string
	for lxs.Next() {
		switch lxs.Current().Type {
		case lexerLiteral:
			s += lxs.Current().Value
		case lexerModifier:
			if len(lxs.Current().Value) < n {
				return nil, ErrInvalidModifier
			}
			if lxs.Current().Value[:n] != modInside.symbols {
				return nil, ErrInvalidModifier
			}
			modInside.content = append(modInside.content, astLiteral(s))
			s = ""
			modInside = modInside.parent
		default:
			return nil, ErrInvalidTypeInModifier
		}
	}
	return mod, nil
}

func modifierDetect(val string) *astModifier {
	mod := new(astModifier)
	if len(val) == 1 {
		mod.symbols = val
		if val == "`" {
			mod.tag = codeTag
		} else {
			mod.tag = emTag
		}
		return mod
	}
	if val[:2] == "**" || val[:2] == "__" {
		mod.symbols = val
		mod.tag = boldTag
	} else {
		mod = modifierDetect(val[:1])
		next := modifierDetect(val[1:])
		next.parent = mod
		mod.content = append(mod.content, next)
	}
	return mod
}
