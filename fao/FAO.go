// GENERATED(gen.js) - DO NOT EDIT BY HAND - See meta/models.json.js

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

type FAO interface {
	Stat(path string) (NodeInfo, bool, error)
	ReadDir(path string) ([]NodeInfo, error)
	Read(path string, dest []byte, off int64) (int, error)
	Write(path string, src []byte, off int64) (int, error)
	Create(path string, name string) (NodeInfo, error)
	Truncate(path string, size uint64) error
	MkDir(path string, name string) (NodeInfo, error)
	Symlink(parent string, name string, target string) (NodeInfo, error)
	Unlink(path string) error
	Move(source string, parent string, name string) error
	ReadAll(path string) ([]byte, error)
}
