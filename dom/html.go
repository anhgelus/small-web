package dom

import (
	"fmt"
	"html/template"
)

func render(tag string, attributes map[string]string, endSlash bool) template.HTML {
	base := fmt.Sprintf(`<%s`, tag)
	for k, v := range attributes {
		base += fmt.Sprintf(` %s="%s"`, k, v)
	}
	if !endSlash {
		return template.HTML(base + `>`)
	}
	return template.HTML(base + ` />`)
}

type Element interface {
	Render() template.HTML
	HasAttribute(string) bool
	SetAttribute(string, string) Element
	RemoveAttribute(string) Element
	ClassList() ClassList
}

type LiteralElement template.HTML

func (e LiteralElement) Render() template.HTML {
	return template.HTML(e)
}

func (LiteralElement) HasAttribute(string) bool {
	return false
}

func (e LiteralElement) SetAttribute(string, string) Element {
	return e
}

func (e LiteralElement) RemoveAttribute(string) Element {
	return e
}

func (e LiteralElement) ClassList() ClassList {
	return nil
}

func NewLiteralElement(s template.HTML) LiteralElement {
	return LiteralElement(s)
}

type VoidElement struct {
	Tag        string
	attributes map[string]string
	cl         ClassList
}

func (e VoidElement) Render() template.HTML {
	e.cl.set(e)
	return render(e.Tag, e.attributes, true)
}

func (e VoidElement) HasAttribute(k string) bool {
	_, ok := e.attributes[k]
	return ok
}

func (e VoidElement) SetAttribute(k, v string) Element {
	e.attributes[k] = v
	return e
}

func (e VoidElement) RemoveAttribute(k string) Element {
	delete(e.attributes, k)
	return e
}

func (e VoidElement) ClassList() ClassList {
	return e.cl
}

func NewVoidElement(tag string) VoidElement {
	return VoidElement{tag, make(map[string]string), NewClassList()}
}

func NewImg(src, alt template.HTML) Element {
	return NewVoidElement("img").SetAttribute("alt", string(alt)).SetAttribute("src", string(src))
}

type ContentElement struct {
	VoidElement
	Contents []Element
}

func (e ContentElement) Render() template.HTML {
	e.cl.set(e)
	base := render(e.Tag, e.attributes, false)
	for _, el := range e.Contents {
		base += el.Render()
	}
	return base + template.HTML(fmt.Sprintf(`</%s>`, e.VoidElement.Tag))
}

func NewContentElement(tag string, contents []Element) ContentElement {
	return ContentElement{NewVoidElement(tag), contents}
}

func NewLiteralContentElement(tag string, content template.HTML) Element {
	return NewContentElement(tag, []Element{NewLiteralElement(content)})
}

func NewParagraph(content template.HTML) Element {
	return NewLiteralContentElement("p", content)
}

func NewHeading(level uint, content template.HTML) Element {
	return NewLiteralContentElement(fmt.Sprintf("h%d", level), content)
}
