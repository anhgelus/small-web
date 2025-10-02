package mardown

import "fmt"

type ParseError struct {
	internal error
	lxs      lexers
}

func (e *ParseError) Error() string {
	return e.internal.Error()
}

func (e *ParseError) Pretty() string {
	lxs := e.lxs
	if lxs.lexers == nil {
		return e.internal.Error()
	}
	current := lxs.current - 1
	for lxs.Before() && lxs.Current().Type != lexerBreak {
	}
	current -= lxs.current
	contxt := ""
	ind := ""
	for lxs.Next() && lxs.Current().Type != lexerBreak {
		contxt += lxs.Current().Value
		ln := len(lxs.Current().Value)
		if lxs.current == current-1 {
			ln--
			ind += "^"
		}
		for range ln {
			ind += "~"
		}
	}
	return fmt.Sprintf("%v\n\n%s\n%s", e, contxt, ind)
}
