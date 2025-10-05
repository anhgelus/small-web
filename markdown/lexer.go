package markdown

import "fmt"

type lexerType string

const (
	lexerBreak lexerType = "break"

	lexerModifier lexerType = "modifier"

	lexerCode lexerType = "code"

	lexerHeader lexerType = "header"
	lexerQuote  lexerType = "quote"
	lexerList   lexerType = "list"

	lexerExternal lexerType = "external"

	lexerLiteral lexerType = "literal"
)

type lexer struct {
	Type  lexerType
	Value string
}

func (l lexer) String() string {
	return fmt.Sprintf("%s(%s)", l.Type, l.Value)
}

type lexers struct {
	current int
	lexers  []lexer
}

func (l *lexers) Next() bool {
	l.current++
	return !l.Finished()
}

func (l *lexers) Current() lexer {
	return l.lexers[l.current]
}

func (l *lexers) Finished() bool {
	return l.current >= len(l.lexers)
}

func (l *lexers) Before() bool {
	l.current--
	return l.current >= 0 && !l.Finished()
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
	fn := func(c rune, t lexerType, validate func(rune) bool) {
		if validate == nil {
			validate = func(r rune) bool { return true }
		}
		if (currentType != t || !validate(c)) && len(previous) > 0 {
			lexs = append(lexs, lexer{Type: currentType, Value: previous})
			previous = ""
		}
		currentType = t
		previous += string(c)
	}
	newLine := true
	literalNext := false
	runes := []rune(s)
	for i, c := range runes {
		if literalNext {
			fn(c, lexerLiteral, nil)
			literalNext = false
			continue
		}
		if c == '\\' {
			literalNext = true
			continue
		}
		switch c {
		case '*', '_':
			if c == '*' && newLine && i < len(runes)-1 && runes[i+1] == ' ' {
				fn(c, lexerList, nil)
			} else {
				if (currentType != lexerModifier && len(previous) > 0) ||
					(len(previous) > 0 && []rune(previous)[0] != c) ||
					len(previous) >= 3 {
					lexs = append(lexs, lexer{Type: currentType, Value: previous})
					previous = ""
				}
				currentType = lexerModifier
				previous += string(c)
			}
			newLine = false
		case '`':
			newLine = false
			fn(c, lexerCode, nil)
		case '\n':
			newLine = true
			fn(c, lexerBreak, nil)
		case '#':
			newLine = false
			fn(c, lexerHeader, nil)
		case '>':
			newLine = false
			fn(c, lexerQuote, nil)
		case '[', ']', '(', ')', '!':
			newLine = false
			fn(c, lexerExternal, func(c rune) bool { return validExternal(previous + string(c)) })
		case '-', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
			newLine = false
			fn(c, lexerList, nil)
		default:
			newLine = false
			fn(c, lexerLiteral, nil)
		}
	}
	if len(previous) > 0 {
		lexs = append(lexs, lexer{Type: currentType, Value: previous})
	}
	lxs.lexers = lexs
	return lxs
}

func validExternal(s string) bool {
	switch s {
	// start
	case "![", "[":
		return true
	// mid
	case "](":
		return true
	// end
	case ")":
		return true
	default:
		return false
	}
}
