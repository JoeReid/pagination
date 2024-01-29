package pagination

import (
	"context"
	"net/http"
	"net/url"
)

type Middleware func(http.Handler) http.Handler

func NewMiddleware(opts ...MiddlewareOpt) Middleware {
	m := &middleware{
		maxItemsDefault: 100,
		maxItemsLimit:   100,
		rewriter:        func(u url.URL) (url.URL, bool) { return u, false },
	}

	for _, opt := range opts {
		opt(m)
	}

	return m.Handler
}

type MiddlewareOpt func(*middleware)

func WithMaxItemsDefault(maxItems int) MiddlewareOpt {
	return func(m *middleware) {
		m.maxItemsDefault = maxItems
	}
}

func WithMaxItemsLimit(maxItems int) MiddlewareOpt {
	return func(m *middleware) {
		m.maxItemsLimit = maxItems
	}
}

func WithBackwardsCompatibility(shim Rewriter) MiddlewareOpt {
	return func(m *middleware) {
		m.rewriter = shim
	}
}

// Rewriter is a function that can be used to rewrite URLs to support legacy pagination methods.
//
// If the URL is rewritten, the second return value should be true.
// If the URL is not rewritten, the given URL should be returned and the second return value should be false.
type Rewriter func(url.URL) (url.URL, bool)

type middleware struct {
	maxItemsDefault int
	maxItemsLimit   int
	rewriter        Rewriter
}

func (m *middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqURL, wasRewritten := m.rewriter(*r.URL)

		state := &state{
			current:      m.enforceRestrictions(reqURL),
			wasRewritten: wasRewritten,
			links:        make(map[string]url.URL),
		}

		if wasRewritten {
			state.links["alternate"] = reqURL
		}

		// serve the request with wrapped response writer and updated context
		next.ServeHTTP(newResponseWriter(w, state), r.WithContext(context.WithValue(r.Context(), stateKey, state)))
	})
}

func (m *middleware) enforceRestrictions(reqURL url.URL) url.URL {
	maxItems, ok := maxItems(reqURL)
	if !ok {
		return setMaxItems(reqURL, m.maxItemsDefault)
	}

	if maxItems > m.maxItemsLimit {
		return setMaxItems(reqURL, m.maxItemsLimit)
	}

	return reqURL
}
