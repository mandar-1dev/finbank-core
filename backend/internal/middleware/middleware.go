package middleware

import "net/http"

// CORS allows the static frontend (served from a different origin during
// local development, e.g. a live-server on :5500) to call the API on :8080.
// This is a local educational simulation with no auth, so an open CORS
// policy is acceptable here.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Idempotency-Key")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Recover turns a panic in any handler into a clean 500 JSON response
// instead of crashing the server or leaking a Go stack trace.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"success":false,"message":"Internal server error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Logger prints a single line per request — enough to follow what the
// simulation is doing without a full logging framework.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println(r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
