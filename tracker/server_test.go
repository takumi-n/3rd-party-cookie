package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestTrackerService_HandleAd(t *testing.T) {
	// Setup
	tracker := NewTrackerService()

	// 1. First Access (No Cookie, Referer: client1)
	req1 := httptest.NewRequest("GET", "/ad.js", nil)
	req1.Header.Set("Referer", "https://client1:9091/")
	w1 := httptest.NewRecorder()

	tracker.HandleAd(w1, req1)

	resp1 := w1.Result()
	cookies1 := resp1.Cookies()

	// Check Cookie creation
	var identifier string
	found := false
	for _, c := range cookies1 {
		if c.Name == "identifier" {
			identifier = c.Value
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Expected identifier cookie to be set")
	}

	// Check response content (Should contain client1 visit record)
	body1 := w1.Body.String()
	if !strings.Contains(body1, "client1") {
		t.Errorf("Expected response to contain 'client1', got %s", body1)
	}

	// 2. Second Access (With Cookie, Referer: camera page)
	req2 := httptest.NewRequest("GET", "/ad.js", nil)
	req2.Header.Set("Referer", "https://client1:9091/products/camera.html")
	req2.AddCookie(&http.Cookie{Name: "identifier", Value: identifier})
	w2 := httptest.NewRecorder()

	tracker.HandleAd(w2, req2)

	// Check response content (Should contain interest:camera ad)
	body2 := w2.Body.String()
	if !strings.Contains(body2, "interest:camera") && !strings.Contains(body2, "最新の camera") {
		t.Errorf("Expected response to contain camera ad, got %s", body2)
	}

	// 3. Third Access from different site (client2)
	req3 := httptest.NewRequest("GET", "/ad.js", nil)
	req3.Header.Set("Referer", "https://client2:9092/")
	req3.AddCookie(&http.Cookie{Name: "identifier", Value: identifier})
	w3 := httptest.NewRecorder()

	tracker.HandleAd(w3, req3)

	body3 := w3.Body.String()
	// Should show history of client1 AND camera interest
	if !strings.Contains(body3, "client1") {
		t.Error("Expected response to show client1 history")
	}
	if !strings.Contains(body3, "最新の camera") {
		t.Error("Expected response to show camera ad")
	}
}

func TestTrackerService_RecordAccess(t *testing.T) {
	tracker := NewTrackerService()
	id := "test-user-1"

	// Test Case 1: Normal host visit
	u1, _ := url.Parse("https://example.com/page")
	tracker.recordAccess(id, u1)

	tracker.mu.RLock()
	if !tracker.trackingData[id]["example.com"] {
		t.Error("Failed to record hostname example.com")
	}
	tracker.mu.RUnlock()

	// Test Case 2: Interest visit (camera)
	u2, _ := url.Parse("https://shop.com/products/camera.html")
	tracker.recordAccess(id, u2)

	tracker.mu.RLock()
	if !tracker.trackingData[id]["interest:camera"] {
		t.Error("Failed to record camera interest")
	}
	if !tracker.trackingData[id]["shop.com"] {
		t.Error("Failed to record hostname shop.com")
	}
	tracker.mu.RUnlock()
}
