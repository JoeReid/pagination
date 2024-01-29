package pagination

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxItems(t *testing.T) {
	tests := []struct {
		name           string
		opts           []MiddlewareOpt
		url            string
		expectMaxItems int
	}{
		{
			name:           "default",
			opts:           []MiddlewareOpt{},
			url:            "http://example.com/items",
			expectMaxItems: 100,
		},
		{
			name:           "default with max items",
			opts:           []MiddlewareOpt{},
			url:            "http://example.com/items?maxItems=50",
			expectMaxItems: 50,
		},
		{
			name:           "default with max items over limit",
			opts:           []MiddlewareOpt{},
			url:            "http://example.com/items?maxItems=150",
			expectMaxItems: 100,
		},
		{
			name:           "custom default and limit no max items",
			opts:           []MiddlewareOpt{WithMaxItemsDefault(10), WithMaxItemsLimit(20)},
			url:            "http://example.com/items",
			expectMaxItems: 10,
		},
		{
			name:           "custom default and limit max items",
			opts:           []MiddlewareOpt{WithMaxItemsDefault(10), WithMaxItemsLimit(20)},
			url:            "http://example.com/items?maxItems=15",
			expectMaxItems: 15,
		},
		{
			name:           "custom default and limit over limit",
			opts:           []MiddlewareOpt{WithMaxItemsDefault(10), WithMaxItemsLimit(20)},
			url:            "http://example.com/items?maxItems=150",
			expectMaxItems: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				m   = NewMiddleware(tt.opts...)
				rec = httptest.NewRecorder()
				req = httptest.NewRequest("GET", tt.url, nil)
			)

			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectMaxItems, MaxItems(r))
			})).ServeHTTP(rec, req)
		})
	}
}

func TestPage(t *testing.T) {
	tests := []struct {
		name       string
		opts       []MiddlewareOpt
		url        string
		expectPage string
	}{
		{
			name:       "default",
			opts:       []MiddlewareOpt{},
			url:        "http://example.com/items",
			expectPage: "",
		},
		{
			name:       "default with page",
			opts:       []MiddlewareOpt{},
			url:        "http://example.com/items?page=abc",
			expectPage: "abc",
		},
		{
			name: "with backwards compatibility",
			opts: []MiddlewareOpt{
				WithBackwardsCompatibility(func(url.URL) (url.URL, bool) {
					u, _ := url.Parse("http://example.com/items?page=test")
					return *u, true
				}),
			},
			url:        "http://example.com/items",
			expectPage: "test",
		},
		{
			name: "with unused backwards compatibility",
			opts: []MiddlewareOpt{
				WithBackwardsCompatibility(func(url.URL) (url.URL, bool) { return url.URL{}, false }),
			},
			url:        "http://example.com/items",
			expectPage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				m   = NewMiddleware(tt.opts...)
				rec = httptest.NewRecorder()
				req = httptest.NewRequest("GET", tt.url, nil)
			)

			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectPage, Page(r))
			})).ServeHTTP(rec, req)
		})
	}
}

func TestSetNext(t *testing.T) {
	tests := []struct {
		name          string
		opts          []MiddlewareOpt
		url           string
		next          *string
		expectHeaders map[string][]string
	}{
		{
			name: "no next",
			opts: []MiddlewareOpt{},
			url:  "http://example.com/items",
			next: nil,
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
			},
		},
		{
			name: "with next",
			opts: []MiddlewareOpt{},
			url:  "http://example.com/items",
			next: str("abc"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=100&page=abc>; rel="next"`},
			},
		},
		{
			name: "with next and max items",
			opts: []MiddlewareOpt{},
			url:  "http://example.com/items?maxItems=10",
			next: str("abc"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=10&page=abc>; rel="next"`},
			},
		},
		{
			name: "with next and max items over limit",
			opts: []MiddlewareOpt{},
			url:  "http://example.com/items?maxItems=150",
			next: str("abc"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=100&page=abc>; rel="next"`},
			},
		},
		{
			name: "with rewritten",
			opts: []MiddlewareOpt{
				WithBackwardsCompatibility(func(url.URL) (url.URL, bool) {
					u, _ := url.Parse("http://example.com/items?page=abc")
					return *u, true
				}),
			},
			url:  "http://example.com/items?oldParam=test",
			next: nil,
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=100&page=abc>; rel="alternate"`},
				"Warning":      {`299 - "Deprecated pagination method. Please use alternate method."`},
			},
		},
		{
			name: "with rewritten and next",
			opts: []MiddlewareOpt{
				WithBackwardsCompatibility(func(url.URL) (url.URL, bool) {
					u, _ := url.Parse("http://example.com/items?page=abc")
					return *u, true
				}),
			},
			url:  "http://example.com/items?oldParam=test",
			next: str("def"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link": {
					`<http://example.com/items?maxItems=100&page=abc>; rel="alternate"`,
					`<http://example.com/items?maxItems=100&page=def>; rel="next"`,
				},
				"Warning": {`299 - "Deprecated pagination method. Please use alternate method."`},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				m   = NewMiddleware(tt.opts...)
				rec = httptest.NewRecorder()
				req = httptest.NewRequest("GET", tt.url, nil)
			)

			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.next != nil {
					SetNext(r, *tt.next)
				}

				w.Write([]byte("test"))
			})).ServeHTTP(rec, req)

			equalHeaders(t, http.Header(tt.expectHeaders), rec.Header())
		})
	}
}

