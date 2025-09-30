package mardown

import "testing"

func TestLex(t *testing.T) {
	lxs := lex("bonjour les gens")
	if lxs.String() != "Lexers[literal(bonjour les gens) ]" {
		t.Errorf("invalid lex, got %s", lxs)
	}
	lxs = lex("# bonjour les gens")
	if lxs.String() != "Lexers[header(#) literal( bonjour les gens) ]" {
		t.Errorf("invalid lex, got %s", lxs)
	}
	lxs = lex("# bonjour les gens\nComment ça va ?")
	if lxs.String() != "Lexers[header(#) literal( bonjour les gens) break(\n) literal(Comment ça va ?) ]" {
		t.Errorf("invalid lex, got %s", lxs)
	}
	lxs = lex("**hey**, what's up?")
	if lxs.String() != "Lexers[modifier(**) literal(hey) modifier(**) literal(, what's up?) ]" {
		t.Errorf("invalid lex, got %s", lxs)
	}
	lxs = lex(`Xxx\_DarkEmperor\_xxX`)
	if lxs.String() != `Lexers[literal(Xxx) escape(\) modifier(_) literal(DarkEmperor) escape(\) modifier(_) literal(xxX) ]` {
		t.Errorf("invalid lex, got %s", lxs)
	}
}
