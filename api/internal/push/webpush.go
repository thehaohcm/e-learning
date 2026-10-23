// Package push sends Web Push notifications (RFC 8030) with VAPID
// authentication (RFC 8292) and aes128gcm payload encryption (RFC 8188/8291),
// using only the Go standard library.
package push

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Subscription is what the browser's PushManager.subscribe() returns.
type Subscription struct {
	Endpoint string
	P256dh   string // base64url, uncompressed P-256 point (65 bytes)
	Auth     string // base64url, 16 bytes
}

// VAPID holds the application server key pair.
type VAPID struct {
	Private *ecdsa.PrivateKey
	Subject string // "mailto:..." or "https://..."
}

// GenerateKey creates a new VAPID private key, returned base64url (32 bytes).
func GenerateKey() (string, error) {
	k, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return "", err
	}
	return b64(k.Bytes()), nil
}

// ParseVAPID loads a base64url private key (32 bytes, as made by GenerateKey).
func ParseVAPID(privateB64, subject string) (*VAPID, error) {
	raw, err := unb64(privateB64)
	if err != nil {
		return nil, fmt.Errorf("vapid key: %w", err)
	}
	ek, err := ecdh.P256().NewPrivateKey(raw)
	if err != nil {
		return nil, fmt.Errorf("vapid key: %w", err)
	}
	pub := ek.PublicKey().Bytes() // 0x04 || X || Y
	key := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{Curve: elliptic.P256(), X: new(big.Int).SetBytes(pub[1:33]), Y: new(big.Int).SetBytes(pub[33:])},
		D:         new(big.Int).SetBytes(raw),
	}
	return &VAPID{Private: key, Subject: subject}, nil
}

// PublicKey is the applicationServerKey the browser needs (base64url, 65 bytes).
func (v *VAPID) PublicKey() string {
	return b64(publicBytes(&v.Private.PublicKey))
}

func publicBytes(p *ecdsa.PublicKey) []byte {
	out := make([]byte, 65)
	out[0] = 4
	p.X.FillBytes(out[1:33])
	p.Y.FillBytes(out[33:])
	return out
}

// authHeader builds "vapid t=<jwt>, k=<public key>" for the push service origin.
func (v *VAPID) authHeader(endpoint string, now time.Time) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	header := b64([]byte(`{"typ":"JWT","alg":"ES256"}`))
	claims, _ := json.Marshal(map[string]any{
		"aud": u.Scheme + "://" + u.Host,
		"exp": now.Add(12 * time.Hour).Unix(),
		"sub": v.Subject,
	})
	signing := header + "." + b64(claims)
	digest := sha256.Sum256([]byte(signing))
	r, s, err := ecdsa.Sign(rand.Reader, v.Private, digest[:])
	if err != nil {
		return "", err
	}
	sig := make([]byte, 64) // JWS ES256: fixed-size r || s, not ASN.1
	r.FillBytes(sig[:32])
	s.FillBytes(sig[32:])
	return "vapid t=" + signing + "." + b64(sig) + ", k=" + v.PublicKey(), nil
}

// Encrypt encrypts a payload for one subscription (RFC 8291, aes128gcm).
func Encrypt(sub Subscription, payload []byte) ([]byte, error) {
	asKey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return encrypt(sub, payload, asKey, salt)
}

func encrypt(sub Subscription, payload []byte, asKey *ecdh.PrivateKey, salt []byte) ([]byte, error) {
	uaRaw, err := unb64(sub.P256dh)
	if err != nil {
		return nil, fmt.Errorf("p256dh: %w", err)
	}
	uaPub, err := ecdh.P256().NewPublicKey(uaRaw)
	if err != nil {
		return nil, fmt.Errorf("p256dh: %w", err)
	}
	authSecret, err := unb64(sub.Auth)
	if err != nil || len(authSecret) < 16 {
		return nil, errors.New("auth: invalid secret")
	}
	shared, err := asKey.ECDH(uaPub)
	if err != nil {
		return nil, err
	}
	asPub := asKey.PublicKey().Bytes()

	// IKM = HKDF(auth_secret, ecdh_secret, "WebPush: info" || 0 || ua_public || as_public, 32)
	info := append(append([]byte("WebPush: info\x00"), uaRaw...), asPub...)
	ikm, err := hkdf.Key(sha256.New, shared, authSecret, string(info), 32)
	if err != nil {
		return nil, err
	}
	cek, err := hkdf.Key(sha256.New, ikm, salt, "Content-Encoding: aes128gcm\x00", 16)
	if err != nil {
		return nil, err
	}
	nonce, err := hkdf.Key(sha256.New, ikm, salt, "Content-Encoding: nonce\x00", 12)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(cek)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	// single record: payload || 0x02 (last-record delimiter), no padding
	plain := append(append([]byte{}, payload...), 2)
	const recordSize = 4096
	if len(plain)+16 > recordSize {
		return nil, errors.New("payload too large")
	}
	var head bytes.Buffer
	head.Write(salt)
	binary.Write(&head, binary.BigEndian, uint32(recordSize))
	head.WriteByte(byte(len(asPub)))
	head.Write(asPub)
	return gcm.Seal(head.Bytes(), nonce, plain, nil), nil
}

// ErrGone: the subscription expired or was revoked; delete it.
var ErrGone = errors.New("push subscription gone")

// Sender delivers notifications.
type Sender struct {
	VAPID  *VAPID
	Client *http.Client
	// AllowHost decides which push service hosts may be contacted (SSRF guard).
	AllowHost func(host string) bool
}

// Send encrypts and posts one message. ttl is how long the push service keeps
// it if the device is offline; topic (optional, ≤32 url-safe chars) makes a
// newer message replace an undelivered older one.
func (s *Sender) Send(ctx context.Context, sub Subscription, payload []byte, ttl time.Duration, topic string) error {
	u, err := url.Parse(sub.Endpoint)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return fmt.Errorf("invalid endpoint")
	}
	if s.AllowHost != nil && !s.AllowHost(u.Hostname()) {
		return fmt.Errorf("push host not allowed: %s", u.Hostname())
	}
	body, err := Encrypt(sub, payload)
	if err != nil {
		return err
	}
	auth, err := s.VAPID.authHeader(sub.Endpoint, time.Now())
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sub.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Encoding", "aes128gcm")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("TTL", fmt.Sprint(int(ttl.Seconds())))
	req.Header.Set("Urgency", "normal")
	if topic != "" {
		req.Header.Set("Topic", topic)
	}
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return nil
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		return ErrGone
	default:
		return fmt.Errorf("push service %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}
}

// KnownPushHost reports whether host belongs to a browser push service
// (Chrome/Edge/Firefox/Safari). Subscriptions to other hosts are refused so
// the server cannot be used to post to arbitrary URLs.
func KnownPushHost(host string) bool {
	host = strings.ToLower(host)
	for _, h := range []string{
		"fcm.googleapis.com", "android.googleapis.com", // Chrome, Android
		"updates.push.services.mozilla.com", // Firefox
		"web.push.apple.com",                // Safari / iOS
	} {
		if host == h {
			return true
		}
	}
	for _, suffix := range []string{".push.apple.com", ".notify.windows.com", ".push.services.mozilla.com"} {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}
	return false
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func unb64(s string) ([]byte, error) {
	s = strings.TrimRight(strings.TrimSpace(s), "=")
	s = strings.NewReplacer("+", "-", "/", "_").Replace(s)
	return base64.RawURLEncoding.DecodeString(s)
}
