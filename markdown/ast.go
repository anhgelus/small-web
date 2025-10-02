package markdown

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"strings"
)

var ErrUnkownLexType = errors.New("unkown lex type")

type block interface {
	Eval() (template.HTML, *ParseError)
}

type tree struct {
	blocks []block
}

func (t *tree) Eval() (template.HTML, *ParseError) {
	var content template.HTML
	for _, c := range t.blocks {
		ct, err := c.Eval()
		if err != nil {
			return "", err
		}
		content += ct
	}
	return content, nil
}

func (t *tree) String() string {
	b, _ := json.MarshalIndent(t, "", "  ")
	return string(b)
}

func ast(lxs *lexers) (*tree, *ParseError) {
	tr := new(tree)
	newLine := true
	for lxs.Next() {
		b, err := getBlock(lxs, newLine)
		if err != nil {
			return nil, err
		}
		if b != nil {
			tr.blocks = append(tr.blocks, b)
		}
		if !lxs.Finished() {
			newLine = lxs.Current().Type == lexerBreak
		}
	}
	return tr, nil
}

func getBlock(lxs *lexers, newLine bool) (block, *ParseError) {
	var b block
	var err *ParseError
	switch lxs.Current().Type {
	case lexerHeader:
		if !newLine {
			b, err = paragraph(lxs, false)
		} else {
			b, err = header(lxs)
		}
	case lexerExternal:
		if newLine && lxs.Current().Value == "![" {
			b, err = external(lxs)
		} else {
			b, err = paragraph(lxs, false)
		}
	case lexerQuote:
		if newLine {
			b, err = quote(lxs)
		} else {
			b, err = paragraph(lxs, false)
		}
	case lexerList:
		if newLine {
			b, err = list(lxs)
		} else {
			b, err = paragraph(lxs, false)
		}
	case lexerCode:
		if !newLine && len(lxs.Current().Value) == 3 {
			return nil, &ParseError{lxs: *lxs, internal: ErrInvalidCodeBlockPosition}
		}
		if len(lxs.Current().Value) == 1 {
			b, err = paragraph(lxs, false)
		} else {
			b, err = code(lxs)
		}
	case lexerLiteral, lexerModifier:
		b, err = paragraph(lxs, false)
	case lexerBreak: // do nothing
	default:
		err = &ParseError{
			lxs:      *lxs,
			internal: errors.Join(ErrUnkownLexType, fmt.Errorf("type received: %s", lxs.Current().Type)),
		}
	}
	return b, err
}

func trimSpace(s template.HTML) template.HTML {
	return template.HTML(strings.TrimSpace(string(s)))
}
