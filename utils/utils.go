package utils

import (
	"fmt"
	"html/template"
	"net/url"
	"regexp"
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
