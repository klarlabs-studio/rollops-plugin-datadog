package datadog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func fixedClock() time.Time { return time.Unix(1_700_000_000, 0) }

func TestQuery_ReturnsLatestPoint(t *testing.T) {
	var gotQuery, gotAPI, gotApp string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("query")
		gotAPI = r.Header.Get("DD-API-KEY")
		gotApp = r.Header.Get("DD-APPLICATION-KEY")
		w.Write([]byte(`{"series":[{"pointlist":[[1700000000000,0.01],[1700000060000,0.04]]}]}`))
	}))
	defer srv.Close()

	p := Provider{BaseURL: srv.URL, APIKey: "k", AppKey: "a", HTTP: srv.Client(), now: fixedClock}
	v, err := p.Query(context.Background(), "avg:trace.errors{service:api}.as_rate()")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if v != 0.04 {
		t.Errorf("value = %v, want 0.04 (latest point)", v)
	}
	if gotQuery != "avg:trace.errors{service:api}.as_rate()" {
		t.Errorf("query not forwarded: %q", gotQuery)
	}
	if gotAPI != "k" || gotApp != "a" {
		t.Errorf("auth headers = %q/%q", gotAPI, gotApp)
	}
}

func TestQuery_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"series":[]}`))
	}))
	defer srv.Close()
	p := Provider{BaseURL: srv.URL, APIKey: "k", AppKey: "a", HTTP: srv.Client(), now: fixedClock}
	if _, err := p.Query(context.Background(), "q"); err == nil || !strings.Contains(err.Error(), "no data") {
		t.Fatalf("empty series must error, got %v", err)
	}
}

func TestQuery_DatadogError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"error":"bad query"}`))
	}))
	defer srv.Close()
	p := Provider{BaseURL: srv.URL, APIKey: "k", AppKey: "a", HTTP: srv.Client(), now: fixedClock}
	if _, err := p.Query(context.Background(), "q"); err == nil || !strings.Contains(err.Error(), "bad query") {
		t.Fatalf("datadog error must propagate, got %v", err)
	}
}

func TestQuery_RequiresKeys(t *testing.T) {
	if _, err := (Provider{BaseURL: "http://x"}).Query(context.Background(), "q"); err == nil {
		t.Fatal("missing keys must error")
	}
}

func TestQuery_WindowInRequest(t *testing.T) {
	var from, to string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		from, to = r.URL.Query().Get("from"), r.URL.Query().Get("to")
		w.Write([]byte(`{"series":[{"pointlist":[[1,1.0]]}]}`))
	}))
	defer srv.Close()
	p := Provider{BaseURL: srv.URL, APIKey: "k", AppKey: "a", Window: 10 * time.Minute, HTTP: srv.Client(), now: fixedClock}
	if _, err := p.Query(context.Background(), "q"); err != nil {
		t.Fatal(err)
	}
	// to = 1700000000, from = to - 600.
	if to != "1700000000" || from != "1699999400" {
		t.Errorf("window from/to = %s/%s, want 1699999400/1700000000", from, to)
	}
}
