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
package debug

import (
	"fmt"

	"github.com/HeyPuter/puter-fuse/services"
)

type ILogger interface {
	Log(format string, args ...interface{})
	Sub(crumbs []string) *Logger
	S(str string) *Logger
}

type Logger struct {
	Crumbs []string
}

func NewLogger(format string, args ...interface{}) *Logger {
	return &Logger{
		Crumbs: []string{fmt.Sprintf(format, args...)},
	}
}

func (l *Logger) log(format, col string, args ...interface{}) {
	str := ""
	for _, crumb := range l.Crumbs {
		str += fmt.Sprintf("[%s] ", crumb)
	}

	// Wrap in colour <col>
	str = fmt.Sprintf("\033[%sm%s\033[0m", col, str)

	str += fmt.Sprintf(format, args...)

	fmt.Println(str)
}

func (l *Logger) Log(format string, args ...interface{}) {
	l.log(format, "34;1", args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(format, "31;1", args...)
}

func (l *Logger) Sub(crumbs []string) *Logger {
	return &Logger{
		Crumbs: append(l.Crumbs, crumbs...),
	}
}

func (l *Logger) S(str string) *Logger {
	return &Logger{
		Crumbs: append(l.Crumbs, str),
	}
}

type LogService struct {
	Logger *Logger

	services services.IServiceContainer
}

func (svc *LogService) Init(services services.IServiceContainer) {
	svc.Logger = &Logger{}
}

func (svc *LogService) Log(msg string) {
	svc.Logger.Log(msg)
}

func (svc *LogService) GetLogger(format string, args ...interface{}) *Logger {
	logger := &Logger{
		Crumbs: append(svc.Logger.Crumbs, fmt.Sprintf(format, args...)),
	}
	return logger
}
