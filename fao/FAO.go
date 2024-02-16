package fao

type FAO interface {
	Stat(path string) (NodeInfo, error)
	ReadDir(path string) ([]NodeInfo, error)
	Read(path string, dest []byte, off int64) (int, error)
	Write(path string, src []byte, off int64) (int, error)
	Truncate(path string, size int64) error
	Link(parent string, name string, target string) error
	ReadAll(path string) ([]byte, error)
}
