package retokenizer

import (
	"github.com/datomar-labs-inc/FCT_Helpers_Go/cache"
	"github.com/golang-jwt/jwt/v4"
	"go.opentelemetry.io/otel/trace"
)

type Opts struct {
	Provider      AuthenticationProvider
	Tracer        trace.Tracer
	Cache         cache.Cache
	JWTSigningKey []byte
}

type ReTokenizer struct {
	provider AuthenticationProvider
	tracer   trace.Tracer
	cache    cache.Cache
	key      []byte
}

func New(opts *Opts) *ReTokenizer {
	return &ReTokenizer{
		provider: opts.Provider,
		tracer:   opts.Tracer,
		cache:    opts.Cache,
		key:      opts.JWTSigningKey,
	}
}

func (rt *ReTokenizer) KeyFunc() func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		return rt.key, nil
	}
}