func TestSetPrev(t *testing.T) {
	tests := []struct {
		name          string
		opts          []MiddlewareOpt
		url           string
		prev          *string
		expectHeaders map[string][]string
	}{
		{
			name: "no next",
			opts: []MiddlewareOpt{},
			url:  "http://example.com/items",
			prev: nil,
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
			},
		},
		{
			name: "with next",
			opts: []MiddlewareOpt{},
			url:  "http://example.com/items",
			prev: str("abc"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=100&page=abc>; rel="prev"`},
			},
		},
		{
			name: "with next and max items",
			opts: []MiddlewareOpt{},
			url:  "http://example.com/items?maxItems=10",
			prev: str("abc"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=10&page=abc>; rel="prev"`},
			},
		},
		{
			name: "with next and max items over limit",
			opts: []MiddlewareOpt{},
			url:  "http://example.com/items?maxItems=150",
			prev: str("abc"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=100&page=abc>; rel="prev"`},
			},
		},
		{
			name: "with rewritten",
			opts: []MiddlewareOpt{
				WithBackwardsCompatibility(func(url.URL) (url.URL, bool) {
					u, _ := url.Parse("http://example.com/items?page=abc")
					return *u, true
				}),
			},
			url:  "http://example.com/items?oldParam=test",
			prev: nil,
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=100&page=abc>; rel="alternate"`},
				"Warning":      {`299 - "Deprecated pagination method. Please use alternate method."`},
			},
		},
		{
			name: "with rewritten and next",
			opts: []MiddlewareOpt{
				WithBackwardsCompatibility(func(url.URL) (url.URL, bool) {
					u, _ := url.Parse("http://example.com/items?page=abc")
					return *u, true
				}),
			},
			url:  "http://example.com/items?oldParam=test",
			prev: str("def"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link": {
					`<http://example.com/items?maxItems=100&page=abc>; rel="alternate"`,
					`<http://example.com/items?maxItems=100&page=def>; rel="prev"`,
				},
				"Warning": {`299 - "Deprecated pagination method. Please use alternate method."`},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				m   = NewMiddleware(tt.opts...)
				rec = httptest.NewRecorder()
				req = httptest.NewRequest("GET", tt.url, nil)
			)

			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.prev != nil {
					SetPrev(r, *tt.prev)
				}

				w.Write([]byte("test"))
			})).ServeHTTP(rec, req)

			equalHeaders(t, http.Header(tt.expectHeaders), rec.Header())
		})
	}
}

