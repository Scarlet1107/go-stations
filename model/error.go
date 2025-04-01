package model

type ErrNotFound struct {
}

func (e *ErrNotFound) Error() string {
	return "todo not found."
}
