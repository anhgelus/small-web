package markdown

import "testing"

func TestError(t *testing.T) {
	v, err := Parse("**bonsoir")
	if err == nil {
		t.Errorf("expected error, got %s", v)
	} else {
		t.Log(err.Pretty())
	}

	v, err = Parse("bo*nso**ir")
	if err == nil {
		t.Errorf("expected error, got %s", v)
	} else {
		t.Log(err.Pretty())
	}

	v, err = Parse("test ``` hehe")
	if err == nil {
		t.Errorf("expected error, got %s", v)
	} else {
		t.Log(err.Pretty())
	}
}
