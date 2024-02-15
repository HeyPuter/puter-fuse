package engine

type ServicesContainer struct {
	Services map[string]IService
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
