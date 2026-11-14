package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestServer(botURL string) *server {
	return &server{
		cfg:     config{internalKey: "test-key", botURL: botURL, botKey: "test-key", botTimeout: 2 * time.Second},
		store:   newPairStore(),
		limiter: newRateLimiter(5, time.Minute),
		bot:     newBotClient(botURL, "test-key", 2*time.Second, "https://whitelist.example.com", "https://sub.example.com"),
	}
}

func doPost(t *testing.T, handler func(http.ResponseWriter, *http.Request), path, body, remoteAddr string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.RemoteAddr = remoteAddr
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func doRedeem(t *testing.T, srv *server, code, ip string) *httptest.ResponseRecorder {
	t.Helper()
	return doPost(t, srv.handleRedeem, "/pair/redeem", `{"code":"`+code+`"}`, ip, nil)
}

func botServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Key") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

type createResp struct {
	Code      string `json:"code"`
	ExpiresIn int    `json:"expires_in"`
}

type redeemResp struct {
	TelegramID    int64          `json:"telegram_id"`
	Subscriptions []subscription `json:"subscriptions"`
}

func TestCreateSuccess(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv.handleCreate, "/pair/create", `{"telegram_id":123456,"ttl":300}`, "127.0.0.1:1234", map[string]string{"X-Internal-Key": "test-key"})
	if rec.Code != http.StatusOK {
		t.Fatalf("create status = %d, want 200, body %s", rec.Code, rec.Body.String())
	}
	var resp createResp
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if !isSixDigits(resp.Code) {
		t.Fatalf("code = %q, want 6 digits", resp.Code)
	}
	if resp.ExpiresIn != 300 {
		t.Fatalf("expires_in = %d, want 300", resp.ExpiresIn)
	}
}

func TestCreateStringTelegramID(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv.handleCreate, "/pair/create", `{"telegram_id":"789"}`, "127.0.0.1:1234", map[string]string{"X-Internal-Key": "test-key"})
	if rec.Code != http.StatusOK {
		t.Fatalf("create status = %d, want 200, body %s", rec.Code, rec.Body.String())
	}
}

func TestCreateNoKey(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv.handleCreate, "/pair/create", `{"telegram_id":1}`, "127.0.0.1:1234", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("create status = %d, want 401", rec.Code)
	}
}

func TestCreateInvalidBody(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv.handleCreate, "/pair/create", `{"telegram_id":0}`, "127.0.0.1:1234", map[string]string{"X-Internal-Key": "test-key"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("create status = %d, want 400", rec.Code)
	}
}

func TestRedeemSuccessTwoKeys(t *testing.T) {
	bot := botServer(t, http.StatusOK, `{"keys":[{"shortUuid":"uuid-wh","type":"_wh"},{"shortUuid":"uuid-basic","type":"_basic"}]}`)
	defer bot.Close()
	srv := newTestServer(bot.URL)
	code, err := srv.store.create(123, 300*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	rec := doRedeem(t, srv, code, "10.0.0.1:1234")
	if rec.Code != http.StatusOK {
		t.Fatalf("redeem status = %d, want 200, body %s", rec.Code, rec.Body.String())
	}
	var resp redeemResp
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.TelegramID != 123 {
		t.Fatalf("telegram_id = %d, want 123", resp.TelegramID)
	}
	if len(resp.Subscriptions) != 2 {
		t.Fatalf("subscriptions = %d, want 2", len(resp.Subscriptions))
	}
	if resp.Subscriptions[0].URL != "https://whitelist.example.com/uuid-wh" {
		t.Fatalf("wh url = %q", resp.Subscriptions[0].URL)
	}
	if resp.Subscriptions[1].URL != "https://sub.example.com/uuid-basic" {
		t.Fatalf("basic url = %q", resp.Subscriptions[1].URL)
	}
	if resp.Subscriptions[0].Label != "Обход" || resp.Subscriptions[1].Label != "Базовый" {
		t.Fatalf("labels = %+v, want Обход/Базовый", resp.Subscriptions)
	}
}

func TestRedeemSuccessOneKey(t *testing.T) {
	bot := botServer(t, http.StatusOK, `{"keys":[{"shortUuid":"uuid-basic","type":"_basic"}]}`)
	defer bot.Close()
	srv := newTestServer(bot.URL)
	code, _ := srv.store.create(5, 300*time.Second)
	rec := doRedeem(t, srv, code, "10.0.0.2:1234")
	if rec.Code != http.StatusOK {
		t.Fatalf("redeem status = %d, want 200", rec.Code)
	}
	var resp redeemResp
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Subscriptions) != 1 || resp.Subscriptions[0].ID != "_basic" {
		t.Fatalf("subscriptions = %+v, want one _basic", resp.Subscriptions)
	}
}

func TestRedeemExpired(t *testing.T) {
	bot := botServer(t, http.StatusOK, `{"keys":[{"shortUuid":"x","type":"_basic"}]}`)
	defer bot.Close()
	srv := newTestServer(bot.URL)
	code, _ := srv.store.create(1, -time.Second)
	rec := doRedeem(t, srv, code, "10.0.0.3:1234")
	if rec.Code != http.StatusGone {
		t.Fatalf("redeem status = %d, want 410", rec.Code)
	}
}

