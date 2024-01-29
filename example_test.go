package pagination_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"

	"github.com/JoeReid/pagination"
	"github.com/go-chi/chi/v5"
)

func ExampleMiddleware_simple() {
	middleware := pagination.NewMiddleware()

	r := chi.NewRouter()
	r.With(middleware).Get("/items", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("request: maxItems=%d page=%q\n", pagination.MaxItems(r), pagination.Page(r))

		pagination.SetNext(r, "test") // dummy example token, replace with your own
		w.Write([]byte("...Data..."))
	})

	// Simulate a request to the handler and print the response headers
	// real applications should use http.ListenAndServe or similar
	simulateRequest(r, "https://example.com/items")
	simulateRequest(r, "https://example.com/items?maxItems=10&page=abc")

	//Output:
	// request: maxItems=100 page=""
	// response: [<https://example.com/items?maxItems=100&page=test>; rel="next"]
	// request: maxItems=10 page="abc"
	// response: [<https://example.com/items?maxItems=10&page=test>; rel="next"]
}

func ExampleMiddleware_backwardsCompatible() {
	middleware := pagination.NewMiddleware(
		// provide a function to rewrite legacy URLs to the new format
		pagination.WithBackwardsCompatibility(shim),
	)

	r := chi.NewRouter()
	r.With(middleware).Get("/items", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("request: maxItems=%d page=%q\n", pagination.MaxItems(r), pagination.Page(r))

		pagination.SetNext(r, "test") // dummy example token, replace with your own
		w.Write([]byte("...Data..."))
	})

	// Simulate a request to the handler and print the response headers
	// real applications should use http.ListenAndServe or similar
	simulateRequest(r, "https://example.com/items")
	simulateRequest(r, "https://example.com/items?limit=10&cursor=abc")

	//Output:
	// request: maxItems=100 page=""
	// response: [<https://example.com/items?maxItems=100&page=test>; rel="next"]
	// request: maxItems=10 page="abc"
	// response: [<https://example.com/items?maxItems=10&page=abc>; rel="alternate" <https://example.com/items?maxItems=10&page=test>; rel="next"]
}

func shim(legacy url.URL) (newURL url.URL, updated bool) {
	newURL = legacy

	q := newURL.Query()

	if value := q.Get("cursor"); value != "" {
		q.Del("cursor")
		q.Set("page", value)
		updated = true
	}

	if legacyValue := q.Get("limit"); legacyValue != "" {
		q.Del("limit")
		q.Set("maxItems", legacyValue)
		updated = true
	}

	newURL.RawQuery = q.Encode()

	return newURL, updated
}

func simulateRequest(handler http.Handler, url string) {
	// Simulate a request to the handler and print the response headers
	req, _ := http.NewRequest("GET", url, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Print the link response headers
	// Note: the order of the links is not guaranteed, so we sort them
	values := rec.Result().Header.Values("Link")
	sort.Strings(values)
	fmt.Println("response:", values)
}
