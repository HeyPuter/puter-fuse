package fao

type ErrDoesNotExist struct {
	Path string
}

func (e *ErrDoesNotExist) Error() string {
	return "Does not exist: " + e.Path
}

type ErrNotDirectory struct {
	Path string
}

func (e *ErrNotDirectory) Error() string {
	return "Not a directory: " + e.Path
}
