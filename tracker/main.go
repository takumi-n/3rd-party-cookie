package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"text/template"

	"github.com/gorilla/mux"
)

var trackingData = map[string]map[string]bool{}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ad.js", ad)
	r.HandleFunc("/me", me)

	http.Handle("/", r)
	err := http.ListenAndServeTLS(":9090", "./ssl/tracker.pem", "./ssl/tracker-key.pem", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func ad(w http.ResponseWriter, r *http.Request) {
	identifier := ""

	c, err := r.Cookie("identifier")
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
	} else {
		identifier = c.Value
	}

	_, ok := trackingData[identifier]
	if !ok {
		trackingData[identifier] = map[string]bool{}
	}

	u, err := url.Parse(r.Referer())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if u.Hostname() != "" {
		trackingData[identifier][u.Hostname()] = true
	}

	data := trackingData[identifier]
	adContent := ""

	for site := range data {
		adContent += fmt.Sprintf("<div>%s を閲覧したことがある</div>", site)
	}

	tmpl := template.Must(template.ParseFiles("./ad.tmpl.js"))
	if err := tmpl.Execute(w, adContent); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func me(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("identifier")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No tracking data"))
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