func TestSetFirst(t *testing.T) {
	tests := []struct {
		name          string
		opts          []MiddlewareOpt
		url           string
		first         *string
		expectHeaders map[string][]string
	}{
		{
			name:  "no next",
			opts:  []MiddlewareOpt{},
			url:   "http://example.com/items",
			first: nil,
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
			},
		},
		{
			name:  "with next",
			opts:  []MiddlewareOpt{},
			url:   "http://example.com/items",
			first: str("abc"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=100&page=abc>; rel="first"`},
			},
		},
		{
			name:  "with next and max items",
			opts:  []MiddlewareOpt{},
			url:   "http://example.com/items?maxItems=10",
			first: str("abc"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=10&page=abc>; rel="first"`},
			},
		},
		{
			name:  "with next and max items over limit",
			opts:  []MiddlewareOpt{},
			url:   "http://example.com/items?maxItems=150",
			first: str("abc"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=100&page=abc>; rel="first"`},
			},
		},
		{
			name: "with rewritten",
			opts: []MiddlewareOpt{
				WithBackwardsCompatibility(func(url.URL) (url.URL, bool) {
					u, _ := url.Parse("http://example.com/items?page=abc")
					return *u, true
				}),
			},
			url:   "http://example.com/items?oldParam=test",
			first: nil,
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=100&page=abc>; rel="alternate"`},
				"Warning":      {`299 - "Deprecated pagination method. Please use alternate method."`},
			},
		},
		{
			name: "with rewritten and next",
			opts: []MiddlewareOpt{
				WithBackwardsCompatibility(func(url.URL) (url.URL, bool) {
					u, _ := url.Parse("http://example.com/items?page=abc")
					return *u, true
				}),
			},
			url:   "http://example.com/items?oldParam=test",
			first: str("def"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link": {
					`<http://example.com/items?maxItems=100&page=abc>; rel="alternate"`,
					`<http://example.com/items?maxItems=100&page=def>; rel="first"`,
				},
				"Warning": {`299 - "Deprecated pagination method. Please use alternate method."`},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				m   = NewMiddleware(tt.opts...)
				rec = httptest.NewRecorder()
				req = httptest.NewRequest("GET", tt.url, nil)
			)

			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.first != nil {
					SetFirst(r, *tt.first)
				}

				w.Write([]byte("test"))
			})).ServeHTTP(rec, req)

			equalHeaders(t, http.Header(tt.expectHeaders), rec.Header())
		})
	}
}

func TestSetLast(t *testing.T) {
	tests := []struct {
		name          string
		opts          []MiddlewareOpt
		url           string
		last          *string
		expectHeaders map[string][]string
	}{
		{
			name: "no next",
			opts: []MiddlewareOpt{},
			url:  "http://example.com/items",
			last: nil,
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
			},
		},
		{
			name: "with next",
			opts: []MiddlewareOpt{},
			url:  "http://example.com/items",
			last: str("abc"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=100&page=abc>; rel="last"`},
			},
		},
		{
			name: "with next and max items",
			opts: []MiddlewareOpt{},
			url:  "http://example.com/items?maxItems=10",
			last: str("abc"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=10&page=abc>; rel="last"`},
			},
		},
		{
			name: "with next and max items over limit",
			opts: []MiddlewareOpt{},
			url:  "http://example.com/items?maxItems=150",
			last: str("abc"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=100&page=abc>; rel="last"`},
			},
		},
		{
			name: "with rewritten",
			opts: []MiddlewareOpt{
				WithBackwardsCompatibility(func(url.URL) (url.URL, bool) {
					u, _ := url.Parse("http://example.com/items?page=abc")
					return *u, true
				}),
			},
			url:  "http://example.com/items?oldParam=test",
			last: nil,
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link":         {`<http://example.com/items?maxItems=100&page=abc>; rel="alternate"`},
				"Warning":      {`299 - "Deprecated pagination method. Please use alternate method."`},
			},
		},
		{
			name: "with rewritten and next",
			opts: []MiddlewareOpt{
				WithBackwardsCompatibility(func(url.URL) (url.URL, bool) {
					u, _ := url.Parse("http://example.com/items?page=abc")
					return *u, true
				}),
			},
			url:  "http://example.com/items?oldParam=test",
			last: str("def"),
			expectHeaders: map[string][]string{
				"Content-Type": {"text/plain; charset=utf-8"},
				"Link": {
					`<http://example.com/items?maxItems=100&page=abc>; rel="alternate"`,
					`<http://example.com/items?maxItems=100&page=def>; rel="last"`,
				},
				"Warning": {`299 - "Deprecated pagination method. Please use alternate method."`},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				m   = NewMiddleware(tt.opts...)
				rec = httptest.NewRecorder()
				req = httptest.NewRequest("GET", tt.url, nil)
			)

			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.last != nil {
					SetLast(r, *tt.last)
				}

				w.Write([]byte("test"))
			})).ServeHTTP(rec, req)

			equalHeaders(t, http.Header(tt.expectHeaders), rec.Header())
		})
	}
}

func equalHeaders(t *testing.T, expect, actual http.Header) {
	for key, values := range expect {
		assert.ElementsMatch(t, values, actual[key])
	}

	for key, values := range actual {
		assert.ElementsMatch(t, expect[key], values)
	}
}

func str(s string) *string {
	return &s
}
