package main

import "net/http"

func main() {
	mux := http.NewServeMux()
	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("GET /app/", fileServer)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	server.ListenAndServe()

}
