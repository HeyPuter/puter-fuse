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
package engine

import (
	"github.com/HeyPuter/puter-fuse-go/services"
	"github.com/btvoidx/mint"
	"github.com/spf13/viper"
)

// using this interface allows for keeping track of which methods
// of viper are being used, in case we ever swap it out.
type IConfig interface {
	GetString(key string) string
}

type ConfigService struct {
	IConfig
}

func (svc *ConfigService) Init(services services.IServiceContainer) {
	// store viper here so we can use a de-coupled interface
	svc.IConfig = viper.GetViper()

	// TODO: config is loaded in main.go right now for simplicity,
	// but it should be loaded here instead.

	mint.Emit(services.E(), ConfigLoadedEvent{})
}

func CreateConfigService() *ConfigService {
	return &ConfigService{}
}
