package pagination

import (
	"net/http"
	"sync"
)

type responseWriter struct {
	http.ResponseWriter
	once  sync.Once
	state *state
}

func (r *responseWriter) Write(b []byte) (int, error) {
	r.writePageHeaders()
	return r.ResponseWriter.Write(b)
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.writePageHeaders()
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseWriter) writePageHeaders() {
	r.once.Do(func() {
		header := r.ResponseWriter.Header()

		if r.state.wasRewritten {
			r.state.links["alternate"] = r.state.current
			header.Add("Warning", `299 - "Deprecated pagination method. Please use alternate method."`)
		}

		for k, v := range r.state.links {
			header.Add("Link", `<`+v.String()+`>; rel="`+k+`"`)
		}
	})
}

func newResponseWriter(w http.ResponseWriter, s *state) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		state:          s,
	}
}
