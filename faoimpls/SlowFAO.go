package faoimpls

import (
	"time"

	"github.com/HeyPuter/puter-fuse-go/fao"
)

type SlowFAO struct {
	fao.ProxyFAO
	Delay time.Duration
}

func CreateSlowFAO(delegate fao.FAO, d time.Duration) *SlowFAO {
	fao := &SlowFAO{
		ProxyFAO: fao.ProxyFAO{
			P_CreateProxyFAO: fao.P_CreateProxyFAO{
				Delegate: delegate,
			},
		},
		Delay: d,
	}

	return fao
}

func (f *SlowFAO) Stat(path string) (fao.NodeInfo, bool, error) {
	time.Sleep(f.Delay)
	return f.Delegate.Stat(path)
}

func (f *SlowFAO) ReadDir(path string) ([]fao.NodeInfo, error) {
	time.Sleep(f.Delay)
	return f.Delegate.ReadDir(path)
}

func (f *SlowFAO) Create(path string, name string) (fao.NodeInfo, error) {
	time.Sleep(f.Delay)
	return f.Delegate.Create(path, name)
}

func (f *SlowFAO) MkDir(parent, path string) (fao.NodeInfo, error) {
	time.Sleep(f.Delay)
	return f.Delegate.MkDir(parent, path)
}

func (f *SlowFAO) Read(path string, dest []byte, off int64) (int, error) {
	time.Sleep(f.Delay)
	return f.Delegate.Read(path, dest, off)
}

func (f *SlowFAO) Write(path string, src []byte, off int64) (int, error) {
	time.Sleep(f.Delay)
	return f.Delegate.Write(path, src, off)
}

func (f *SlowFAO) Truncate(path string, size uint64) error {
	time.Sleep(f.Delay)
	return f.Delegate.Truncate(path, size)
}

func (f *SlowFAO) Symlink(parent, name, target string) (fao.NodeInfo, error) {
	time.Sleep(f.Delay)
	return f.Delegate.Symlink(parent, name, target)
}

func (f *SlowFAO) Unlink(path string) error {
	time.Sleep(f.Delay)
	return f.Delegate.Unlink(path)
}

func (f *SlowFAO) Move(source, parent, name string) error {
	time.Sleep(f.Delay)
	return f.Delegate.Move(source, parent, name)
}
