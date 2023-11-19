package utils

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/url"
	"regexp"

	"github.com/OmBudhiraja/go-htmx-chat/internal/db"
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

func FlattenUser(u db.User) map[string]interface{} {
	flattened := make(map[string]interface{})

	flattened["Id"] = u.Id
	flattened["Name"] = u.Name
	flattened["Email"] = u.Email

	if u.Image.Valid {
		flattened["Image"] = u.Image.String
	}

	return flattened
}
