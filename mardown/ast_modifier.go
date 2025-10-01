package mardown

import (
	"encoding/json"
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
)

type astModifier struct {
	symbols string
	tag     modifierTag
	content []block
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

func (a *astModifier) String() string {
	content := "["
	for _, c := range a.content {
		content += "\n\t"
		if v, ok := c.(fmt.Stringer); ok {
			content += v.String()
		} else {
			b, _ := json.MarshalIndent(a.content, "\t", "  ")
			content += string(b)
		}
		content += ",\n\t"
	}
	content += "]"
	return fmt.Sprintf("modifier{sym: %s, tag: %s, content: %s\n}", a.symbols, a.tag, content)
}

func modifier(lxs *lexers) (*astModifier, error) {
	current := lxs.Current().Value
	mod, err := modifierDetect(current)
	if err != nil {
		return nil, err
	}
	var s string
	for lxs.Next() {
		switch lxs.Current().Type {
		case lexerLiteral, lexerHeader, lexerList:
			s += lxs.Current().Value
		case lexerModifier:
			if lxs.Current().Value == mod.symbols {
				mod.content = append(mod.content, astLiteral(s))
				return mod, nil
			}
			if len(s) != 0 {
				mod.content = append(mod.content, astLiteral(s))
				s = ""
			}
			c, err := modifier(lxs)
			if err != nil {
				return nil, err
			}
			mod.content = append(mod.content, c)
		case lexerBreak:
			lxs.Before() // because we did not use it
			if len(s) != 0 {
				return nil, ErrInvalidModifier
			}
			return mod, nil
		case lexerExternal:
			if lxs.Current().Value == "!" {
				s += lxs.Current().Value
			}
		default:
			return nil, ErrInvalidTypeInModifier
		}
	}
	if len(s) != 0 {
		return nil, ErrInvalidModifier
	}
	return mod, nil
}

func modifierDetect(val string) (*astModifier, error) {
	mod := new(astModifier)
	switch len(val) {
	case 1:
		mod.symbols = val
		mod.tag = emTag
		return mod, nil
	case 2:
		mod.symbols = val
		mod.tag = boldTag
		return mod, nil
	default:
		return nil, ErrInvalidModifier
	}
}
