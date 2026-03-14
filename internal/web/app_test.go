package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"equinox/internal/demo"
)

func TestIndexRendersAndRoutesOrder(t *testing.T) {
	snapshot, err := demo.LoadFixtureSnapshot()
	if err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
	app, err := New(snapshot)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}

	form := url.Values{
		"cluster": {"prop-001"},
		"side":    {"buy_yes"},
		"limit":   {"0.60"},
		"size":    {"1000"},
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Cross-venue event clustering") {
		t.Fatalf("expected page title in body")
	}
	if !strings.Contains(body, "Polymarket") {
		t.Fatalf("expected routed venue in body")
	}
}
