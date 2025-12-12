package markdown

import (
	"fmt"
	"html/template"
	"regexp"
)

var ExternalLink = regexp.MustCompile(`https?://`)

type astLink struct {
	content  block
	href     block
	addSpace bool
}

func (a *astLink) Eval(opt *Option) (template.HTML, *ParseError) {
	content, err := a.content.Eval(opt)
	if err != nil {
		return "", err
	}
	href, err := a.href.Eval(opt)
	if err != nil {
		return "", err
	}
	rr := opt.RenderLink(string(content), string(href))
	if a.addSpace {
		return " " + rr, nil
	}
	return rr, nil
}

func RenderLink(content, href string) template.HTML {
	if !ExternalLink.Match([]byte(href)) {
		return template.HTML(fmt.Sprintf(`<a href="%s">%s</a>`, href, content))
	}
	return template.HTML(fmt.Sprintf(`<a href="%s" target="_blank" rel="noreferer">%s</a>`, href, content))
}

type astImage struct {
	alt    block
	src    block
	source []*astParagraph
}

func (a *astImage) Eval(opt *Option) (template.HTML, *ParseError) {
	alt, err := a.alt.Eval(opt)
	if err != nil {
		return "", err
	}
	src, err := a.src.Eval(opt)
	if err != nil {
		return "", err
	}
	src = template.HTML(opt.ImageSource(string(src)))
	if a.source == nil {
		return template.HTML(fmt.Sprintf(`<figure><img alt="%s" src="%s"></figure>`, alt, src)), nil
	}
	var s template.HTML
	for _, c := range a.source {
		ct, err := c.Eval(opt)
		if err != nil {
			return "", err
		}
		s += ct + " "
	}
	s = s[:len(s)-1]
	return template.HTML(fmt.Sprintf(`<figure><img alt="%s" src="%s"><figcaption>%s</figcaption></figure>`, alt, src, s)), nil
}

func external(lxs *lexers) (block, *ParseError) {
	tp := lxs.Current().Value
	if !lxs.Next() {
		return astLiteral(tp), nil
	}
	lxs.Before() // because we called Next
	addSpace := lxs.Before() && lxs.Current().Type == lexerBreak
	lxs.Next() // because we called Before
	var b block
	var err *ParseError
	switch tp {
	case "![":
		b, err = image(lxs)
	case "[":
		b, err = link(lxs, addSpace)
	default:
		b = astLiteral(tp)
	}
	return b, err
}

func link(lxs *lexers, addSpace bool) (block, *ParseError) {
	lk := new(astLink)
	lk.addSpace = addSpace
	start := lxs.current
	content, href, _, ok := parseExternal(lxs, false)
	if !ok {
		return reset(lxs, start), nil
	}
	lk.content = astLiteral(content)
	lk.href = astLiteral(href)
	return lk, nil
}

func image(lxs *lexers) (block, *ParseError) {
	img := new(astImage)
	start := lxs.current
	alt, src, source, ok := parseExternal(lxs, true)
	if !ok {
		return reset(lxs, start), nil
	}
	img.alt = astLiteral(alt)
	img.src = astLiteral(src)
	img.source = source
	return img, nil
}

func parseExternal(lxs *lexers, withSource bool) (string, string, []*astParagraph, bool) {
	next := false
	var s string
	var first string
	var end string
	var ps []*astParagraph
	n := 0
	fn := func() bool {
		p, err := paragraph(lxs, true)
		if err != nil {
			return false
		}
		ps = append(ps, p)
		n = 0
		return true
	}
	for lxs.Next() && n < 2 {
		switch lxs.Current().Type {
		case lexerBreak:
			if !withSource {
				return "", "", nil, false
			}
			n += len(lxs.Current().Value)
			if n < 2 && first != "" && end != "" {
				if !lxs.Next() {
					return first, end, ps, true
				}
				ok := fn()
				if !ok {
					return "", "", nil, false
				}
			}
			lxs.Before() // because we must parse lexerBreak and the next call must parse the next value
		case lexerExternal:
			if first != "" && end != "" {
				return "", "", nil, false
			}
			if n > 0 && (first == "" || end == "") {
				return "", "", nil, false
			}
			n = 0
			if !next {
				if lxs.Current().Value != "](" || !lxs.Next() {
					return "", "", nil, false
				}
				lxs.Before() // because we called Next
				first = s
				s = ""
				next = true
			} else {
				if lxs.Current().Value != ")" {
					return "", "", nil, false
				}
				if !withSource {
					return first, s, nil, true
				}
				end = s
				s = ""
				if lxs.Next() && lxs.Current().Type != lexerBreak {
					return "", "", nil, false
				}
				lxs.Before() // because we called Next
			}
		default:
			if ps != nil {
				return "", "", nil, false
			}
			n = 0
			s += lxs.Current().Value
		}
	}
	if !withSource {
		return "", "", nil, false
	}
	return first, end, ps, true
}

func reset(lxs *lexers, start int) block {
	lxs.current = start
	return astLiteral(lxs.Current().Value)
}
