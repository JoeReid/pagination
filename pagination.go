package pagination

import (
	"fmt"
	"net/http"
	"strconv"
)

type Paginator struct {
	request           *http.Request
	responseHeader    http.Header
	maxItemsDefault   int
	maxItemsLimit     int
	compatibilityShim CompatibilityShim
}

func (p *Paginator) Response(w http.ResponseWriter) {
	for k, v := range p.responseHeader {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
}

func (p *Paginator) Values() (maxItems int, page string, depricated bool) {
	maxItems, page, depricated = p.compatibilityShim(p.request)
	if !depricated {
		if requested, err := strconv.Atoi(p.request.URL.Query().Get("maxItems")); err == nil {
			maxItems = requested
		}

		page = p.request.URL.Query().Get("page")
	}

	// Enforce restrictions and defaults
	maxItems = p.imposeRestrictions(maxItems)

	// If using depricated format, add an alternate header to indicate the new format
	if depricated {
		alt := *p.request.URL
		alt.Query().Set("maxItems", strconv.Itoa(maxItems))
		alt.Query().Set("page", page)

		p.responseHeader.Add("Link", fmt.Sprintf(`<%s>; rel="alternate"`, alt.String()))
		p.responseHeader.Add("Warning", `299 - "Deprecated pagination method. Please use alternate method."`)
	}

	// Add the prev header now we have processed this request
	prev := *p.request.URL
	prev.Query().Set("maxItems", strconv.Itoa(maxItems))
	prev.Query().Set("page", page)
	p.responseHeader.Add("Link", fmt.Sprintf(`<%s>; rel="prev"`, prev.String()))

	return
}

func (p *Paginator) SetNext(page string) {
	next := *p.request.URL
	next.Query().Set("page", page)

	p.responseHeader.Add("Link", fmt.Sprintf(`<%s>; rel="next"`, next.String()))
}

func (p *Paginator) SetFirst(page string) {
	first := *p.request.URL
	first.Query().Set("page", page)

	p.responseHeader.Add("Link", fmt.Sprintf(`<%s>; rel="first"`, first.String()))
}

func (p *Paginator) SetLast(page string) {
	last := *p.request.URL
	last.Query().Set("page", page)

	p.responseHeader.Add("Link", fmt.Sprintf(`<%s>; rel="last"`, last.String()))
}

func (p *Paginator) imposeRestrictions(maxItems int) int {
	if maxItems == 0 {
		maxItems = p.maxItemsDefault
	}

	if maxItems > p.maxItemsLimit {
		maxItems = p.maxItemsLimit
	}

	return maxItems
}

func NewPaginator(r *http.Request, opts ...PaginatorOpt) *Paginator {
	rtn := &Paginator{
		request:           r,
		responseHeader:    make(http.Header),
		maxItemsDefault:   100,
		maxItemsLimit:     100,
		compatibilityShim: func(r *http.Request) (int, string, bool) { return 0, "", false },
	}

	for _, opt := range opts {
		opt(rtn)
	}

	return rtn
}

type PaginatorOpt func(*Paginator)

// WithMaxItemsDefault sets the default value for maxItems if none is provided in the request.
func WithMaxItemsDefault(items int) PaginatorOpt {
	return func(p *Paginator) {
		p.maxItemsDefault = items
	}
}

// WithMaxItemsLimit sets the maximum value for maxItems that can be requested by the client.
func WithMaxItemsLimit(maxItems int) PaginatorOpt {
	return func(p *Paginator) {
		p.maxItemsLimit = maxItems
	}
}

// WithCompatibilityShim allows you to provide backwards compatibility for APIs that already
// provide an alternate pagination method.
func WithCompatibilityShim(shim CompatibilityShim) PaginatorOpt {
	return func(p *Paginator) {
		p.compatibilityShim = shim
	}
}

// CompatibilityShim is a function that identifies deprecated pagination requests
// and converts them to the new format.
//
// This function should only return depricated==true if the request is a deprecated
// format and has been converted to the new format.
//
// If depricated==false is returned, the other returned values should be zero values.
type CompatibilityShim func(*http.Request) (maxItems int, page string, ok bool)
