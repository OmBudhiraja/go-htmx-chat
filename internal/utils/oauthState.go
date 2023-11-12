package utils

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

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/hkdf"
)

type contextKey string

const RedirectUrlContextKey contextKey = "redirectUrl"

func GenerateState(redirectUrl string) (string, *http.Cookie) {
	stateMaxAge := 15 * 60 // 15 minutes
	state := generateRandomState()
	encrptionSecret := getDerivedEncryptionKey(StateCookieName)

	now := time.Now()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"value":       state,
		"redirectUrl": redirectUrl,
		"iss":         now.Unix(),
		"exp":         now.Add(time.Duration(stateMaxAge) * time.Second).Unix(),
	}).SignedString(encrptionSecret)

	if err != nil {
		panic(err.Error())
	}

	return state, CreateCookie(StateCookieName, token, now.Add(time.Duration(stateMaxAge)*time.Second))
}

func ValidateState(w http.ResponseWriter, r *http.Request) bool {

	stateFromAuthServer := r.URL.Query().Get("state")

	// get state coookie
	stateCookie, err := r.Cookie(StateCookieName)

	if err != nil {
		return false
	}

	defer http.SetCookie(w, DeleteCookie(StateCookieName))

	encrptionSecret := getDerivedEncryptionKey(StateCookieName)

	token, err := jwt.Parse(stateCookie.Value, func(token *jwt.Token) (interface{}, error) {
		return encrptionSecret, nil
	})

	if err != nil {
		fmt.Println("err token parse", err)
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok || !token.Valid {
		fmt.Println("claims error:", claims)
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

func generateRandomState() string {
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
