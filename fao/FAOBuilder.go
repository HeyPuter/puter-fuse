package fao

type FAOBuilder interface {
	Set(fao FAO)
	Add(fao FAOProxy)
	Build() FAO
}
