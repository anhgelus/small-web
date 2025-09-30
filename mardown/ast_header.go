package mardown

import (
	"errors"
	"fmt"
	"html/template"
)

var ErrInvalidHeader = errors.New("invalid header")

type astHeader struct {
	level   uint
	content *astParagraph
}

func (a *astHeader) Eval() (template.HTML, error) {
	if a.level > 6 {
		return "", ErrInvalidHeader
	}
	var content template.HTML
	content, err := a.content.Eval()
	if err != nil {
		return "", err
	}
	return template.HTML(fmt.Sprintf("<h%d>%s</h%d>", a.level, content, a.level)), nil
}

func header(lxs lexers) (*astHeader, error) {
	b := &astHeader{level: uint(len(lxs.Current().Value))}
	var err error
	b.content, err = paragraph(lxs, true)
	if err != nil {
		return nil, err
	}
	return b, nil
}
