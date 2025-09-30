package mardown

import "fmt"

type lexerType string

const (
	lexerBreak  lexerType = "break"
	lexerEscape lexerType = "escape"

	lexerModifier lexerType = "modifier"
	lexerCode     lexerType = "code"
	lexerHeader   lexerType = "header"
	lexerQuote    lexerType = "quote"

	lexerExternal lexerType = "external"

	lexerLiteral lexerType = "literal"
)

type lexer struct {
	Type  lexerType
	Value string
}

func (l *lexer) String() string {
	return fmt.Sprintf("%s(%s)", l.Type, l.Value)
}

type lexers struct {
	current int
	lexers  []lexer
}

func (l *lexers) Next() lexer {
	l.current++
	return l.lexers[l.current]
}

func (l *lexers) Finished() bool {
	return l.current+1 >= len(l.lexers)
}

func (l *lexers) String() string {
	s := "Lexers["
	for _, l := range l.lexers {
		s += l.String() + " "
	}
	return s + "]"
}

func lex(s string) *lexers {
	lxs := &lexers{current: -1}
	var lexs []lexer
	var currentType lexerType
	var previous string
	fn := func(c rune, t lexerType) {
		if currentType != t && len(previous) > 0 {
			lexs = append(lexs, lexer{Type: currentType, Value: previous})
			previous = ""
		}
		currentType = t
		previous += string(c)
	}
	for _, c := range []rune(s) {
		switch c {
		case '*', '_':
			fn(c, lexerModifier)
		case '`':
			fn(c, lexerCode)
		case '\n':
			fn(c, lexerBreak)
		case '#':
			fn(c, lexerHeader)
		case '>':
			fn(c, lexerQuote)
		case '[', ']', '(', ')', '!':
			fn(c, lexerExternal)
		case '\\':
			fn(c, lexerEscape)
		default:
			fn(c, lexerLiteral)
		}
	}
	if len(previous) > 0 {
		lexs = append(lexs, lexer{Type: currentType, Value: previous})
	}
	lxs.lexers = lexs
	return lxs
}
