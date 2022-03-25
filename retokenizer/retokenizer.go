package retokenizer

import (
	"github.com/datomar-labs-inc/FCT_Helpers_Go/cache"
	"go.opentelemetry.io/otel/trace"
)

type ReTokenizer struct {
	provider AuthenticationProvider
	tracer   trace.Tracer
	cache    cache.Cache
}

func New(provider AuthenticationProvider, tracer trace.Tracer, cache cache.Cache) *ReTokenizer {
	return &ReTokenizer{
		provider: provider,
		tracer:   tracer,
		cache:    cache,
	}
}
