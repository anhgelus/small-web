package markdown

import "testing"

func TestError(t *testing.T) {
	v, err := Parse("**bonsoir", nil)
	if err == nil {
		t.Errorf("expected error, got %s", v)
	} else {
		t.Log(err.Pretty())
	}

	v, err = Parse("bo*nso**ir", nil)
	if err == nil {
		t.Errorf("expected error, got %s", v)
	} else {
		t.Log(err.Pretty())
	}

	v, err = Parse("test ``` hehe", nil)
	if err == nil {
		t.Errorf("expected error, got %s", v)
	} else {
		t.Log(err.Pretty())
	}
}
