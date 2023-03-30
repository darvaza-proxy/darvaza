// Package respond assists handlers to produce reponses
package respond

import (
	"darvaza.org/core"
	"darvaza.org/darvaza/shared/web/qlist"
)

// A Responder is a narrowed set of supported Media Types
type Responder struct {
	registry *Registry

	supported []string
	ql        qlist.QualityList
}

// Supports sets additional supported media types on the
// [Responder], but they need to exist in the registry
func (res *Responder) Supports(types ...string) *Responder {
	for _, ct := range types {
		qv, ok := res.registry.GetParsed(ct)
		if !ok {
			core.Panicf("registry doesn't support %q", ct)
		}

		res.supported = append(res.supported, ct)
		res.ql = append(res.ql, qv)
	}

	return res
}
