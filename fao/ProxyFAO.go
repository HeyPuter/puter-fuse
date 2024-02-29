// GENERATED(gen.js) - DO NOT EDIT BY HAND - See meta/models.json.js

package fao

import (
	"io"
)

type P_CreateProxyFAO struct {
	Delegate FAO
}

type ProxyFAO struct {
	P_CreateProxyFAO
}

func CreateProxyFAO(params P_CreateProxyFAO) *ProxyFAO {
	return &ProxyFAO{params}
}

func (p *ProxyFAO) Stat(path string) (NodeInfo, bool, error) {
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
func (p *ProxyFAO) Create(path string, name string) (NodeInfo, error) {
	return p.Delegate.Create(path, name)
}
func (p *ProxyFAO) Truncate(path string, size uint64) error {
	return p.Delegate.Truncate(path, size)
}
func (p *ProxyFAO) MkDir(path string, name string) (NodeInfo, error) {
	return p.Delegate.MkDir(path, name)
}
func (p *ProxyFAO) Symlink(parent string, name string, target string) (NodeInfo, error) {
	return p.Delegate.Symlink(parent, name, target)
}
func (p *ProxyFAO) Unlink(path string) error {
	return p.Delegate.Unlink(path)
}
func (p *ProxyFAO) Move(source string, parent string, name string) error {
	return p.Delegate.Move(source, parent, name)
}
func (p *ProxyFAO) ReadAll(path string) (io.ReadCloser, error) {
	return p.Delegate.ReadAll(path)
}
