package markdown

import "testing"

func TestLex(t *testing.T) {
	opt := new(Option)
	lxs := lex("bonjour les gens", opt)
	if lxs.String() != "Lexers[literal(bonjour les gens) ]" {
		t.Errorf("invalid lex, got %s", lxs)
	}
	lxs = lex("# bonjour les gens", opt)
	if lxs.String() != "Lexers[header(#) literal( bonjour les gens) ]" {
		t.Errorf("invalid lex, got %s", lxs)
	}
	lxs = lex("# bonjour les gens\nComment ça va ?", opt)
	if lxs.String() != `Lexers[header(#) literal( bonjour les gens) break({\n}) literal(Comment ça va ?) ]` {
		t.Errorf("invalid lex, got %s", lxs)
	}
	lxs = lex("***hey***, what's up?", opt)
	if lxs.String() != "Lexers[modifier(***) literal(hey) modifier(***) literal(, what's up?) ]" {
		t.Errorf("invalid lex, got %s", lxs)
	}
	lxs = lex(`Xxx\_DarkEmperor\_xxX`, opt)
	if lxs.String() != `Lexers[literal(Xxx_DarkEmperor_xxX) ]` {
		t.Errorf("invalid lex, got %s", lxs)
	}
	lxs = lex(`* list`, opt)
	if lxs.String() != `Lexers[list(*) literal( list) ]` {
		t.Errorf("invalid lex, got %s", lxs)
	}
	lxs = lex(`> [!NOTE] title
> hey`, opt)
	if lxs.String() != `Lexers[quote(>) literal( ) callout([!) literal(NOTE) callout(]) literal( title) break({\n}) quote(>) literal( hey) ]` {
		t.Errorf("invalid lex, got %s", lxs)
	}
}

func TestLex_Replacer(t *testing.T) {
	opt := &Option{
		Replaces: map[rune]string{'~': "&thinsp;"},
	}
	lxs := lex("bonjour les gens", opt)
	if lxs.String() != "Lexers[literal(bonjour les gens) ]" {
		t.Errorf("invalid lex, got %s", lxs)
	}
	lxs = lex("bonjour les gens~!", opt)
	if lxs.String() != "Lexers[literal(bonjour les gens) replace(~) external(!) ]" {
		t.Errorf("invalid lex, got %s", lxs)
	}
}
