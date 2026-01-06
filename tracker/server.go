package main

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"text/template"
)

// TrackerService holds the tracking data and logic
type TrackerService struct {
	// trackingData map[identifier]map[category/site]bool
	trackingData map[string]map[string]bool
	mu           sync.RWMutex
	tmpl         *template.Template
}

// NewTrackerService creates a new instance of TrackerService
func NewTrackerService() *TrackerService {
	return &TrackerService{
		trackingData: make(map[string]map[string]bool),
		tmpl:         template.Must(template.ParseFiles("./ad.tmpl.js")),
	}
}

// HandleAd handles the /ad.js request
func (s *TrackerService) HandleAd(w http.ResponseWriter, r *http.Request) {
	identifier := s.getOrSetIdentifier(w, r)
	
	// Parse Referer
	referer := r.Referer()
	if referer == "" {
		// If no referer, just return current ads (or nothing)
		s.renderAd(w, identifier)
		return
	}

	u, err := url.Parse(referer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Record tracking info
	s.recordAccess(identifier, u)

	// Render response
	s.renderAd(w, identifier)
}

// HandleMe handles the /me request
func (s *TrackerService) HandleMe(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("identifier")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No tracking data"))
		return
	}

	identifier := c.Value
	s.mu.RLock()
	data := s.trackingData[identifier]
	s.mu.RUnlock()

	w.Write([]byte(fmt.Sprintf("%v", data)))
}

func (s *TrackerService) getOrSetIdentifier(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie("identifier")
	if err == nil {
		return c.Value
	}

	identifier := makeRandomStr(10)
	cookie := &http.Cookie{
		Name:     "identifier",
		Value:    identifier,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, cookie)
	return identifier
}

func (s *TrackerService) recordAccess(identifier string, u *url.URL) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.trackingData[identifier]; !ok {
		s.trackingData[identifier] = map[string]bool{}
	}

	if u.Hostname() != "" {
		s.trackingData[identifier][u.Hostname()] = true
	}

	if strings.Contains(u.Path, "camera") {
		s.trackingData[identifier]["interest:camera"] = true
	}
	if strings.Contains(u.Path, "pc") {
		s.trackingData[identifier]["interest:pc"] = true
	}
}

func (s *TrackerService) renderAd(w http.ResponseWriter, identifier string) {
	s.mu.RLock()
	data := s.trackingData[identifier]
	// Copy data to avoid holding lock during template execution (though simple here)
	// For this simple case, we generate the string inside lock or copy keys.
	// Let's generate adContent string here.
	adContent := ""
	for key := range data {
		if strings.HasPrefix(key, "interest:") {
			category := strings.TrimPrefix(key, "interest:")
			adContent += fmt.Sprintf("<div style='background-color: #e0f7fa; padding: 10px; margin: 5px; border: 1px solid #00acc1;'><strong>おすすめ:</strong> 最新の %s をチェック！</div>", category)
		} else {
			adContent += fmt.Sprintf("<div>%s を閲覧したことがある</div>", key)
		}
	}
	s.mu.RUnlock()

	w.Header().Add("Content-Type", "text/javascript")
	if err := s.tmpl.Execute(w, adContent); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
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
