package template

import "fmt"

type ErrNotYetRefreshed struct {
	error
}

type ErrNotValidFile struct {
	Path string
}

func (err *ErrNotValidFile) Error() string {
	return fmt.Sprintf("'%s' not a valid file", err.Path)
}
