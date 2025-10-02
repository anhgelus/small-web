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
	super   bool
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
	if a.super {
		return content, nil
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
	return fmt.Sprintf("modifier{sym: %s, tag: %s, super: %v, content: %s\n}", a.symbols, a.tag, a.super, content)
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
			if mod.super && []rune(mod.symbols)[0] == []rune(lxs.Current().Value)[0] &&
				len(mod.symbols) >= len(lxs.Current().Value) {
				mod.symbols = mod.symbols[len(lxs.Current().Value):]
				subMod, err := modifierDetect(lxs.Current().Value)
				if err != nil {
					return nil, err
				}
				if !subMod.super {
					subMod.content = append(subMod.content, astLiteral(s))
					mod, err = modifierDetect(mod.symbols) // this trick is so cool :D
					if err != nil {
						return nil, err
					}
				} else {
					subMod, _ = modifierDetect("**")
					subEm, _ := modifierDetect("*")
					subEm.content = append(subEm.content, astLiteral(s))
					subMod.content = append(subMod.content, subEm)
				}
				s = ""
				mod.content = append(mod.content, subMod)
				if len(mod.symbols) == 0 {
					return mod, nil
				}
			} else {
				if lxs.Current().Value == mod.symbols {
					mod.content = append(mod.content, astLiteral(s))
					return mod, nil
				} else if len(s) != 0 {
					mod.content = append(mod.content, astLiteral(s))
					s = ""
				}
				c, err := modifier(lxs)
				if err != nil {
					return nil, err
				}
				mod.content = append(mod.content, c)
			}
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
	mod.symbols = val
	switch len(val) {
	case 1:
		mod.tag = emTag
	case 2:
		mod.tag = boldTag
	case 3:
		mod.super = true
	default:
		return nil, ErrInvalidModifier
	}
	return mod, nil
}
