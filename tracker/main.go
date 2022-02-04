package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

var trackingData = map[string]map[string]bool{}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ad", ad)
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

	trackingData[identifier][u.Hostname()] = true

	data := trackingData[identifier]
	adContent := ""

	for site := range data {
		adContent += fmt.Sprintf("<div>%s を閲覧したことがある</div>", site)
	}

	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "https://localhost:8080")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Write([]byte(fmt.Sprintf(`
const ad = document.getElementById('ad');
ad.style.display = 'flex';
ad.style.flexDirection = 'column';
ad.style.alignItems = 'center';
ad.style.justifyContent = 'center';
ad.style.width = 300;
ad.style.height = 250;
ad.style.border = 'red solid 1px';
ad.innerHTML = '%s';
	`, adContent)))
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
