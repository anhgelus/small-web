package mardown

type astHeader struct {
	level   uint
	content []block
}

func (a *astHeader) Eval() error {
	return nil
}

func header(lxs lexers) (block, error) {
	b := &astHeader{level: uint(len(lxs.Current().Value))}
	for lxs.Next() && lxs.Current().Type != lexerBreak {
		bl, err := getBlock(lxs)
		if err != nil {
			return nil, err
		}
		if h, ok := bl.(*astHeader); ok {
			//TODO: handle
		}
		b.content = append(b.content, bl)
	}
	return b, nil
}
