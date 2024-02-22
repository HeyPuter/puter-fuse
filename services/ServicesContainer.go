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
