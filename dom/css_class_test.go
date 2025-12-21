package dom

import "testing"

func TestClassList(t *testing.T) {
	cl := NewClassList()
	if cl.Has("foo") {
		t.Errorf("class foo was never added")
	}
	cl.Add("foo")
	if !cl.Has("foo") {
		t.Errorf("class foo was added")
	}
	cl.Add("bar")
	if !cl.Has("bar") {
		t.Errorf("class bar was added")
	}
	if !cl.Has("foo") {
		t.Errorf("class foo was not removed")
	}
	cl.Remove("foo")
	if cl.Has("foo") {
		t.Errorf("class foo was removed")
	}
	if !cl.Has("bar") {
		t.Errorf("class bar was not removed")
	}
	cl.Toggle("foo")
	if !cl.Has("foo") {
		t.Errorf("class foo was toggled (added)")
	}
	cl.Toggle("foo")
	if cl.Has("foo") {
		t.Errorf("class foo was toggled (removed)")
	}
}
