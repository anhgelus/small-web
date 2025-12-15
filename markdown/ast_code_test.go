package markdown

import "testing"

func TestCode(t *testing.T) {
	got, err := Parse("`mono`", nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `<p><code>mono</code></p>` {
		t.Errorf("invalid value, got %s", got)
	}

	got, err = Parse("bonjour `code` !", nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `<p>bonjour <code>code</code> !</p>` {
		t.Errorf("invalid value, got %s", got)
	}

	got, err = Parse(
		"```\n"+"raw\nhehe"+"```",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `<pre><code>raw
hehe</code></pre>` {
		t.Errorf("invalid value, got %s", got)
	}
}
