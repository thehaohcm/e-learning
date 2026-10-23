package push

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// RFC 8291, section 5 (example encryption).
func TestEncryptRFC8291Vector(t *testing.T) {
	asPriv, _ := unb64("yfWPiYE-n46HLnH0KqZOF1fJJU3MYrct3AELtAQ-oRw")
	asKey, err := ecdh.P256().NewPrivateKey(asPriv)
	if err != nil {
		t.Fatal(err)
	}
	salt, _ := unb64("DGv6ra1nlYgDCS1FRnbzlw")
	sub := Subscription{
		P256dh: "BCVxsr7N_eNgVRqvHtD0zTZsEc6-VV-JvLexhqUzORcxaOzi6-AYWXvTBHm4bjyPjs7Vd8pZGH6SRpkNtoIAiw4",
		Auth:   "BTBZMqHH6r4Tts7J_aSIgg",
	}
	got, err := encrypt(sub, []byte("When I grow up, I want to be a watermelon"), asKey, salt)
	if err != nil {
		t.Fatal(err)
	}
	want := "DGv6ra1nlYgDCS1FRnbzlwAAEABBBP4z9KsN6nGRTbVYI_c7VJSPQTBtkgcy27mlmlMoZIIgDll6e3vCYLocInmYWAmS6TlzAC8wEqKK6PBru3jl7A_yl95bQpu6cVPTpK4Mqgkf1CXztLVBSt2Ks3oZwbuwXPXLWyouBWLVWGNWQexSgSxsj_Qulcy4a-fN"
	if b64(got) != want {
		t.Fatalf("ciphertext mismatch\n got %s\nwant %s", b64(got), want)
	}
}

// decrypt is the user agent side, to check Encrypt with random keys.
func decrypt(t *testing.T, uaKey *ecdh.PrivateKey, auth, body []byte) []byte {
	t.Helper()
	salt, idlen := body[:16], int(body[20])
	asPubRaw := body[21 : 21+idlen]
	asPub, err := ecdh.P256().NewPublicKey(asPubRaw)
	if err != nil {
		t.Fatal(err)
	}
	shared, _ := uaKey.ECDH(asPub)
	info := append(append([]byte("WebPush: info\x00"), uaKey.PublicKey().Bytes()...), asPubRaw...)
	ikm, _ := hkdf.Key(sha256.New, shared, auth, string(info), 32)
	cek, _ := hkdf.Key(sha256.New, ikm, salt, "Content-Encoding: aes128gcm\x00", 16)
	nonce, _ := hkdf.Key(sha256.New, ikm, salt, "Content-Encoding: nonce\x00", 12)
	block, _ := aes.NewCipher(cek)
	gcm, _ := cipher.NewGCM(block)
	plain, err := gcm.Open(nil, nonce, body[21+idlen:], nil)
	if err != nil {
		t.Fatal(err)
	}
	if plain[len(plain)-1] != 2 {
		t.Fatalf("missing record delimiter")
	}
	return plain[:len(plain)-1]
}

func TestSendRoundTrip(t *testing.T) {
	uaKey, _ := ecdh.P256().GenerateKey(rand.Reader)
	auth := []byte("0123456789abcdef")
	priv, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	v, err := ParseVAPID(priv, "mailto:test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	var gotBody []byte
	var gotHeader http.Header
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotHeader = r.Header
		if strings.HasSuffix(r.URL.Path, "/gone") {
			w.WriteHeader(http.StatusGone)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	s := &Sender{VAPID: v, Client: srv.Client()}
	sub := Subscription{Endpoint: srv.URL + "/push/abc", P256dh: b64(uaKey.PublicKey().Bytes()), Auth: b64(auth)}
	payload := []byte(`{"title":"Time to study","body":"12 cards due"}`)
	if err := s.Send(context.Background(), sub, payload, time.Hour, "study"); err != nil {
		t.Fatal(err)
	}
	if string(decrypt(t, uaKey, auth, gotBody)) != string(payload) {
		t.Fatal("payload mismatch")
	}
	if gotHeader.Get("Content-Encoding") != "aes128gcm" || gotHeader.Get("TTL") != "3600" || gotHeader.Get("Topic") != "study" {
		t.Fatalf("headers: %v", gotHeader)
	}

	// VAPID header: signature verifies with the advertised public key, claims are right.
	h := gotHeader.Get("Authorization")
	if !strings.HasPrefix(h, "vapid t=") {
		t.Fatalf("auth header %q", h)
	}
	parts := strings.SplitN(strings.TrimPrefix(h, "vapid t="), ", k=", 2)
	if parts[1] != v.PublicKey() {
		t.Fatal("k is not the public key")
	}
	jwt := strings.Split(parts[0], ".")
	digest := sha256.Sum256([]byte(jwt[0] + "." + jwt[1]))
	sig, _ := unb64(jwt[2])
	if len(sig) != 64 || !ecdsa.Verify(&v.Private.PublicKey, digest[:], new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:])) {
		t.Fatal("bad JWT signature")
	}
	rawClaims, _ := unb64(jwt[1])
	var claims struct {
		Aud string `json:"aud"`
		Sub string `json:"sub"`
		Exp int64  `json:"exp"`
	}
	json.Unmarshal(rawClaims, &claims)
	if claims.Aud != srv.URL || claims.Sub != "mailto:test@example.com" || claims.Exp <= time.Now().Unix() {
		t.Fatalf("claims %+v", claims)
	}

	// 410 Gone → ErrGone
	sub.Endpoint = srv.URL + "/gone"
	if err := s.Send(context.Background(), sub, payload, time.Hour, ""); err != ErrGone {
		t.Fatalf("want ErrGone, got %v", err)
	}

	// host guard
	s.AllowHost = KnownPushHost
	if err := s.Send(context.Background(), sub, payload, time.Hour, ""); err == nil || !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("want host refused, got %v", err)
	}
}

func TestParseVAPIDPublicKey(t *testing.T) {
	// RFC 8291 application server key pair: public key derived from the private key.
	v, err := ParseVAPID("yfWPiYE-n46HLnH0KqZOF1fJJU3MYrct3AELtAQ-oRw", "mailto:x@y.z")
	if err != nil {
		t.Fatal(err)
	}
	if v.PublicKey() != "BP4z9KsN6nGRTbVYI_c7VJSPQTBtkgcy27mlmlMoZIIgDll6e3vCYLocInmYWAmS6TlzAC8wEqKK6PBru3jl7A8" {
		t.Fatalf("public key %s", v.PublicKey())
	}
}

func TestKnownPushHost(t *testing.T) {
	for host, ok := range map[string]bool{
		"fcm.googleapis.com":                true,
		"updates.push.services.mozilla.com": true,
		"web.push.apple.com":                true,
		"api.push.apple.com":                true,
		"wns2-par02p.notify.windows.com":    true,
		"evil.com":                          false,
		"fcm.googleapis.com.evil.com":       false,
		"localhost":                         false,
	} {
		if KnownPushHost(host) != ok {
			t.Errorf("%s: want %v", host, ok)
		}
	}
}
