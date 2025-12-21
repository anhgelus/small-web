package dom

import (
	"html/template"
	"testing"
)

func TestRender(t *testing.T) {
	fn := func(tag string, attributes map[string]string, endSlash bool, expected string) func(*testing.T) {
		return func(t *testing.T) {
			t.Parallel()
			got := string(render(tag, attributes, endSlash))
			if got != expected {
				t.Errorf("invalid value, got %s", got)
			}
		}
	}
	t.Run("render", func(t *testing.T) {
		t.Run("simple", fn("p", map[string]string{}, false, "<p>"))
		t.Run("endslash", fn("img", map[string]string{}, true, "<img />"))
		t.Run("attributes", fn("a", map[string]string{"href": "link"}, false, `<a href="link">`))
	})
}

func TestLiteralElement(t *testing.T) {
	fn := func(raw, expected string) func(*testing.T) {
		return func(t *testing.T) {
			t.Parallel()
			got := NewLiteralElement(template.HTML(raw))
			if string(got.Render()) != expected {
				t.Errorf("invalid value, got %s", got)
			}
		}
	}
	t.Run("render", func(t *testing.T) {
		t.Run("simple", fn("hello world", "hello world"))
	})
}

func TestVoidElement(t *testing.T) {
	fn := func(tag string, attributes map[string]string, expected string) func(*testing.T) {
		return func(t *testing.T) {
			t.Parallel()
			e := NewVoidElement(tag)
			for k, v := range attributes {
				e.SetAttribute(k, v)
				if !e.HasAttribute(k) {
					t.Errorf("doesn't not have attribute %s after being added", k)
				}
			}
			got := string(e.Render())
			if got != expected {
				t.Errorf("invalid value, got %s", got)
			}
		}
	}
	t.Run("no_attributes", func(t *testing.T) {
		t.Run("simple1", fn("br", nil, "<br />"))
		t.Run("simple2", fn("img", nil, "<img />"))
	})
	t.Run("attributes", func(t *testing.T) {
		t.Run("one", fn("img", map[string]string{"src": "link"}, `<img src="link" />`))
		t.Run("two", fn("img", map[string]string{"src": "link", "alt": "well"}, `<img src="link" alt="well" />`))
	})
}

func TestContentElement(t *testing.T) {
	fn := func(tag string, els []Element, attributes map[string]string, expected string) func(*testing.T) {
		return func(t *testing.T) {
			t.Parallel()
			e := NewContentElement(tag, els)
			for k, v := range attributes {
				e.SetAttribute(k, v)
				if !e.HasAttribute(k) {
					t.Errorf("doesn't not have attribute %s after being added", k)
				}
			}
			got := string(e.Render())
			if got != expected {
				t.Errorf("invalid value, got %s", got)
			}
		}
	}
	fnLiteral := func(tag string, lit string, attributes map[string]string, expected string) func(*testing.T) {
		return fn(tag, []Element{NewLiteralElement(template.HTML(lit))}, attributes, expected)
	}
	t.Run("no_attributes", func(t *testing.T) {
		t.Run("simple", fnLiteral("p", "", nil, `<p></p>`))
		t.Run("literal", fnLiteral("p", "content", nil, `<p>content</p>`))
		t.Run("elements", fn("div", []Element{NewVoidElement("img"), NewVoidElement("br")}, nil, `<div><img /><br /></div>`))
	})
	t.Run("attributes", func(t *testing.T) {
		t.Run("simple_one", fnLiteral("script", "", map[string]string{"src": "link"}, `<script src="link"></script>`))
		t.Run("literal_one", fnLiteral("p", "content", map[string]string{"class": "hey"}, `<p class="hey">content</p>`))
	})
}

func TestElement_ClassList(t *testing.T) {
	fn := func(e Element, expected string, clazz ...string) func(*testing.T) {
		return func(t *testing.T) {
			t.Parallel()
			cl := e.ClassList()
			for _, s := range clazz {
				cl.Add(s)
				if !cl.Has(s) {
					t.Errorf("doesn't have class %s after being added", s)
				}
			}
			got := string(e.Render())
			if got != expected {
				t.Errorf("invalid value, got %s", got)
			}
		}
	}
	t.Run("add", func(t *testing.T) {
		t.Run("empty", fn(NewVoidElement("img"), `<img />`))
		t.Run("one", fn(NewVoidElement("img"), `<img class="bg" />`, "bg"))
		t.Run("two", fn(NewVoidElement("img"), `<img class="bg large" />`, "bg", "large"))
	})
}
