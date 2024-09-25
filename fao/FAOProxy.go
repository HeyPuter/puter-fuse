package fao

type FAOProxy interface {
	FAO
	SetDelegate(delegate FAO)
}
