// Package auth handles password hashing, JWT session tokens for the web app
// and opaque bearer tokens for study devices (ESP32).
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const tokenTTL = 30 * 24 * time.Hour

var ErrInvalidToken = errors.New("invalid token")

// ---------------------------------------------------------------- passwords

func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, pw string) bool {
	if hash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// ---------------------------------------------------------------- JWT (web users)

type Claims struct {
	UserID int64  `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type Issuer struct{ secret []byte }

func NewIssuer(secret string) *Issuer { return &Issuer{secret: []byte(secret)} }

func (i *Issuer) Issue(userID int64, role string) (string, time.Time, error) {
	exp := time.Now().Add(tokenTTL)
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprint(userID),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.secret)
	return s, exp, err
}

func (i *Issuer) Parse(token string) (*Claims, error) {
	claims := &Claims{}
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return i.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !t.Valid || claims.UserID == 0 {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ---------------------------------------------------------------- device tokens

// NewDeviceToken returns a random secret for a device and the hash to store.
// Only the hash is kept in the database.
func NewDeviceToken() (token, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	token = "dev_" + hex.EncodeToString(b)
	return token, HashDeviceToken(token), nil
}

func HashDeviceToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

// NewPairCode returns a 6-digit code shown on the device screen.
func NewPairCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
