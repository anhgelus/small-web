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
