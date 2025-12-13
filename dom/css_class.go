package dom

import "strings"

type ClassList map[string]struct{}

func (cl ClassList) set(e Element) Element {
	if len(cl) == 0 {
		return e
	}
	classes := ""
	for k := range cl {
		classes += k + " "
	}
	classes = strings.TrimSpace(classes)
	return e.SetAttribute("class", classes)
}

func (cl ClassList) Has(v string) bool {
	_, ok := cl[v]
	return ok
}

func (cl ClassList) Add(v string) ClassList {
	cl[v] = struct{}{}
	return cl
}

func (cl ClassList) Remove(v string) ClassList {
	delete(cl, v)
	return cl
}

func (cl ClassList) Toggle(v string) ClassList {
	if cl.Has(v) {
		cl.Remove(v)
	} else {
		cl.Add(v)
	}
	return cl
}

func NewClassList() ClassList {
	return ClassList(make(map[string]struct{}))
}
