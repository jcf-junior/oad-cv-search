package handler

import "net/http"

func NewRouter(candidates *CandidateHandler, frontendDir string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /candidates", candidates.Upload)
	mux.HandleFunc("GET /candidates/count", candidates.Count)
	mux.HandleFunc("GET /candidates/{id}", candidates.Get)
	mux.HandleFunc("POST /search", candidates.Search)
	mux.HandleFunc("POST /search/keyword", candidates.SearchKeyword)
	mux.Handle("GET /", http.FileServer(http.Dir(frontendDir)))

	return cors(mux)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
