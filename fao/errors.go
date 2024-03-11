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
package fao

import (
	"fmt"
	"syscall"
)

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

type FAOError struct {
	Errno syscall.Errno
	From  error
}

func (e *FAOError) Error() string {
	return e.From.Error()
}

func Errorf(errno syscall.Errno, format string, a ...interface{}) error {
	return &FAOError{
		Errno: errno,
		From:  fmt.Errorf(format, a...),
	}
}
