package mardown

import (
	"fmt"
	"html/template"
	"regexp"
)

var regexOrdered = regexp.MustCompile(`\d+\.`)

type listType string

const (
	listUnordered listType = "ul"
	listOrdered   listType = "ol"
)

type astList struct {
	tag     listType
	content []*astParagraph
}

func (a *astList) Eval() (template.HTML, error) {
	var content template.HTML
	for _, c := range a.content {
		ct, err := c.Eval()
		if err != nil {
			return "", err
		}
		content += template.HTML(fmt.Sprintf("<li>%s</li>", trimSpace(ct)))
	}
	return template.HTML(fmt.Sprintf("<%s>%s</%s>", a.tag, content, a.tag)), nil
}

func list(lxs *lexers) (block, error) {
	tree := new(astList)
	tree.tag = detectListType(lxs.Current().Value)
	if len(tree.tag) == 0 {
		return paragraph(lxs, false)
	}
	n := 0
	for lxs.Next() && n < 2 {
		switch lxs.Current().Type {
		case lexerBreak:
			n++
		case lexerList:
			n = 0
			tp := detectListType(lxs.Current().Value)
			if tp != tree.tag {
				lxs.Before() // because we dit not use it
				return tree, nil
			}
		default:
			n = 0
			c, err := paragraph(lxs, true)
			if err != nil {
				return nil, err
			}
			tree.content = append(tree.content, c)
		}
	}
	return tree, nil
}

func detectListType(val string) listType {
	first := []rune(val)[0]
	if first == '-' || first == '*' {
		if len(val) > 1 {
			return ""
		}
		return listUnordered
	}
	if !regexOrdered.MatchString(val) {
		return ""
	}
	return listOrdered
}
