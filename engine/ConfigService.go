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
