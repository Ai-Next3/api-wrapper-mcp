package main

import (
	"log"
	"net/http"
	"os"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("!!!!!! [TEST] RECEIVED REQUEST: Method=%s, Path=%s, Headers=%v", r.Method, r.URL.Path, r.Header)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, world!"))
}

func main() {
	http.HandleFunc("/", helloHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	addr := ":" + port

	log.Printf("Starting DUMB TEST SERVER on http://localhost:%s", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}
