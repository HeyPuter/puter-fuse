package faoimpls

import "github.com/HeyPuter/puter-fuse/fao"

// Builds a composite FAO without adding any additional behavior.
// Useful as a placeholder where other builders might be used.
type NullFAOBuilder struct {
	subject fao.FAO
}

func (fb *NullFAOBuilder) Set(subject fao.FAO) {
	fb.subject = subject
}

func (fb *NullFAOBuilder) Add(proxy fao.FAOProxy) {
	proxy.SetDelegate(fb.subject)
	fb.subject = proxy
}

func (fb *NullFAOBuilder) Build() fao.FAO {
	return fb.subject
}
