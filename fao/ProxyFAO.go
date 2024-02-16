// GENERATED(gen.js) - DO NOT EDIT BY HAND - See meta/models.json.js

package fao

type P_CreateProxyFAO struct {
	Delegate FAO
}

type ProxyFAO struct {
	P_CreateProxyFAO
}

func CreateProxyFAO(params P_CreateProxyFAO) *ProxyFAO {
	return &ProxyFAO{params}
}

func (p *ProxyFAO) Stat(path string) (NodeInfo, error) {
	return p.Delegate.Stat(path)
}
func (p *ProxyFAO) ReadDir(path string) ([]NodeInfo, error) {
	return p.Delegate.ReadDir(path)
}
func (p *ProxyFAO) Read(path string, dest []byte, off int64) (int, error) {
	return p.Delegate.Read(path, dest, off)
}
func (p *ProxyFAO) Write(path string, src []byte, off int64) (int, error) {
	return p.Delegate.Write(path, src, off)
}
func (p *ProxyFAO) Truncate(path string, size int64) error {
	return p.Delegate.Truncate(path, size)
}
func (p *ProxyFAO) Link(parent string, name string, target string) error {
	return p.Delegate.Link(parent, name, target)
}
func (p *ProxyFAO) ReadAll(path string) ([]byte, error) {
	return p.Delegate.ReadAll(path)
}