func TestRedeemReuse(t *testing.T) {
	bot := botServer(t, http.StatusOK, `{"keys":[{"shortUuid":"x","type":"_basic"}]}`)
	defer bot.Close()
	srv := newTestServer(bot.URL)
	code, _ := srv.store.create(1, 300*time.Second)
	if rec := doRedeem(t, srv, code, "10.0.0.4:1234"); rec.Code != http.StatusOK {
		t.Fatalf("first redeem = %d, want 200", rec.Code)
	}
	if rec := doRedeem(t, srv, code, "10.0.0.4:1234"); rec.Code != http.StatusNotFound {
		t.Fatalf("second redeem = %d, want 404", rec.Code)
	}
}

func TestRedeemNotFound(t *testing.T) {
	srv := newTestServer("")
	rec := doRedeem(t, srv, "000000", "10.0.0.5:1234")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("redeem status = %d, want 404", rec.Code)
	}
}

func TestRedeemNoSubscriptionsKeepsCode(t *testing.T) {
	bot := botServer(t, http.StatusOK, `{"keys":[]}`)
	defer bot.Close()
	srv := newTestServer(bot.URL)
	code, _ := srv.store.create(1, 300*time.Second)
	rec := doRedeem(t, srv, code, "10.0.0.6:1234")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("redeem status = %d, want 404", rec.Code)
	}
	if _, st := srv.store.lookup(code); st != lookupOK {
		t.Fatalf("code burned after one no-subscriptions redeem, status = %d", st)
	}
}

func TestRedeemBotUnavailableKeepsCode(t *testing.T) {
	bot := botServer(t, http.StatusOK, `{"keys":[{"shortUuid":"x","type":"_basic"}]}`)
	url := bot.URL
	bot.Close()
	srv := newTestServer(url)
	code, _ := srv.store.create(1, 300*time.Second)
	rec := doRedeem(t, srv, code, "10.0.0.7:1234")
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("redeem status = %d, want 502", rec.Code)
	}
	if _, st := srv.store.lookup(code); st != lookupOK {
		t.Fatalf("code burned on upstream failure, status = %d", st)
	}
}

func TestRedeemBot500KeepsCode(t *testing.T) {
	bot := botServer(t, http.StatusInternalServerError, `{"error":"boom"}`)
	defer bot.Close()
	srv := newTestServer(bot.URL)
	code, _ := srv.store.create(1, 300*time.Second)
	rec := doRedeem(t, srv, code, "10.0.0.8:1234")
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("redeem status = %d, want 502", rec.Code)
	}
	if _, st := srv.store.lookup(code); st != lookupOK {
		t.Fatalf("code burned on upstream 500, status = %d", st)
	}
}

func TestRedeemRateLimit(t *testing.T) {
	bot := botServer(t, http.StatusOK, `{"keys":[{"shortUuid":"x","type":"_basic"}]}`)
	defer bot.Close()
	srv := newTestServer(bot.URL)
	for i := 0; i < 6; i++ {
		code, _ := srv.store.create(int64(i+1), 300*time.Second)
		rec := doRedeem(t, srv, code, "192.168.1.10:1234")
		if i < 5 && rec.Code != http.StatusOK {
			t.Fatalf("attempt %d status = %d, want 200", i+1, rec.Code)
		}
		if i == 5 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("attempt 6 status = %d, want 429", rec.Code)
		}
	}
}

func TestRedeemBurnAfterFailedAttempts(t *testing.T) {
	bot := botServer(t, http.StatusOK, `{"keys":[]}`)
	defer bot.Close()
	srv := newTestServer(bot.URL)
	code, _ := srv.store.create(1, 300*time.Second)
	for i := 0; i < 5; i++ {
		rec := doRedeem(t, srv, code, "10.0.0.9:1234")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("attempt %d status = %d, want 404", i+1, rec.Code)
		}
	}
	if _, st := srv.store.lookup(code); st != lookupNotFound {
		t.Fatalf("code not burned after 5 failures, status = %d", st)
	}
}

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handleHealth(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("health status = %d, want 200", rec.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("health status = %q, want ok", resp["status"])
	}
}

func TestMapSubscriptions(t *testing.T) {
	keys := []botKey{{ShortUUID: "a1", Type: "_wh"}, {ShortUUID: "b2", Type: "_basic"}, {ShortUUID: "c3", Type: "unknown"}}
	subs := mapSubscriptions(keys, "https://whitelist.example.com", "https://sub.example.com")
	if len(subs) != 2 {
		t.Fatalf("mapped = %d, want 2", len(subs))
	}
	if subs[0].URL != "https://whitelist.example.com/a1" || subs[1].URL != "https://sub.example.com/b2" {
		t.Fatalf("mapped urls = %+v", subs)
	}
	if subs[0].Label != "Обход" || subs[1].Label != "Базовый" {
		t.Fatalf("mapped labels = %+v, want Обход/Базовый", subs)
	}
}

func TestGenerateCodeFormat(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 200; i++ {
		code, err := generateCode()
		if err != nil {
			t.Fatal(err)
		}
		if !isSixDigits(code) {
			t.Fatalf("code = %q, want 6 digits", code)
		}
		seen[code] = true
	}
	if len(seen) < 190 {
		t.Fatalf("unique codes = %d, want >= 190", len(seen))
	}
}

func TestStoreCleanupExpired(t *testing.T) {
	s := newPairStore()
	expired, _ := s.create(1, -time.Second)
	valid, _ := s.create(2, 300*time.Second)
	s.cleanupExpired()
	if _, st := s.lookup(expired); st != lookupNotFound {
		t.Fatalf("expired code still present, status = %d", st)
	}
	if _, st := s.lookup(valid); st != lookupOK {
		t.Fatalf("valid code missing, status = %d", st)
	}
}
