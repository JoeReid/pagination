package pagination

import (
	"net/http"
	"net/url"
)

func MaxItems(r *http.Request) int {
	p, ok := r.Context().Value(stateKey).(*state)
	if !ok {
		return 0
	}

	rtn, _ := maxItems(p.current)
	return rtn
}

func Page(r *http.Request) string {
	p, ok := r.Context().Value(stateKey).(*state)
	if !ok {
		return ""
	}

	rtn, _ := page(p.current)
	return rtn
}

func SetNext(r *http.Request, page string) {
	setLink(r, "next", page)
}

func SetPrev(r *http.Request, page string) {
	setLink(r, "prev", page)
}

func SetFirst(r *http.Request, page string) {
	setLink(r, "first", page)
}

func SetLast(r *http.Request, page string) {
	setLink(r, "last", page)
}

func setLink(r *http.Request, name, page string) {
	p, ok := r.Context().Value(stateKey).(*state)
	if !ok {
		return
	}

	p.links[name] = setPage(p.current, page)
}

type stateContextKey string

var stateKey = stateContextKey("pagination.state")

type state struct {
	current      url.URL
	wasRewritten bool
	links        map[string]url.URL
}
