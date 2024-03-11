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
// GENERATED(gen.js) - DO NOT EDIT BY HAND - See meta/models.json.js

package fao

import (
	"io"
)

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
	ReadAll(path string) (io.ReadCloser, error)
}
