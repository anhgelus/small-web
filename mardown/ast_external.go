package mardown

import (
	"fmt"
	"html/template"
)

type astLink struct {
	content block
	href    block
}

func (a *astLink) Eval() (template.HTML, error) {
	content, err := a.content.Eval()
	if err != nil {
		return "", err
	}
	href, err := a.href.Eval()
	if err != nil {
		return "", err
	}
	return template.HTML(fmt.Sprintf(`<a href="%s">%s</a>`, href, content)), nil
}

type astImage struct {
	alt    block
	src    block
	source *astParagraph
}

func (a *astImage) Eval() (template.HTML, error) {
	alt, err := a.alt.Eval()
	if err != nil {
		return "", err
	}
	src, err := a.src.Eval()
	if err != nil {
		return "", err
	}
	if a.source == nil {
		return template.HTML(fmt.Sprintf(`<figure><img alt="%s" src="%s"></figure>`, alt, src)), nil
	}
	source, err := a.source.Eval()
	if err != nil {
		return "", err
	}
	return template.HTML(fmt.Sprintf(`<figure><img alt="%s" src="%s"><figcaption>%s</figcaption></figure>`, alt, src, source)), nil
}

func external(lxs *lexers) (block, error) {
	tp := lxs.Current().Value
	if !lxs.Next() {
		return astLiteral(tp), nil
	}
	lxs.Before() // because we call Next
	var b block
	var err error
	switch tp {
	case "![":
		b, err = image(lxs)
	case "[":
		b, err = link(lxs)
	default:
		b = astLiteral(tp)
	}
	return b, err
}

func link(lxs *lexers) (block, error) {
	lk := new(astLink)
	start := lxs.current
	content, href, _, ok := parseExternal(lxs, 1)
	if !ok {
		return reset(lxs, start), nil
	}
	lk.content = astLiteral(content)
	lk.href = astLiteral(href)
	return lk, nil
}

func image(lxs *lexers) (block, error) {
	img := new(astImage)
	start := lxs.current
	alt, src, _, ok := parseExternal(lxs, 2)
	if !ok {
		return reset(lxs, start), nil
	}
	img.alt = astLiteral(alt)
	img.src = astLiteral(src)
	//img.source = astLiteral(source)
	return img, nil
}

func parseExternal(lxs *lexers, maxBreak int) (string, string, string, bool) {
	next := false
	var s string
	var first string
	var end string
	n := 0
	for lxs.Next() && n < maxBreak {
		switch lxs.Current().Type {
		case lexerBreak:
			n++
		case lexerExternal:
			if n > 0 && (first == "" || end == "") {
				return "", "", "", false
			}
			n = 0
			if !next {
				if lxs.Current().Value != "](" || !lxs.Next() {
					return "", "", "", false
				}
				lxs.Before() // because we called Next
				first = s
				s = ""
				next = true
			} else {
				if lxs.Current().Value != ")" {
					return "", "", "", false
				}
				if maxBreak == 1 {
					return first, s, "", true
				}
				end = s
				s = ""
			}
		default:
			n = 0
			s += lxs.Current().Value
		}
	}
	if maxBreak == 1 {
		return "", "", "", false
	}
	return first, end, s, true
}

func reset(lxs *lexers, start int) block {
	lxs.current = start
	return astLiteral(lxs.Current().Value)
}
