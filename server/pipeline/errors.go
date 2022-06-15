package pipeline

type ErrNotFound struct {
	Msg string
}

func (e ErrNotFound) Error() string {
	return e.Msg
}

func IsErrNotFound(err error) bool {
	_, ok := err.(ErrNotFound)
	return ok
}
