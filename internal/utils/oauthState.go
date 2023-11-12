package utils

import (
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

func GenerateState() (string, *http.Cookie) {
	stateMaxAge := 15 * 60 // 15 minutes
	state := generateRandomState()
	encrptionSecret := getDerivedEncryptionKey(StateCookieName)

	now := time.Now()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"value": state,
		"iss":   now.Unix(),
		"exp":   now.Add(time.Duration(stateMaxAge) * time.Second).Unix(),
		// "jti": "",
	}).SignedString(encrptionSecret)

	if err != nil {
		panic(err.Error())
	}

	return state, CreateCookie(StateCookieName, token, now.Add(time.Duration(stateMaxAge)*time.Second))
}

func ValidateState(stateFromAuthServer, cookieValue string) bool {

	defer DeleteCookie(StateCookieName)

	encrptionSecret := getDerivedEncryptionKey(StateCookieName)

	token, err := jwt.Parse(cookieValue, func(token *jwt.Token) (interface{}, error) {
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

	stateValue := claims["value"].(string)

	// decodedValue, err := base64.URLEncoding.DecodeString(encoded)
	// decodedState, err := base64.URLEncoding.DecodeString(stateFromAuthServer)

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
