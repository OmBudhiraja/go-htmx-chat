package utils

import (
	"net/http"
	"os"
	"time"
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
