package auth

import (
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	StateCookieName   = "auth.temp-state"
	SessionCookieName = "auth.session-token"
	PKCECookieName    = "auth.pkce-token"
)

func CreateCookie(name, value string, expiry time.Time) *http.Cookie {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		Expires:  expiry,
	}
	return &cookie
}

func DeleteCookie(name string) *http.Cookie {
	cookie := http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		Expires:  time.Now().AddDate(0, 0, -1),
	}
	return &cookie
}

func CreateSignedCookie(name, value string, expiry time.Time, tokenData ...map[string]interface{}) *http.Cookie {
	encrptionSecret := getDerivedEncryptionKey(name)
	now := time.Now()

	tokenClaims := jwt.MapClaims{
		"value": value,
		"iss":   now.Unix(),
		"exp":   expiry.Unix(),
	}

	if len(tokenData) > 0 {
		for key, value := range tokenData[0] {
			tokenClaims[key] = value
		}
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims).SignedString(encrptionSecret)

	if err != nil {
		panic(err.Error())
	}

	cookie := CreateCookie(name, token, expiry)
	return cookie
}

func DecodeSignedCookie(name string, r *http.Request) (jwt.MapClaims, error) {
	cookie, err := r.Cookie(name)

	if err != nil {
		return nil, err
	}

	encrptionSecret := getDerivedEncryptionKey(name)

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return encrptionSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok || !token.Valid {
		return nil, err
	}

	return claims, nil
}
