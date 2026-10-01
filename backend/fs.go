package backend

import (
	"errors"
	"io/fs"
)

type JoinFS struct {
	fs []fs.FS
}

func Join(fs ...fs.FS) *JoinFS {
	return &JoinFS{fs}
}

func (j *JoinFS) Open(name string) (fs.File, error) {
	println(name)
	for _, ff := range j.fs {
		f, err := ff.Open(name)
		if err == nil {
			return f, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	return nil, fs.ErrNotExist
}
