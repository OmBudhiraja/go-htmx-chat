package utils

import (
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/hkdf"
)

func IsValidURL(str string) bool {
	_, err := url.ParseRequestURI(str)

	if err != nil {
		return false
	}

	u, err := url.Parse(str)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	return true
}

func RenderMessageWithLinks(message string) template.HTML {
	urlRegex := regexp.MustCompile(`\bhttps?://\S+\b`)

	sanitizedMessage := template.HTMLEscapeString(message)

	// Replace URLs with <a> tags
	messageWithLinks := urlRegex.ReplaceAllStringFunc(sanitizedMessage, func(matchedURL string) string {
		return fmt.Sprintf("<a href=\"%s\" target=\"_blank\">%s</a>", matchedURL, matchedURL)
	})

	return template.HTML(messageWithLinks)
}

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}

	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func NewNullInteger(i int64) sql.NullInt64 {
	if i <= 0 {
		return sql.NullInt64{}
	}

	return sql.NullInt64{
		Int64: i,
		Valid: true,
	}
}

const (
	stateCookieName = "auth.temp-state"
)

func GenerateState() (string, *http.Cookie) {
	stateMaxAge := 15 * 60 // 15 minutes
	state := generateRandomState()
	encrptionSecret := GetDerivedEncryptionKey(stateCookieName)

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

	return state, CreateCookie("state", token, now.Add(time.Duration(stateMaxAge)*time.Second))
}

func ValidateState(stateFromAuthServer, cookieValue string) bool {
	encrptionSecret := GetDerivedEncryptionKey(stateCookieName)

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
	_, err := cryptoRand.Read(randomBytes)
	if err != nil {
		panic(err.Error())
	}

	randomState := base64.URLEncoding.EncodeToString(randomBytes)
	return randomState
}

func GetDerivedEncryptionKey(salt string) []byte {
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
