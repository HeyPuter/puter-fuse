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

	"github.com/HeyPuter/puter-fuse/debug"
	"github.com/HeyPuter/puter-fuse/fao"
)

// LogFAO struct definition.
type LogFAO struct {
	fao.ProxyFAO
	Log debug.ILogger
}

// CreateLogFAO creates a new instance of LogFAO.
func CreateLogFAO(delegate fao.FAO, logger debug.ILogger) *LogFAO {
	return &LogFAO{
		ProxyFAO: fao.ProxyFAO{
			P_CreateProxyFAO: fao.P_CreateProxyFAO{
				Delegate: delegate,
			},
		},
		Log: logger,
	}
}

// Implementing the Stat method with logging.
func (f *LogFAO) Stat(path string) (fao.NodeInfo, bool, error) {
	f.Log.S("LogFAO").Log("Stat called with path: %s", path)
	return f.Delegate.Stat(path)
}

// Implementing the ReadDir method with logging.
func (f *LogFAO) ReadDir(path string) ([]fao.NodeInfo, error) {
	f.Log.S("LogFAO").Log("ReadDir called with path: %s", path)
	return f.Delegate.ReadDir(path)
}

// You would continue to implement the remaining methods in a similar fashion,
// logging the method name and parameters before delegating the operation to the Delegate.

// Example for Read method
func (f *LogFAO) Read(path string, dest []byte, off int64) (int, error) {
	f.Log.S("LogFAO").Log("Read called with path: %s, off: %d", path, off)
	return f.Delegate.Read(path, dest, off)
}

// Implementing the Write method with logging.
func (f *LogFAO) Write(path string, src []byte, off int64) (int, error) {
	f.Log.S("LogFAO").Log("Write called with path: %s, off: %d", path, off)
	return f.Delegate.Write(path, src, off)
}

// Implementing the Truncate method with logging.
func (f *LogFAO) Truncate(path string, size uint64) error {
	f.Log.S("LogFAO").Log("Truncate called with path: %s, size: %d", path, size)
	return f.Delegate.Truncate(path, size)
}

// Implementing the Create method with logging.
func (f *LogFAO) Create(path string, name string) (fao.NodeInfo, error) {
	f.Log.S("LogFAO").Log("Create called with path: %s, name: %s", path, name)
	return f.Delegate.Create(path, name)
}

// Implementing the MkDir method with logging.
func (f *LogFAO) MkDir(path string, name string) (fao.NodeInfo, error) {
	f.Log.S("LogFAO").Log("MkDir called with path: %s, name: %s", path, name)
	return f.Delegate.MkDir(path, name)
}

// Implementing the Symlink method with logging.
func (f *LogFAO) Symlink(parent, name, target string) (fao.NodeInfo, error) {
	f.Log.S("LogFAO").Log("Symlink called with parent: %s, name: %s, target: %s", parent, name, target)
	return f.Delegate.Symlink(parent, name, target)
}

// Implementing the Unlink method with logging.
func (f *LogFAO) Unlink(path string) error {
	f.Log.S("LogFAO").Log("Unlink called with path: %s", path)
	return f.Delegate.Unlink(path)
}

// Implementing the Move method with logging.
func (f *LogFAO) Move(source, parent, name string) error {
	f.Log.S("LogFAO").Log("Move called with source: %s, parent: %s, name: %s", source, parent, name)
	return f.Delegate.Move(source, parent, name)
}

// Implementing the ReadAll method with logging.
func (f *LogFAO) ReadAll(path string) (io.ReadCloser, error) {
	f.Log.S("LogFAO").Log("ReadAll called with path: %s", path)
	return f.Delegate.ReadAll(path)
}
