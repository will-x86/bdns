package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendQuery_RoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/dns-message" {
			t.Errorf("content-type = %q, want application/dns-message", ct)
		}
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/dns-message")
		_, _ = w.Write(body) // echo the query back as the "response"
	}))
	defer srv.Close()

	query := []byte{0x12, 0x34, 0x01, 0x00, 0x00, 0x01}
	resp, err := NewDoHClient(srv.URL).SendQuery(query)
	if err != nil {
		t.Fatalf("SendQuery: %v", err)
	}
	if string(resp) != string(query) {
		t.Errorf("resp = %x, want %x", resp, query)
	}
}

func TestSendQuery_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	if _, err := NewDoHClient(srv.URL).SendQuery([]byte{0, 0}); err == nil {
		t.Fatal("want error on non-200 response")
	}
}

func TestSendQuery_Unreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // nothing is listening now

	if _, err := NewDoHClient(url).SendQuery([]byte{0, 0}); err == nil {
		t.Fatal("want error when endpoint is unreachable")
	}
}
