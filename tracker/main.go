package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var trackingData = map[string]map[string]string{}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/track", track)
	r.HandleFunc("/me", me)

	http.Handle("/", r)
	err := http.ListenAndServeTLS(":9090", "./tracker.pem", "./tracker-key.pem", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func track(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("identifier")

	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "https://localhost:8080")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	identifier := ""

	if err != nil {
		identifier = makeRandomStr(10)
		cookie := &http.Cookie{
			Name:     "identifier",
			Value:    identifier,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		}
		http.SetCookie(w, cookie)
		w.Write([]byte(fmt.Sprintf("You are a new user: %v", identifier)))
	} else {
		identifier = c.Value
		w.Write([]byte(fmt.Sprintf("You are an existing user: %v", identifier)))
	}

	_, ok := trackingData[identifier]
	if !ok {
		trackingData[identifier] = map[string]string{}
	}

	params := r.URL.Query()
	for k, v := range params {
		first := v[0]
		trackingData[identifier][k] = first
	}
}

func me(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("identifier")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("We've not tracked you"))
		return
	}

	identifier := c.Value
	data := trackingData[identifier]

	w.Write([]byte(fmt.Sprintf("%v", data)))
}

func makeRandomStr(digit uint32) string {
	b := make([]byte, digit)
	if _, err := rand.Read(b); err != nil {
		return ""
	}

	var result string
	for _, v := range b {
		result += string(v%byte(94) + 33)
	}
	return result
}
