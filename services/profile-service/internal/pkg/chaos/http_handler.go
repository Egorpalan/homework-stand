package chaos

import (
	"net/http"

	"profile-service/internal/pkg/chaos/mode"
)

func ModeSetHandler(store *mode.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Query().Get("mode")
		method := r.URL.Query().Get("method")

		md := mode.Mode(m)

		switch md {
		case mode.OK, mode.Slow, mode.Error, mode.Flaky, mode.RareError, mode.Overload:
			if method == "" {
				store.SetGlobal(md)
				w.Write([]byte("global mode set to " + m))
				return
			}

			store.SetForMethod(method, md)
			w.Write([]byte("mode for " + method + " set to " + m))

		default:
			http.Error(w, "unknown mode", http.StatusBadRequest)
		}
	}
}
