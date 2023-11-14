package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/hkdf"
)

type contextKey string

const RedirectUrlContextKey contextKey = "redirectUrl"
const stateMaxAge = time.Minute * 15
const pkceMaxAge = time.Minute * 15

func GenerateState(redirectUrl string) (string, *http.Cookie) {
	state := generateRandomString()

	return state, CreateSignedCookie(StateCookieName, state, time.Now().Add(stateMaxAge), map[string]interface{}{
		"redirectUrl": redirectUrl,
	})
}

func ValidateState(w http.ResponseWriter, r *http.Request) bool {

	stateFromAuthServer := r.URL.Query().Get("state")

	defer http.SetCookie(w, DeleteCookie(StateCookieName))

	claims, err := DecodeSignedCookie(StateCookieName, r)

	if err != nil {
		fmt.Println("err token parse", err)
		return false
	}

	// handle expired token
	exp, ok := claims["exp"].(float64)

	if !ok {
		fmt.Println("exp error:", exp)
		return false
	}

	if time.Now().After(time.Unix(int64(exp), 0)) {
		fmt.Println("expired state cookie")
		return false
	}

	// attach redirect url to request
	req := r.WithContext(context.WithValue(r.Context(), RedirectUrlContextKey, claims["redirectUrl"]))
	*r = *req

	stateValue := claims["value"].(string)

	if err != nil {
		fmt.Println("err", err)
		return false
	}

	return stateValue == stateFromAuthServer
}

func GeneratePKCECode() (string, *http.Cookie) {
	codeVerifier := generateRandomString()
	h := sha256.New()
	h.Write([]byte(codeVerifier))

	codeChallenge := base64.URLEncoding.EncodeToString(h.Sum(nil))

	return codeChallenge, CreateSignedCookie(PKCECookieName, codeVerifier, time.Now().Add(pkceMaxAge))
}

func ValidatePKCECode(w http.ResponseWriter, r *http.Request) string {

	defer http.SetCookie(w, DeleteCookie(PKCECookieName))

	claims, err := DecodeSignedCookie(PKCECookieName, r)

	if err != nil {
		fmt.Println("err token parse", err)
		return ""
	}

	exp, ok := claims["exp"].(float64)

	if !ok {
		fmt.Println("exp error:", exp)
		return ""
	}

	if time.Now().After(time.Unix(int64(exp), 0)) {
		fmt.Println("expired pkce cookie")
		return ""
	}

	codeVerifier := claims["value"].(string)

	return codeVerifier
}

func generateRandomString() string {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err.Error())
	}

	randomState := base64.URLEncoding.EncodeToString(randomBytes)
	return randomState
}

func getDerivedEncryptionKey(salt string) []byte {
	hash := sha256.New

	s := os.Getenv("AUTH_SECRET")

	if s == "" {
		s = "auth secret"
	}

	info := []byte("")

	secret := []byte(s)

	kdf := hkdf.New(hash, secret, []byte(salt), info)

	key := make([]byte, 32)
	_, _ = io.ReadFull(kdf, key)

	return key
}
