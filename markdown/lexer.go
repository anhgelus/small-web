package markdown

import (
	"fmt"
	"strings"
)

type lexerType string

const (
	lexerBreak lexerType = "break"

	lexerModifier lexerType = "modifier"

	lexerCode lexerType = "code"

	lexerHeading lexerType = "header"
	lexerQuote   lexerType = "quote"
	lexerList    lexerType = "list"

	lexerExternal lexerType = "external"
	lexerCallout  lexerType = "callout"

	lexerLiteral lexerType = "literal"
	lexerReplace lexerType = "replace"
)

type lexer struct {
	Type  lexerType
	Value string
}

func (l lexer) String() string {
	return fmt.Sprintf("%s(%s)", l.Type, strings.ReplaceAll(l.Value, "\n", `{\n}`))
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
	var sb strings.Builder
	// 8 for "Lexers[" and "]"
	sb.Grow(8 + len(l.lexers)*4)
	// always return nil error
	sb.WriteString("Lexers[")
	for _, l := range l.lexers {
		sb.WriteString(l.String())
		sb.WriteString(" ")
	}
	sb.WriteString("]")
	return sb.String()
}

func lex(s string, opt *Option) *lexers {
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
		case '*':
			if newLine && i < len(runes)-1 && runes[i+1] == ' ' {
				fn(c, lexerList, nil)
				newLine = false
				continue
			}
			fallthrough
		case '_':
			if (currentType != lexerModifier && len(previous) > 0) ||
				(len(previous) > 0 && []rune(previous)[0] != c) ||
				len(previous) >= 3 {
				lexs = append(lexs, lexer{Type: currentType, Value: previous})
				previous = ""
			}
			currentType = lexerModifier
			previous += string(c)
		case '`':
			fn(c, lexerCode, nil)
		case '\n':
			fn(c, lexerBreak, nil)
		case '#':
			fn(c, lexerHeading, nil)
		case '>':
			fn(c, lexerQuote, nil)
		case '[', ']', '!':
			if !newLine && i < len(runes)-1 {
				next := runes[i+1]
				runes := []rune(previous)
				if (c == '[' && next == '!') ||
					(c == '!' && len(runes) > 0 && previous[len(previous)-1] == '[') ||
					(c == ']' && next != '(') {
					allSpace := true
					for i := 0; allSpace && i < len(runes); i++ {
						if runes[i] != ' ' {
							allSpace = false
						}
					}
					if allSpace {
						previous = ""
					}
					fn(c, lexerCallout, nil)
					continue
				}
			}
			fallthrough
		case '(', ')':
			fn(c, lexerExternal, func(c rune) bool { return validExternal(previous + string(c)) })
		case '-', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
			fn(c, lexerList, nil)
		default:
			if _, ok := opt.Replaces[c]; ok {
				fn(c, lexerReplace, func(c rune) bool { return false })
			} else {
				fn(c, lexerLiteral, nil)
			}
		}
		newLine = c == '\n'
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
