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
package services

import mint "github.com/btvoidx/mint/context"

type ServicesContainer struct {
	Services map[string]IService
	Emitter  *mint.Emitter
}

type IServiceContainer interface {
	Set(name string, service IService)
	Get(name string) IService
	All() map[string]IService

	E() *mint.Emitter
}

func (svc *ServicesContainer) Init() {
	svc.Services = map[string]IService{}
}

func (svc *ServicesContainer) Set(name string, service IService) {
	svc.Services[name] = service
}

func (svc *ServicesContainer) Get(name string) IService {
	return svc.Services[name]
}

func (svc *ServicesContainer) All() map[string]IService {
	return svc.Services
}

func (svc *ServicesContainer) E() *mint.Emitter {
	return svc.Emitter
}
