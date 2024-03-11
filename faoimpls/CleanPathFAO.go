/*
 * Copyright (C) 2024  Puter Technologies Inc.
 *
 * This file is part of puter-fuse.
 *
 * puter-fuse is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
package faoimpls

import (
	"io"
	"path/filepath"

	"github.com/HeyPuter/puter-fuse-go/fao"
)

type CleanPathFAO struct {
	fao.ProxyFAO
}

func (f *CleanPathFAO) Stat(path string) (fao.NodeInfo, bool, error) {
	path = filepath.Clean(path)
	return f.Delegate.Stat(path)
}

func (f *CleanPathFAO) ReadDir(path string) ([]fao.NodeInfo, error) {
	path = filepath.Clean(path)
	return f.Delegate.ReadDir(path)
}

func (f *CleanPathFAO) Read(path string, dest []byte, off int64) (int, error) {
	path = filepath.Clean(path)
	return f.Delegate.Read(path, dest, off)
}

func (f *CleanPathFAO) Write(path string, src []byte, off int64) (int, error) {
	path = filepath.Clean(path)
	return f.Delegate.Write(path, src, off)
}

func (f *CleanPathFAO) Create(path string, name string) (fao.NodeInfo, error) {
	path = filepath.Clean(path)
	return f.Delegate.Create(path, name)
}

func (f *CleanPathFAO) Truncate(path string, size uint64) error {
	path = filepath.Clean(path)
	return f.Delegate.Truncate(path, size)
}

func (f *CleanPathFAO) MkDir(path string, name string) (fao.NodeInfo, error) {
	path = filepath.Clean(path)
	return f.Delegate.MkDir(path, name)
}

func (f *CleanPathFAO) Symlink(parent string, name string, target string) (fao.NodeInfo, error) {
	parent = filepath.Clean(parent)
	return f.Delegate.Symlink(parent, name, target)
}

func (f *CleanPathFAO) Unlink(path string) error {
	path = filepath.Clean(path)
	return f.Delegate.Unlink(path)
}

func (f *CleanPathFAO) Move(source string, parent string, name string) error {
	source = filepath.Clean(source)
	parent = filepath.Clean(parent)
	return f.Delegate.Move(source, parent, name)
}

func (f *CleanPathFAO) ReadAll(path string) (io.ReadCloser, error) {
	path = filepath.Clean(path)
	return f.Delegate.ReadAll(path)
}
