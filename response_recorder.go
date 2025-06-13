package traefik_lazy_serve

import (
	"net/http"
)

// responseRecorder is a custom ResponseWriter that captures
// the response status code, headers, and body for inspection or replay.
type responseRecorder struct {
	http.ResponseWriter             // Embedded original ResponseWriter
	statusCode          int         // Captured status code
	header              http.Header // Captured headers
	body                []byte      // Captured response body
}

// Header returns the captured response headers.
func (rr *responseRecorder) Header() http.Header {
	return rr.header
}

// WriteHeader captures the response status code.
func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
}

// Write captures the response body data.
func (rr *responseRecorder) Write(b []byte) (int, error) {
	rr.body = b
	return len(b), nil
}

// WriteToOriginal writes the captured headers, status code, and body
// to the original ResponseWriter.
func (rr *responseRecorder) WriteToOriginal(w http.ResponseWriter) {
	// Write captured headers
	for k, vv := range rr.header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	// Write captured status code
	if rr.statusCode != 0 {
		w.WriteHeader(rr.statusCode)
	}

	// Write captured body
	w.Write(rr.body)
}
