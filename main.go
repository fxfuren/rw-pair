package main

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultTTLSeconds = 300
const maxFailedAttempts = 5

type config struct {
	addr        string
	internalKey string
	botURL      string
	botKey      string
	botTimeout  time.Duration
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func loadConfig() (config, error) {
	cfg := config{
		addr:        getenv("PAIR_ADDR", ":8080"),
		internalKey: os.Getenv("PAIR_INTERNAL_KEY"),
		botURL:      os.Getenv("BOT_SUBSCRIPTIONS_URL"),
		botKey:      os.Getenv("BOT_INTERNAL_KEY"),
		botTimeout:  5 * time.Second,
	}
	if cfg.internalKey == "" {
		return cfg, errors.New("PAIR_INTERNAL_KEY is not set")
	}
	if cfg.botURL == "" {
		return cfg, errors.New("BOT_SUBSCRIPTIONS_URL is not set")
	}
	if cfg.botKey == "" {
		cfg.botKey = cfg.internalKey
	}
	if v := os.Getenv("BOT_TIMEOUT_SECONDS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return cfg, fmt.Errorf("invalid BOT_TIMEOUT_SECONDS: %q", v)
		}
		cfg.botTimeout = time.Duration(n) * time.Second
	}
	return cfg, nil
}

type flexibleID int64

func (f *flexibleID) UnmarshalJSON(b []byte) error {
	var n int64
	if err := json.Unmarshal(b, &n); err == nil {
		*f = flexibleID(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return err
	}
	*f = flexibleID(n)
	return nil
}

func (f flexibleID) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(f))
}

type server struct {
	cfg     config
	store   *pairStore
	limiter *rateLimiter
	bot     *botClient
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func isInternalSource(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate()
}

func (s *server) checkInternalKey(r *http.Request) bool {
	provided := r.Header.Get("X-Internal-Key")
	if provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(s.cfg.internalKey)) == 1
}

type createRequest struct {
	TelegramID flexibleID `json:"telegram_id"`
	TTL        int        `json:"ttl"`
}

func (s *server) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !isInternalSource(r) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if !s.checkInternalKey(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if int64(req.TelegramID) <= 0 {
		writeError(w, http.StatusBadRequest, "telegram_id is required")
		return
	}
	ttlSeconds := req.TTL
	if ttlSeconds == 0 {
		ttlSeconds = defaultTTLSeconds
	}
	if ttlSeconds < 0 || ttlSeconds > 3600 {
		writeError(w, http.StatusBadRequest, "ttl out of range")
		return
	}
	code, err := s.store.create(int64(req.TelegramID), time.Duration(ttlSeconds)*time.Second)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create code")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"code": code, "expires_in": ttlSeconds})
}

func isSixDigits(s string) bool {
	if len(s) != 6 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

type redeemRequest struct {
	Code string `json:"code"`
}

func (s *server) handleRedeem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ip := clientIP(r)
	if !s.limiter.allow(ip) {
		w.Header().Set("Retry-After", "60")
		writeError(w, http.StatusTooManyRequests, "too many attempts")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req redeemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if !isSixDigits(req.Code) {
		writeError(w, http.StatusBadRequest, "invalid code")
		return
	}
	entry, status := s.store.lookup(req.Code)
	switch status {
	case lookupNotFound:
		writeError(w, http.StatusNotFound, "code not found")
		return
	case lookupExpired:
		writeError(w, http.StatusGone, "code expired")
		return
	}
	subs, err := s.bot.fetchSubscriptions(r.Context(), entry.telegramID)
	if err != nil {
		if errors.Is(err, errNoSubscriptions) {
			s.store.recordFailure(req.Code, maxFailedAttempts)
			writeError(w, http.StatusNotFound, "no subscriptions")
			return
		}
		writeError(w, http.StatusBadGateway, "upstream unavailable")
		return
	}
	s.store.delete(req.Code)
	writeJSON(w, http.StatusOK, map[string]any{"telegram_id": entry.telegramID, "subscriptions": subs})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	store := newPairStore()
	stop := make(chan struct{})
	go store.runCleanup(time.Minute, stop)
	srv := &server{
		cfg:     cfg,
		store:   store,
		limiter: newRateLimiter(5, time.Minute),
		bot:     newBotClient(cfg.botURL, cfg.botKey, cfg.botTimeout),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/pair/create", srv.handleCreate)
	mux.HandleFunc("/pair/redeem", srv.handleRedeem)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/pair/health", handleHealth)
	httpSrv := &http.Server{
		Addr:              cfg.addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	log.Printf("rw-pair listening on %s", cfg.addr)
	log.Fatal(httpSrv.ListenAndServe())
}
