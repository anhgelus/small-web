package markdown

import "testing"

func TestParagraph(t *testing.T) {
	c, err := Parse("bonsoir", nil)
	if err != nil {
		t.Fatal(err)
	}
	if c != "<p>bonsoir</p>" {
		t.Errorf("failed, got %s", c)
	}
}

func TestParagraph_Replacer(t *testing.T) {
	opt := &Option{
		Replaces: map[rune]string{'~': "&thinsp;"},
	}
	c, err := Parse("bonsoir", opt)
	if err != nil {
		t.Fatal(err)
	}
	if c != "<p>bonsoir</p>" {
		t.Errorf("failed, got %s", c)
	}
	c, err = Parse("bonsoir~!", opt)
	if err != nil {
		t.Fatal(err)
	}
	if c != "<p>bonsoir&thinsp;!</p>" {
		t.Errorf("failed, got %s", c)
	}
}
