package mardown

import (
	"errors"
	"fmt"
	"html/template"
)

var ErrInvalidHeader = errors.New("invalid header")

type astHeader struct {
	level   uint
	content []block
}

func (a *astHeader) Eval() (template.HTML, error) {
	if a.level > 6 {
		return "", ErrInvalidHeader
	}
	var content template.HTML
	for _, b := range a.content {
		c, err := b.Eval()
		if err != nil {
			return "", err
		}
		content += c
	}
	return template.HTML(fmt.Sprintf("<h%d>%s<h%d>", a.level, content, a.level)), nil
}

func header(lxs lexers) (block, error) {
	b := &astHeader{level: uint(len(lxs.Current().Value))}
	for lxs.Next() && lxs.Current().Type != lexerBreak {
		bl, err := getBlock(lxs)
		if err != nil {
			return nil, err
		}
		// if this is a header, just consider it as literal #
		if h, ok := bl.(*astHeader); ok {
			var s string
			for range h.level {
				s += "#"
			}
			b.content = append(b.content, astLiteral(s))
			b.content = append(b.content, h.content...)
		}
		b.content = append(b.content, bl)
	}
	return b, nil
}
