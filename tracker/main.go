package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	tracker := NewTrackerService()

	r := mux.NewRouter()
	r.HandleFunc("/ad.js", tracker.HandleAd)
	r.HandleFunc("/me", tracker.HandleMe)

	http.Handle("/", r)
	err := http.ListenAndServeTLS(":9090", "./ssl/tracker.pem", "./ssl/tracker-key.pem", nil)
	if err != nil {
		log.Fatalln(err)
	}
}
