package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleStatus(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleStatus)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Validate JSON structure
	var resp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Errorf("handler returned invalid JSON: %v", err)
	}

	expectedKeys := []string{"volume", "brightness", "media", "apps_hash", "text_focused"}
	for _, key := range expectedKeys {
		if _, ok := resp[key]; !ok {
			t.Errorf("JSON response missing key %s", key)
		}
	}
}

func TestMiddleware_NoCookie_ServesPairingPage(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	handler := authMiddleware(mux)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusFound)
	}
	if loc := rr.Header().Get("Location"); loc != "/pair.html" {
		t.Errorf("Expected redirect to /pair.html, got %s", loc)
	}
}

func TestMiddleware_APIRoutesReturn401JSON_NotHTML(t *testing.T) {
	req, _ := http.NewRequest("POST", "/api/mouse", nil)
	rr := httptest.NewRecorder()
	
	mux := http.NewServeMux()
	mux.HandleFunc("/api/mouse", handleMouse)
	
	handler := authMiddleware(mux)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestRequestHandler_AcceptsTraffic(t *testing.T) {
	req, _ := http.NewRequest("POST", "/api/pair/request", nil)
	rr := httptest.NewRecorder()
	
	handler := http.HandlerFunc(handlePairRequest)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected 200 OK, got %v", status)
	}
}
