package account

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/auth" {
			t.Errorf("Expected to request '/v2/auth', got: %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json header, got: %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer eip191:J9k/ZYu9QOxLZGV8l4guRkYd30wZ1YCeM+5bLt3/3FQgjEB26qRijQi+yF15nDLxW/qP5yBnlFFp5D5uAS+ADwA=" {
			t.Errorf("Expected Authorization: Bearer eip191:J9k/ZYu9QOxLZGV8l4guRkYd30wZ1YCeM+5bLt3/3FQgjEB26qRijQi+yF15nDLxW/qP5yBnlFFp5D5uAS+ADwA= header, got: %s", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":{"token": {"secret": "farcaster access token"}}}`))
	}))
	defer server.Close()

	// This is the mnemonic from https://replit.com/@VarunSrinivasa4/Merkle-V2-Custody-Bearer-Token-Example?v=1#index.js
	account := NewAccountService(server.URL, "spare trash wide forest stand solution donate wonder mixed crisp busy silent")
	account.clock = func() time.Time {
		return time.Date(2022, time.March, 11, 8, 0, 0, 0, time.UTC)
	}
	value, _ := account.GetAccessToken(3600)
	if value != "farcaster access token" {
		t.Errorf("Expected 'farcaster access token', got %s", value)
	}
}
